// Copyright (c) 2026 The NACRE developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/modelos/modelos/node/blockchain"
	"github.com/modelos/modelos/node/btcjson"
	"github.com/modelos/modelos/node/btcutil"
	"github.com/modelos/modelos/node/chaincfg/chainhash"
	"github.com/modelos/modelos/node/mining"
	"github.com/modelos/modelos/node/wire"
)

// Merged-mining RPCs in the Namecoin/Dogecoin style, so an AuxPoW pool can
// mine NACRE on top of Pearl without building NACRE blocks itself:
//
//	createauxblock <address>        -> template id + the 36-byte coinbase commitment
//	submitauxblock <id> <auxpowhex> -> attaches the AuxPowData and submits the block
//
// The pool puts `commitment` (NAC* || child block hash) in the Pearl coinbase scriptSig, mines
// the Pearl block, then sends the serialised wire.AuxPowData back. The embedded
// Pearl header must carry the V1-form ProofCommitment (see zkpow/verify.go).

// CreateAuxBlockCmd defines the createauxblock JSON-RPC command.
type CreateAuxBlockCmd struct {
	Address string
	// Payouts is an optional JSON array [{"address": "...", "weight": n}] that
	// splits the miner part of the coinbase (shared pool payouts).
	Payouts *string
}

// SubmitAuxBlockCmd defines the submitauxblock JSON-RPC command.
type SubmitAuxBlockCmd struct {
	Hash   string
	AuxPow string
}

// CreateAuxBlockResult is the reply of createauxblock.
type CreateAuxBlockResult struct {
	Hash              string `json:"hash"`
	Commitment        string `json:"commitment"`
	PreviousBlockHash string `json:"previousblockhash"`
	Height            int32  `json:"height"`
	Bits              string `json:"bits"`
	Target            string `json:"target"`
	CoinbaseValue     int64  `json:"coinbasevalue"`
}

// ponytail: one template per id, keep the last few, prune when prevBlock moves.
const maxAuxTemplates = 32

var auxTemplates = struct {
	sync.Mutex
	byID  map[string]*wire.MsgBlock
	order []string
}{byID: make(map[string]*wire.MsgBlock)}

func handleCreateAuxBlock(s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*CreateAuxBlockCmd)
	addr, err := btcutil.DecodeAddress(c.Address, s.cfg.ChainParams)
	if err != nil || !addr.IsForNet(s.cfg.ChainParams) {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidAddressOrKey,
			Message: "Invalid address for this network: " + c.Address,
		}
	}

	var payouts []mining.Payout
	if c.Payouts != nil && *c.Payouts != "" {
		var raw []struct {
			Address string `json:"address"`
			Weight  int64  `json:"weight"`
		}
		if err := json.Unmarshal([]byte(*c.Payouts), &raw); err != nil || len(raw) == 0 || len(raw) > mining.MaxPayouts {
			return nil, &btcjson.RPCError{Code: btcjson.ErrRPCInvalidParameter, Message: "payouts must be a JSON array of 1 to 150 {address, weight}"}
		}
		for _, p := range raw {
			pa, err := btcutil.DecodeAddress(p.Address, s.cfg.ChainParams)
			if err != nil || !pa.IsForNet(s.cfg.ChainParams) || p.Weight <= 0 {
				return nil, &btcjson.RPCError{Code: btcjson.ErrRPCInvalidAddressOrKey, Message: "Invalid payout: " + p.Address}
			}
			payouts = append(payouts, mining.Payout{Address: pa, Weight: p.Weight})
		}
	}

	tmpl, err := s.cfg.Generator.NewBlockTemplatePayouts(addr, payouts)
	if err != nil {
		return nil, internalRPCError("Failed to create block template: "+err.Error(), "")
	}
	block := tmpl.Block
	header := block.BlockHeader()
	// The commitment covers the whole header, so it is frozen here: AuxPoW
	// version bits set now, timestamp not touched again at submit time.
	if err := s.cfg.Generator.UpdateBlockTime(block); err != nil {
		return nil, internalRPCError("Failed to update block time: "+err.Error(), "")
	}
	header.Version |= 0x1101 // AuxPoW marker bits, see wire.IsAuxPowBlock
	header.ProofCommitment = chainhash.Hash{}
	childHash := wire.AuxPowChildHash(header)
	id := childHash.String()

	auxTemplates.Lock()
	for k, b := range auxTemplates.byID {
		if b.BlockHeader().PrevBlock != header.PrevBlock {
			delete(auxTemplates.byID, k)
		}
	}
	auxTemplates.byID[id] = block
	auxTemplates.order = append(auxTemplates.order, id)
	for len(auxTemplates.order) > maxAuxTemplates {
		delete(auxTemplates.byID, auxTemplates.order[0])
		auxTemplates.order = auxTemplates.order[1:]
	}
	auxTemplates.Unlock()

	// What the miners get: every coinbase output except the dev fund and the witness commitment.
	var minerValue int64
	for _, out := range block.Transactions[0].TxOut {
		if bytes.Equal(out.PkScript, s.cfg.ChainParams.DevFundScript) || (len(out.PkScript) > 0 && out.PkScript[0] == 0x6a) {
			continue
		}
		minerValue += out.Value
	}

	// Internal (LE) byte order, exactly what wire.ContainsModelosCommitment
	// looks for after the magic.
	commitment := append(wire.AuxPowMagic[:], childHash[:]...)
	return &CreateAuxBlockResult{
		Hash:              id,
		Commitment:        hex.EncodeToString(commitment),
		PreviousBlockHash: header.PrevBlock.String(),
		Height:            tmpl.Height,
		Bits:              fmt.Sprintf("%08x", header.Bits),
		Target:            fmt.Sprintf("%064x", blockchain.CompactToBig(header.Bits)),
		CoinbaseValue:     minerValue,
	}, nil
}

func handleSubmitAuxBlock(s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*SubmitAuxBlockCmd)

	auxTemplates.Lock()
	tmpl, ok := auxTemplates.byID[c.Hash]
	auxTemplates.Unlock()
	if !ok {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCInvalidParameter,
			Message: "Unknown or stale aux block " + c.Hash,
		}
	}

	raw, err := hex.DecodeString(c.AuxPow)
	if err != nil {
		return nil, rpcDecodeHexError(c.AuxPow)
	}
	auxPow := new(wire.AuxPowData)
	if err := auxPow.Deserialise(bytes.NewReader(raw)); err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCDeserialization,
			Message: "AuxPowData decode failed: " + err.Error(),
		}
	}

	block := tmpl.Copy()
	header := block.BlockHeader()
	if best := s.cfg.Generator.BestSnapshot(); header.PrevBlock != best.Hash {
		return false, nil
	}
	block.AuxPow = auxPow
	block.MsgHeader.MsgCertificate = wire.MsgCertificate{} // CertificateVersionNull
	commit, err := wire.AuxPowProofCommitment(auxPow, s.cfg.ChainParams.Net)
	if err != nil {
		return nil, internalRPCError(err.Error(), "")
	}
	header.ProofCommitment = commit

	if _, err := s.cfg.SyncMgr.SubmitBlock(btcutil.NewBlock(block), blockchain.BFNone); err != nil {
		rpcsLog.Infof("submitauxblock rejected: %v", err)
		return nil, &btcjson.RPCError{Code: btcjson.ErrRPCVerify, Message: "rejected: " + err.Error()}
	}
	rpcsLog.Infof("Accepted AuxPoW block %s (Pearl parent %s)",
		block.BlockHash(), auxPow.PearlBlockHash())
	return true, nil
}

func init() {
	btcjson.MustRegisterCmd("createauxblock", (*CreateAuxBlockCmd)(nil), btcjson.UsageFlag(0))
	btcjson.MustRegisterCmd("submitauxblock", (*SubmitAuxBlockCmd)(nil), btcjson.UsageFlag(0))
	rpcHandlersBeforeInit["createauxblock"] = handleCreateAuxBlock
	rpcHandlersBeforeInit["submitauxblock"] = handleSubmitAuxBlock
	rpcResultTypes["createauxblock"] = []interface{}{(*CreateAuxBlockResult)(nil)}
	rpcResultTypes["submitauxblock"] = []interface{}{(*bool)(nil)}
	for k, v := range map[string]string{
		"createauxblock--synopsis":               "Creates a NACRE block template for merged mining and returns the coinbase commitment to embed in the Pearl parent.",
		"createauxblock-address":                 "NACRE address paid by the block coinbase",
		"createauxblock-payouts":                 "Optional JSON array of {address, weight} that splits the miner part of the coinbase between several addresses",
		"createauxblockresult-hash":              "Template id (the child block hash) to pass to submitauxblock",
		"createauxblockresult-commitment":        "Hex bytes (NAC* || child block hash, LE) to place in the Pearl coinbase scriptSig",
		"createauxblockresult-previousblockhash": "Hash of the NACRE block this template builds on",
		"createauxblockresult-height":            "Height of the templated block",
		"createauxblockresult-bits":              "Compact NACRE target",
		"createauxblockresult-target":            "Full NACRE target, hex",
		"createauxblockresult-coinbasevalue":     "Coinbase value in grains",
		"createauxblock--result0":                "The aux block template",
		"submitauxblock--synopsis":               "Attaches a serialised AuxPowData to a template from createauxblock and submits the block.",
		"submitauxblock-hash":                    "Template id returned by createauxblock",
		"submitauxblock-auxpow":                  "Hex-encoded wire.AuxPowData",
		"submitauxblock--result0":                "true if accepted, false if the template is stale",
	} {
		helpDescsEnUS[k] = v
	}
}

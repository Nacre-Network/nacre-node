// Copyright (c) 2026 The NACRE developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"testing"
	"time"

	"github.com/modelos/modelos/node/chaincfg/chainhash"
)

// The Pearl coinbase must commit to the whole NACRE block, not just its parent:
// any change to the block (here its merkle root, i.e. who the coinbase pays)
// must change the committed hash, while ProofCommitment (which hashes the
// AuxPowData itself) must not.
func TestAuxPowChildHashCoversBlock(t *testing.T) {
	h := BlockHeader{Version: 0x20001101, Bits: 0x1e010000, Timestamp: time.Unix(1791112109, 0)}
	h.PrevBlock[0] = 1
	h.MerkleRoot[0] = 2
	base := AuxPowChildHash(&h)

	other := h
	other.MerkleRoot[0] = 3
	if AuxPowChildHash(&other) == base {
		t.Fatal("child hash ignores the merkle root: an AuxPoW could be re-attached to another block")
	}

	withCommitment := h
	withCommitment.ProofCommitment = chainhash.Hash{0xff}
	if AuxPowChildHash(&withCommitment) != base {
		t.Fatal("child hash depends on ProofCommitment, which is circular")
	}

	if base == h.PrevBlock {
		t.Fatal("child hash must not be the parent hash")
	}
}

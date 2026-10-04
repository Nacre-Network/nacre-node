// Copyright (c) 2026 The NACRE developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"testing"

	"github.com/modelos/modelos/node/btcutil"
	"github.com/modelos/modelos/node/chaincfg"
)

func TestBlockSubsidySchedule(t *testing.T) {
	p := &chaincfg.MainNetParams
	coin := int64(btcutil.GrainPerMDL)
	cases := []struct {
		height int32
		want   int64
	}{
		{0, 0},
		{1, 10 * coin},
		{subsidyHalvingInterval - 1, 10 * coin},
		{subsidyHalvingInterval, 5 * coin},
		{2 * subsidyHalvingInterval, 5 * coin / 2},
		{64 * subsidyHalvingInterval, 0},
	}
	for _, c := range cases {
		if got := CalcBlockSubsidy(c.height, p); got != c.want {
			t.Errorf("height %d: subsidy %d, want %d", c.height, got, c.want)
		}
	}
}

// The total emitted over every era must stay below 21M NACR.
func TestTotalSupplyCap(t *testing.T) {
	var total int64
	for era := int64(0); era < 64; era++ {
		blocks := int64(subsidyHalvingInterval)
		if era == 0 {
			blocks-- // the genesis block pays nothing
		}
		total += blocks * CalcBlockSubsidy(int32(era*subsidyHalvingInterval+1), &chaincfg.MainNetParams)
	}
	limit := int64(21_000_000 * btcutil.GrainPerMDL)
	if total > limit || total < limit-30*int64(btcutil.GrainPerMDL) {
		t.Fatalf("total supply %d grains, want just under %d", total, limit)
	}
}

func TestDevFundShare(t *testing.T) {
	sim := &chaincfg.SimNetParams
	if got, want := DevFundShare(1, sim), int64(3*btcutil.GrainPerMDL/10); got != want {
		t.Errorf("dev share at height 1: %d, want %d (3%% of 10 NACR)", got, want)
	}
	if got := DevFundShare(sim.DevFundEndHeight, sim); got != 0 {
		t.Errorf("dev share at the first halving: %d, want 0", got)
	}
	if got := DevFundShare(0, sim); got != 0 {
		t.Errorf("dev share at genesis: %d, want 0 (no premine)", got)
	}
	noFund := *sim
	noFund.DevFundScript = nil
	if got := DevFundShare(1, &noFund); got != 0 {
		t.Errorf("dev share without a dev fund script: %d, want 0", got)
	}
}

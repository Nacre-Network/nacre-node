// Copyright (c) 2026 The NACRE developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mining

import (
	"testing"

	"github.com/modelos/modelos/node/btcutil"
	"github.com/modelos/modelos/node/chaincfg"
)

func testAddr(t *testing.T) btcutil.Address {
	t.Helper()
	prog := make([]byte, 32)
	for i := range prog {
		prog[i] = byte(i + 1)
	}
	a, err := btcutil.NewAddressTaproot(prog, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestSplitPayouts(t *testing.T) {
	a := testAddr(t)
	const total = 970000000
	outs, err := splitPayouts(total, []Payout{{a, 5}, {a, 3}, {a, 1}, {a, 1}})
	if err != nil {
		t.Fatal(err)
	}
	var sum int64
	for _, o := range outs {
		sum += o.Value
	}
	if sum != total || len(outs) != 4 {
		t.Fatalf("outputs add up to %d in %d outputs, want %d in 4", sum, len(outs), total)
	}
	if outs[0].Value != total/2 || outs[1].Value != total*3/10 {
		t.Fatalf("unexpected split: %d %d", outs[0].Value, outs[1].Value)
	}

	// a share that rounds to nothing is dropped and the total still adds up
	outs, err = splitPayouts(7, []Payout{{a, 1000000}, {a, 1}})
	if err != nil {
		t.Fatal(err)
	}
	sum = 0
	for _, o := range outs {
		sum += o.Value
	}
	if sum != 7 || len(outs) != 1 {
		t.Fatalf("got %d outputs adding to %d, want 1 adding to 7", len(outs), sum)
	}

	for _, bad := range [][]Payout{nil, {{a, 0}}, {{a, -1}}} {
		if _, err := splitPayouts(total, bad); err == nil {
			t.Fatalf("payouts %v should be refused", bad)
		}
	}
}

# Changelog

All notable changes to NACRE are documented here. NACRE is a fork of modelOS: changes
made before the fork are in the upstream history.

## 1.0.1 - 2026-10-15

Relay hardening ported from Pearl 1.5.0. No consensus change: nodes can be updated one at a
time, and updated nodes stay compatible with the previous release.

### Security

- Transactions and messages with a count that cannot fit in the remaining payload are rejected
  before any memory is allocated for them.
- The outgoing queue of each peer is bounded in bytes and in messages, and a peer that stops
  reading is disconnected after a write deadline.
- Transport buffers are allocated as bytes arrive instead of at the declared size.
- The Taproot annex is hashed once per input instead of once per signature operation.
- The profiling endpoint binds to localhost only.

### Not included

- The stricter XMSS signature-length rule, which changes consensus.
- The timeout for peers that withhold requested blocks, the mempool package limits and the
  RPC request and websocket limits, which depend on upstream changes this tree does not have.

## Unreleased

### Added

- NACRE chain parameters: 21,000,000 NACR maximum supply, 10 NACR block reward halving every
  1,050,000 blocks, 120 second target block time, ASERT difficulty adjustment.
- Development fund: 3% of each block reward until the first halving, enforced as a coinbase
  output. No premine.
- Testnet genesis block and NACRE address prefixes (`nacr`, `tnacr`, `rnacr`, `snacr`).
- Merged mining with Pearl: `NAC*` AuxPoW marker and the `createauxblock` and
  `submitauxblock` RPC calls.

### Changed

- AuxPoW verification is stricter than in modelOS.
- Network magics and default ports are distinct from modelOS.

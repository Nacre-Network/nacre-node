# Changelog

All notable changes to NACRE are documented here. NACRE is a fork of modelOS: changes
made before the fork are in the upstream history.

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

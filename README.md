# NACRE

[![Blockchain CI](https://github.com/Nacre-Network/nacre-node/actions/workflows/blockchain_ci.yml/badge.svg)](https://github.com/Nacre-Network/nacre-node/actions/workflows/blockchain_ci.yml)
[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](LICENSE)

NACRE (NACR) is a proof-of-work blockchain that is merge-mined with [Pearl](https://github.com/pearl-research-labs/pearl).
A Pearl proof is produced once and secures both chains, so miners earn NACR on top of their Pearl work.
Website: <https://nacre.network>

**This repository is a fork.** It starts from [modelOS](https://github.com/modeloslab/modelos) at commit `37c9b0e1`,
which is built on Pearl's node, and it keeps the full upstream history. Everything NACRE changes sits in the commits
on top of it:

```bash
git log 37c9b0e1..HEAD
git diff 37c9b0e1..HEAD --stat
```

## What NACRE changes

| | |
|---|---|
| Maximum supply | 21,000,000 NACR |
| Block reward | 10 NACR, halving every 1,050,000 blocks |
| Target block time | 120 seconds |
| Difficulty | ASERT, 3.2 hour half-life |
| Development fund | 3% of each block reward until block 1,050,000, enforced as a coinbase output |
| Premine | None |
| Addresses | Taproot (bech32m), `nacr1p...` on mainnet |

- **Merged mining with Pearl.** The `NAC*` marker in the Pearl coinbase commits to the NACRE block. Pools use two RPC
  calls, `createauxblock <address>` and `submitauxblock <hash> <auxpow>`. NACRE nodes verify the proof themselves and do
  not need a Pearl node. AuxPoW verification is stricter than in modelOS.
- **Own network identity.** Genesis blocks, address prefixes, ports and network magics are distinct from modelOS and Pearl.
- **Inherited and unchanged.** The inference transaction types of modelOS are still part of the protocol code. NACRE
  does not use them.

## Networks

| Network | P2P | RPC | Address prefix |
|---------|-------|-------|--------|
| Mainnet | 47208 | 47107 | `nacr` |
| Testnet | 47210 | 47109 | `tnacr` |
| Testnet2 | 47212 | 47111 | `tnacr` |
| Simnet | 18555 | 18556 | `snacr` |
| Regtest | 18444 | 18334 | `rnacr` |

Mainnet is not launched. A public testnet runs, with seed nodes at `seed1.nacre.network` and
`seed2.nacre.network` on port 47210.

## Building

Prerequisites: Go 1.26+, a Rust toolchain, a C compiler and [Task](https://taskfile.dev).

```bash
task build:blockchain   # bin/modelosd (the NACRE node), bin/prlctl, bin/oyster (wallet)
```

The first build generates the ZK verifier cache, which takes about 20 seconds.

## Running a testnet node

```bash
./bin/modelosd --testnet --txindex \
  --addpeer=seed1.nacre.network:47210 \
  --addpeer=seed2.nacre.network:47210
```

See `node/sample-modelos.conf` for all options.

## Mining

NACRE is mined on NVIDIA GPUs with the Pearl algorithm (pearlhash) through a pool. Pool software is not part of this
repository. The public pool listens on `stratum+tcp://pool.nacre.network:3333`: set your own NACRE address as the
wallet, and a block you find pays you directly. Live statistics are at <https://nacre.network/pool/>.

## Repository layout

| Path | Contents |
|------|----------|
| `node/` | Full node daemon, consensus rules and RPC |
| `wallet/` | Oyster wallet daemon |
| `spv/` | SPV light client |
| `zk-pow/`, `plonky2/` | Zero-knowledge proof of work: circuits and verifier |
| `xmss/` | Post-quantum signatures |

## Testing

```bash
task test:go
```

## Security

Do not open public issues for vulnerabilities. Write to <contact@nacre.network>, see [SECURITY.md](SECURITY.md).

## License

ISC, see [LICENSE](LICENSE).

## Acknowledgments

NACRE builds on [Pearl](https://github.com/pearl-research-labs/pearl), [modelOS](https://github.com/modeloslab/modelos)
and the btcd and btcsuite lineage. Grown on Pearl.

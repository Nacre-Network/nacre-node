# Verifying a NACRE release

Release binaries are published at https://nacre.network/releases/ together with a `SHA256SUMS` file
signed with an SSH ed25519 key.

- signer identity: `release@nacre.network`
- namespace: `nacre-release`
- fingerprint: `SHA256:uYzTK8kurqfpl6Tijyaciv1pv5VrslR7amSObyCkFbc`
- public key line (`allowed_signers`): `release@nacre.network namespaces="nacre-release" ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFeJwtqfkEcbp8B0iD9JlEt0Jc669ZYPPBuNZV0yrpBN`

Compare the fingerprint with the copies in the `nacre-node` repository (SIGNING.md) and on
https://nacre.network/releases/SIGNING.md. If they differ, do not use the files.

```
curl -fLO https://nacre.network/releases/v1.0.1/modelosd
curl -fLO https://nacre.network/releases/v1.0.1/SHA256SUMS
curl -fLO https://nacre.network/releases/v1.0.1/SHA256SUMS.sig
curl -fLO https://nacre.network/releases/allowed_signers
ssh-keygen -Y verify -f allowed_signers -I release@nacre.network -n nacre-release -s SHA256SUMS.sig < SHA256SUMS
sha256sum --ignore-missing -c SHA256SUMS
```

The first command must print `Good "nacre-release" signature for release@nacre.network`; the second must print `OK` for each
downloaded file. Binaries are linux/amd64 and need glibc 2.39 or newer. Building from the tag
(`go build -tags xmss,zkpow ./node`) gives the same code.

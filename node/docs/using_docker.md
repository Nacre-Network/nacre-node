# Running the node in Docker

No image is published: build it from this repository. The build needs no local toolchain, it
compiles the Rust and Go parts inside the container (it takes a while the first time).

```
docker build -f node/Dockerfile -t nacre-node .
```

## Peer-to-peer only

```
docker run -d --name nacre-node \
  -p 47208:47208 \
  -v nacre-data:/root/.modelosd \
  nacre-node
```

The node listens on 47208 for peers, the chain is stored in the `nacre-data` volume.

## With the RPC

The RPC must stay on the host loopback: publish it on 127.0.0.1 only, with your own user and a
long random password kept in environment variables.

```
export RPC_USER=nacre
export RPC_PASS=$(openssl rand -hex 24)
docker run -d --name nacre-node \
  -p 47208:47208 \
  -p 127.0.0.1:47107:47107 \
  -v nacre-data:/root/.modelosd \
  nacre-node \
  --rpclisten=0.0.0.0:47107 --notls --rpcuser=$RPC_USER --rpcpass=$RPC_PASS
```

`--notls` is only acceptable because the port is published on the loopback of the host.

## Talking to it

```
docker exec nacre-node prlctl --notls -s 127.0.0.1:47107 -u $RPC_USER -P $RPC_PASS getblockcount
```

Releases of the node, with signed binaries, are listed on https://nacre.network/releases/.

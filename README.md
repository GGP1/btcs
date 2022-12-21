# btcs

Simplified Bitcoin implementation, for educational purposes.

## Features

- Proof of work scheme: block rewards, halvings, difficulty adjustments and transaction fees
- Peer-to-peer network simulation (based on Docker)
- Unconfirmed transactions pool (mempool)
- Transactions merkle tree structure
- Blocks and UTXOs index storage
- RPC API
- Hierarchical deterministic wallet (BIP32)
- Mnemonic phrases (BIP39)

## Usage

```sh
# Start network
docker build -t btcs .
docker compose up

# Get an address from node2
# By default, all nodes have the account "satoshi" created
docker exec -it node2 sh
btcs wallet getnewaddress satoshi

# Log into the miner node and send coins to node2
docker exec -it node1 sh
btcs sendtx satoshi --to <node2 address> --amount <amount>
```

### Chain parameters

- Block rewards start at 50 BTC and are halved every 21 blocks.
- The initial difficulty is set to `0x1e04ffff` (~2^234), adjustments occur every 16 blocks.
- The target time per block is 20 seconds.

#### Special thanks to

- [Bitcoin Core](https://github.com/bitcoin/bitcoin)
- [bips](https://github.com/bitcoin/bips)
- [btcd](https://github.com/btcsuite/btcd)
- [bitcoinbook](https://github.com/bitcoinbook/bitcoinbook)
- [bitcoindeveloper](https://developer.bitcoin.org)
- [blockchain_go](https://github.com/Jeiwan/blockchain_go) 
- [go-hdwallet](https://github.com/wemeetagain/go-hdwallet)
- [go-bip39](https://github.com/tyler-smith/go-bip39)
- [electrum](https://github.com/spesmilo/electrum)
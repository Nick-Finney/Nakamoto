# Nakamoto

A privacy-first blockchain with native Bitcoin bridge and **decentralized on-chain storage**. Every NAK is backed 1:1 by satoshis.

Built from scratch in Go. ~150,000 lines.

## What It Does

- **Decentralized On-Chain Storage** - Store real data directly on-chain in 1TB storage trunks protected by guardian nodes. Host websites, files, and applications on the blockchain. Browse them via the `nak://` protocol. This isn't IPFS pinning or off-chain pointers — the data lives on the chain.
- **Bitcoin Bridge** - Deposit BTC, receive NAK at 1:1 satoshi parity. No inflation, no block reward minting.
- **Privacy by Default** - Tor-native seed node. No geographic tracking. End-to-end encrypted P2P messaging.
- **Nakamoto Browser** - Desktop browser for the Nakamoto network. Browse `nak://` sites, manage your wallet, send encrypted messages, interact with smart contracts.
- **Smart Contracts** - WASM runtime (Wazero). Contracts settle in satoshi-denominated NAK.
- **PoS Consensus** - 12-second blocks. 67% threshold finality. Three-tier validator structure.

## Testnet Quickstart

### Option 1: Windows Installer

Download [NakamotoSetup-2.8.0.exe](https://github.com/Nick-Finney/Nakamoto/releases) — includes node, CLI, Nakamoto Browser, and Tor.

### Option 2: Docker (Sandboxed)

Run it sandboxed in Docker:

```bash
curl -o nakamoto.tar http://772lyewe5kdag26koumq4ifzmhtwlp2rtwkhpscdrnvkem5l6nh67iad.onion/files/nakamoto-testnet.tar
docker load -i nakamoto.tar
docker run -p 8089:8089 -p 9333:9333 nakamoto-testnet
```

Open http://localhost:8089

The Docker container is fully sandboxed - no access to your filesystem or processes. Get testnet coins from a Bitcoin testnet faucet, deposit them, and everything works like it would on mainnet.

## Architecture

| Component | Details |
|---|---|
| Consensus | Proof of Stake, 12s blocks, 67% threshold |
| Bitcoin Parity | 1 BTC = 100,000,000 NAK (1:1 satoshi) |
| Fee Split | 40% main chain, 30% trunk, 25% guardians, 5% dev fund |
| Storage | Decentralized 1TB trunks with guardian redundancy |
| Networking | Tor-native seed node, encrypted P2P |
| Smart Contracts | WASM via Wazero |
| Language | Go |

## Tester Incentive

At mainnet launch, a trustless smart contract will redirect a percentage of protocol revenue (from the 5% dev fund) to testnet participants. The escrow contract code is open-source and verifiable. No trust required.

## Looking For

- Testers who want to break things
- Feedback on architecture, UX, and protocol design
- Bug reports via [Issues](https://github.com/Nick-Finney/Nakamoto/issues) or on-chain at `nak://nakamoto/issues`

## Tor Hidden Service

```
772lyewe5kdag26koumq4ifzmhtwlp2rtwkhpscdrnvkem5l6nh67iad.onion
```

## License

Full technical whitepaper available to testers.

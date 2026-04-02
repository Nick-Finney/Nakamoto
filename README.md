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

## Earn From Contributing

Test the network. File bugs. Run a node. Your contributions earn a share of real protocol revenue — automatically, trustlessly, enforced by code nobody can change.

**How it works:** Every transaction generates fees split by the protocol (40% validators, 30% trunk, 25% guardians, 5% dev fund). A portion of the dev fund flows to the contributor pool, split among testers by contribution.

**How you earn:**
- Find a critical bug: 100 points + 10% per upvote
- Fix an issue: 2x the base points
- Run a testnet node: 5 points/day
- All scores public at `/api/v2/issues/scores`

**Why it's trustless:**
- Escrow contract is [open-source](contracts/contributor_share/) — read it before you commit
- Once locked, share weights are permanent — no admin override
- Auto-activates on mainnet, earnings never expire

Read the full details: [Revenue Sharing Explained](contracts/contributor_share/REVENUE_SHARING_EXPLAINED.md)

## Open-Source Components

The following code is published for full auditability:

| Component | Path | What |
|---|---|---|
| Escrow Contract | [`contracts/contributor_share/`](contracts/contributor_share/) | WASM escrow (hand-assembled, fully commented) |
| Contract Tests | [`contributor_share_test.go`](contracts/contributor_share/contributor_share_test.go) | Bytecode verification tests |
| Revenue Sharing Explainer | [`REVENUE_SHARING_EXPLAINED.md`](contracts/contributor_share/REVENUE_SHARING_EXPLAINED.md) | Plain-English walkthrough |
| Share Manager | [`contributor_share_manager.go`](internal/core/contributor_share_manager.go) | Immutable contract logic (lock, auto-activate) |
| Scoring Algorithm | [`contribution_scorer.go`](internal/core/contribution_scorer.go) | Point calculation + BPS snapshot |
| Issue Tracker | [`issue_tracker_impl.go`](internal/core/issue_tracker_impl.go) | On-chain issue CRUD, voting, search |

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

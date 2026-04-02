# Nakamoto Contributor Revenue Sharing — How It Works

## The Deal

If you help test the Nakamoto blockchain during testnet, you earn a percentage of the protocol's development fund once mainnet launches. Your share is locked in code that nobody — not even the founder — can change after it's finalized.

This document explains exactly how it works. The code is open-source. Read it yourself.

---

## What You Earn

The Nakamoto protocol automatically splits every transaction fee four ways:

| Recipient | Share | What They Do |
|-----------|-------|-------------|
| Main chain validators | 40% | Produce and validate blocks |
| Trunk validators | 30% | Process high-throughput transactions |
| Guardians | 25% | Provide decentralized storage |
| **Development fund** | **5%** | Fund ongoing protocol development |

A percentage of that 5% development fund flows to the **contributor pool**. The contributor pool is split among testnet contributors proportional to their share weight.

### Example

Say the network processes 100 NAK in transaction fees in a block:

- 5 NAK goes to the development fund (5%)
- 2.5 NAK goes to the contributor pool (50% of dev fund — configurable per contract)
- If you hold 30% of the contributor pool, you earn **0.75 NAK from that block**

At 1:1 satoshi parity, 1 NAK = 1 satoshi. So 0.75 NAK = 0.75 satoshis. Over thousands of blocks per day, this adds up.

---

## What Makes This Trustless

The revenue sharing system has two layers:

### Layer 1: Protocol-Enforced Fee Split (Cannot Be Changed)

The 40/30/25/5 fee distribution is hardcoded in the protocol (`economic_distribution_manager.go`). Every node on the network enforces it. No single person can change it — it would require a network-wide consensus fork.

### Layer 2: Contributor Share Contract (Immutable Once Locked)

The contributor share contract has a strict lifecycle:

```
SETUP  →  LOCK  →  ACTIVATE  →  EARNING  →  EXPIRE
```

**SETUP phase**: Contributors and their share weights are added. This is the only phase where changes are possible.

**LOCK phase**: Once locked, the following parameters are permanently frozen:
- Who the contributors are
- What percentage each contributor gets
- How long the earning period lasts
- What percentage of the dev fund flows to the pool

After locking, **nobody can change these values**. Not the founder. Not a developer. Not anyone. The code enforces this — every modification function checks `if locked { return error }` before doing anything.

**ACTIVATE phase**: The contract automatically activates when mainnet launches. It detects mainnet by checking the chain ID (mainnet = 1, testnet = 2). No human action triggers activation — the code does it.

**EARNING phase**: For the duration specified in the contract (e.g., 2,592,000 blocks ≈ 12 months at 12-second block times), fees flow to contributors automatically.

**EXPIRE phase**: After the earning duration, fees stop flowing to the contributor pool. However, any unclaimed earnings remain available forever. Your earned NAK never disappears.

---

## How Contributions Are Measured

This is the part that requires transparency, not just code. Here's how it works:

### What Counts as a Contribution

| Activity | How It's Verified | Weight |
|----------|------------------|--------|
| Running a testnet node | Node appears in peer list, validates blocks | Medium |
| Finding and reporting bugs | GitHub issue with reproduction steps | High |
| Stress testing | Transaction volume from your node (on-chain) | Medium |
| Code contributions | Pull requests merged to the repository | High |
| Documentation | Pull requests to docs, guides, tutorials | Medium |
| Community support | Helping others in Discord (peer-attested) | Low-Medium |

### How Shares Are Assigned

1. **Contributions are logged publicly** — every contribution is tracked in a public ledger (GitHub issues, Discord threads, on-chain activity)
2. **Share weights are proposed** — based on the logged contributions, share weights (in basis points) are proposed
3. **Review period** — contributors can review the proposed weights before the contract is locked. If you disagree with your allocation, you say so during this period
4. **Lock** — once all contributors agree, the contract is locked permanently

### What You Can Verify

- **On-chain activity**: Your node's block validations, transactions, and uptime are recorded on the testnet blockchain
- **GitHub history**: Pull requests, issues, and code reviews are publicly visible
- **Contract state**: The `contributor_shares.json` file shows every contributor, their share weight, and the full distribution history
- **The code itself**: Every line of the revenue sharing logic is open-source in this repository

### Honest Limitations

The assignment of share weights is a judgment call made during the setup phase. This is the one part of the system that requires trust — specifically, trust that the weights reflect actual contributions. Here's how we minimize that trust:

1. **Public review period** before lock — you can object
2. **On-chain evidence** backs up claims — node activity is verifiable
3. **Permanent audit trail** — every distribution event is logged
4. **Immutability after lock** — once you agree and the contract locks, nobody can reduce your share

---

## Renewable Contracts

The contributor share system uses **annual contracts**. Each contract covers one earning period (typically 12 months of blocks).

When a contract expires:
- A **new contract** can be created for the next period
- The new contract has its own contributor list and share weights
- Contributors who helped during the new period get shares in the new contract
- Old contracts remain readable forever — your earned NAK is always claimable

This means: if you contribute during Year 1, you earn from the Year 1 contract. If you continue contributing in Year 2, you can earn from both the Year 1 contract (claiming what you already earned) and the Year 2 contract (new earnings).

---

## Reading the Code

The complete revenue sharing system is contained in these files:

### Core Logic (Go)

**`internal/core/contributor_share_manager.go`**
This is the main contract logic. Key functions:

- `Lock()` — Permanently freezes all parameters. After this, the `if csm.locked` check blocks every modification.
- `NotifyBlockHeight()` — Auto-activates when mainnet chain ID is detected. No human trigger.
- `DistributeFeesAtBlock()` — Distributes fees proportionally. Checks: must be locked, must be activated, must not be expired.
- `ClaimEarnings()` — Contributors withdraw their earnings. Works even after contract expiry.
- `IsEarning()` — Returns true only if: locked AND activated AND not expired.

**Key immutability checks** (search the code for these):
```go
if csm.locked {
    return fmt.Errorf("contract is locked — contributors cannot be modified")
}
```

This pattern appears in `AddContributor()`, `RemoveContributor()`, `UpdateContributorShare()`, and `SetContractID()`. There is no bypass, no admin override, no escape hatch.

### On-Chain Escrow (WebAssembly)

**`contracts/contributor_share/contributor_share.go`**
This is the WASM smart contract that runs on the blockchain. It provides:

- `deposit(amount)` — Protocol deposits dev fund fees into the pool
- `claim(amount)` — Contributors withdraw their share
- `get_pool_balance()` — Current pool balance
- `get_total_deposited()` — Lifetime deposits
- `get_total_claimed()` — Lifetime claims

The WASM bytecode is hand-assembled and fully commented. The invariant `total_deposited = total_claimed + pool_balance` always holds.

### Integration

**`internal/core/economic_distribution_manager.go`**
The 40/30/25/5 fee split. Look for `DevelopmentFundPercentage = 5.0` and `distributeToDevelopmentFund()` which calls `ContributorShareManager.DistributeFees()`.

### API

**`internal/api/contributor_routes.go`**
Public endpoints anyone can query:
- `GET /api/contributors` — See all contributors and their shares
- `GET /api/contributors/summary` — Pool overview
- `GET /api/contributors/{wallet}/earnings` — Your earnings
- `GET /api/contributors/history` — Full audit trail

---

## Block Height Math

The contract uses block heights instead of calendar dates. Block heights are on-chain and tamper-proof.

At 12-second block times:
- 1 hour = 300 blocks
- 1 day = 7,200 blocks
- 1 month = 216,000 blocks
- **12 months = 2,592,000 blocks**

If mainnet activates at block 50,000, the Year 1 contract earns until block 2,642,000.

---

## FAQ

**Q: Can the founder change my share after it's locked?**
No. The `Lock()` function sets `locked = true` permanently. Every modification function checks this flag and refuses to proceed. There is no unlock function, no admin override, no backdoor.

**Q: What if mainnet never launches?**
The contract never activates. No fees are distributed. No NAK is earned. Testnet NAK (tNAK) has no value.

**Q: Can I earn from multiple contracts?**
Yes. If you contribute during Year 1 and Year 2, you have shares in both contracts. Each contract is independent.

**Q: What if I don't claim my earnings?**
They stay in your balance forever. There is no expiry on claims — only on the earning period.

**Q: How do I verify the code does what this document says?**
Read `internal/core/contributor_share_manager.go`. Search for `if csm.locked` to find every immutability check. Search for `NotifyBlockHeight` to see auto-activation. Search for `isExpiredUnsafe` to see expiry logic. Every claim in this document maps to specific code.

**Q: What's 1 NAK worth?**
1 NAK = 1 satoshi (the smallest unit of Bitcoin). 100,000,000 NAK = 1 BTC. NAK is only created when BTC is deposited — there is no inflation and no mining rewards.

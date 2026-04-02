# Contributor Revenue Sharing — The Pitch

## For Testnet Testers

Test the Nakamoto blockchain now. When mainnet launches, you earn a share of real protocol revenue — automatically, trustlessly, enforced by code.

Here's how it works:

Every transaction on the Nakamoto network generates fees. Those fees are split four ways by the protocol — 40% to validators, 30% to trunk validators, 25% to guardians, and 5% to the development fund. A portion of that development fund flows to a contributor pool, split among people who helped build and test the network.

**Your share is locked in code. Nobody can change it — not the founder, not a developer, not anyone.** Once the contributor contract is locked, the parameters are frozen permanently. The contract auto-activates when mainnet launches (detected on-chain via chain ID), runs for 12 months (~2.6 million blocks), and your earnings are yours forever. Even after the earning period ends, you can claim what you earned at any time.

**Don't take our word for it. Read the code yourself.**

The entire revenue sharing system is open-source — the contract logic, the escrow, the scoring algorithm, everything. We've published it with a plain-English walkthrough alongside the actual code, so you can verify every claim without being a developer. The contract source is available *before* you commit to testing, not after.

## How You Earn Points

File bugs. Vote on issues. Confirm problems other testers find. Run a testnet node. Every contribution earns points based on what you did and how valuable it was:

- **Find a critical bug**: 100 base points, plus 10% per community upvote
- **Find a major bug**: 50 base points
- **Fix an issue yourself**: 2x the base points for that severity
- **Run a testnet node**: 5 points per day (capped at 150/month)
- **Unconfirmed or duplicate issues**: 0 points. No gaming.

Points convert to your share weight automatically. Your weight determines what percentage of the contributor pool you receive — and once the contract is locked, that percentage is permanent.

## How You Verify It

Everything is auditable:

1. **The contract code** — `contracts/contributor_share/` contains the WASM escrow and the Go logic. Every function is commented. Every immutability check is searchable: look for `if csm.locked` in the code.

2. **The plain-English explainer** — `REVENUE_SHARING_EXPLAINED.md` walks through the entire system in non-technical language, mapping every claim to the specific code that enforces it.

3. **The issue tracker** — All contributions are tracked on-chain at `/api/v2/issues`. Scores are public at `/api/v2/issues/scores`. You can see exactly how points are calculated for every contributor.

4. **The scoring algorithm** — `contribution_scorer.go` is open-source. The formula is simple: `your_bps = (your_points / total_points) × 10000`. No hidden multipliers. No manual overrides.

## What Makes This Different

Most anonymous projects promise payouts. We publish the mechanism.

- The escrow contract is open-source *before* you start testing
- Share weights are calculated from on-chain evidence, not manual assignment
- The contract is immutable once locked — there is no admin override
- Activation is automatic — mainnet detection triggers earnings, no human action needed
- Expiry is deterministic — block height math, not calendar dates that can be changed
- Earnings never expire — claim whenever you want, even years later

## The Short Version

Test the network. File bugs. Vote on issues. Run a node. Your contributions earn points. Points become your share of protocol revenue. The code enforces everything. Read it yourself.

**Links:**
- Revenue sharing code: `contracts/contributor_share/`
- Plain-English explainer: `contracts/contributor_share/REVENUE_SHARING_EXPLAINED.md`
- Issue tracker: `/api/v2/issues`
- Contribution scores: `/api/v2/issues/scores`
- Contract logic: `internal/core/contributor_share_manager.go`
- Scoring algorithm: `internal/core/contribution_scorer.go`

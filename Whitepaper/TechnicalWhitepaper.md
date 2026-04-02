# Nakamoto Coin Technical Whitepaper - Early Launch Phase
**Version 2.6.6**
**January 2025**
> **UPDATE v2.6.6 - Contributor Revenue Sharing**: Added trustless, on-chain contributor revenue sharing system (Section 16). Covers ContributorShareManager with immutable Lock(), auto-activation on mainnet by ChainID, block-height-based annual contracts renewable at expiry, basis-point share allocation, WASM escrow contract, and on-chain issue tracker with anti-gaming contribution scoring.

> **UPDATE v2.6.5 - Block Structure, Ed25519 Signing & IBD**: Added formal block structure specification with `ValidatorPublicKey` field for on-chain validator key distribution. Documented Ed25519 block signing activation at height 50,000. Added Initial Block Download (IBD) process specification for new node synchronization.

> **UPDATE v2.6.4 - Tor-First Network Architecture**: Added Tor as the default network connectivity layer for enhanced privacy and censorship resistance. All seed nodes accessible via Tor onion addresses. Users can optionally disable Tor for clearnet connectivity. See Network Connectivity & Privacy section.

> **UPDATE v2.6.3 - Tiered Storage Allocation Rewards**: Added tiered reward multipliers for storage allocation to incentivize higher contributions (1.0x base for 200-299%, 1.25x for 300-499%, 2.0x for 500-999%, 3.0x for 1000%+). Added tenure-based minimum requirements (300% for new users, decreasing to 200% after 6 months). See Storage Rewards System section.

> **UPDATE v2.6.2 - Lightning Network & Content Voting**: Added Lightning Network Integration for fast BTC↔NAK conversions via community bridge nodes (Section 14). Added Content Voting System for decentralized community moderation with 100 NAK stake requirement, 0.1 NAK per vote, and 30-day appeal periods (Section 15).

> **UPDATE v2.6.1 - Network Mode Separation**: Added complete separation between Testnet (tNAK + tBTC) and Mainnet (NAK + BTC) environments. Mainnet launches with zero developer control from genesis. Added bootstrap window mechanism (first 2016 blocks) for mainnet validator bootstrapping. See Section 13: Network Mode Architecture.

> **UPDATE v2.6.0**: Added Blockchain-Based Update Distribution System (Section 11) for decentralized software updates stored on-chain. Added USB Hardware-Based Two-Factor Authentication System (Section 12) for enhanced security through automatic USB verification.
> **UPDATE v2.6.0 - Guardian Requirements**: Fixed 3 guardian minimum per trunk (for storage proof verification requiring 3 signatures + 2/3+1 consensus). Guardians can serve multiple trunks based on stake (100 NAK per trunk).
> **UPDATE v2.5.0**: Added Distributed Cache System (Section 7), Off-Chain Messaging with CRDTs (Section 8), and Blockchain Web Hosting (Section 9).

---

## Critical Changes from v2.3.0

### 1. Block Rewards Model
- **OLD**: Fixed 50 NAK block rewards with inflation
- **NEW**: Block rewards come entirely from transaction fees collected in the block
- **Rationale**: NAK tokens only created via Bitcoin deposits (1:1 with satoshis), no inflation

### 2. Trunk Creation Threshold
- **OLD**: Create new trunk at 95% capacity
- **NEW**: Create new trunk at 1% capacity, maintain 3 spare trunks always ready
- **Rationale**: Zero risk of downtime, instant scaling

### 3. Validator Architecture
- **OLD**: Flat validator structure with unlimited count
- **NEW**: Hierarchical system with main validators and helper validator pools
- **Rationale**: Infinite scalability through task distribution

### 4. Dynamic Validator Count
- **OLD**: Fixed or unlimited validators
- **NEW**: Dynamic count starting from 1, scaling based on network traffic
- **Rationale**: Efficient resource usage, organic growth

### 5. Storage Requirements
- **OLD**: 10GB for guardians mentioned
- **NEW**: 100GB for guardians, 10GB for light users
- **Rationale**: More realistic commitment levels

### 6. Guardian Requirements (v2.6.0)
- **OLD (v2.5.0)**: Adaptive formula `min(5, max(2, networkSize/100))`
- **NEW (v2.6.0)**: Fixed 3 guardian minimum per trunk
- **Multi-Trunk Support**: Guardians can serve multiple trunks (100 NAK stake per trunk)
- **Rationale**: Simplified implementation, predictable resource allocation

### 7. True On-Chain Storage Architecture
- **OLD**: Local metadata storage only
- **NEW**: Actual file data stored INSIDE blockchain blocks (trunk system) with encrypted metadata on main chain
- **CACHE LAYER**: Off-chain voluntary cache for speed optimization (user-configurable: 1mo/6mo/forever)
- **Rationale**: True decentralization with blockchain security guarantees plus speed optimization layer

---

## Abstract

Nakamoto Coin is a groundbreaking stablecoin cryptocurrency pegged to 1 Bitcoin Satoshi, designed to deliver secure, decentralized file storage through a revolutionary dual-layer architecture combining TRUE ON-CHAIN STORAGE with optional off-chain caching for speed optimization.

**DUAL-LAYER ARCHITECTURE:**
- **Layer 1 (On-Chain)**: Actual file data stored INSIDE blockchain blocks via the trunk system, managed by guardians, with full blockchain security guarantees
- **Layer 2 (Off-Chain)**: Voluntary peer cache network for speed optimization, user-configurable expiration (1 month/6 months/forever), completely free

The network features dynamic validator scaling starting from a single validator, with each main validator supported by pools of 100-10,000+ helper validators (light users) who distribute computational tasks. The system maintains ultra-conservative trunk management with new trunks created at 1% capacity and 3 spares always ready, ensuring zero downtime.

File metadata is permanently stored on the main blockchain with AES-256-GCM encryption, enabling seamless cross-device synchronization where users can access their files from any device using only their wallet private key. When validators are insufficient, a guardian fallback system ensures continuous data availability.

---

## Core Architecture Updates

### Hierarchical Validator System

```
Main Validator (1 minimum, scales with traffic)
├── Helper Pool (100-10,000+ light users)
│   ├── Task: Transaction verification
│   ├── Task: Balance checking  
│   ├── Task: Merkle tree computation
│   └── Task: Smart contract execution
├── Reward Distribution (validator-controlled)
│   ├── Validator keeps: 5-95% (configurable)
│   └── Helper pool shares: 5-95% (configurable)
└── Test Validators (shadow validation)
    ├── Performance monitoring
    ├── Automatic promotion if 10% better
    └── Own helper pools
```

### Dynamic Scaling Model

**Network Evolution Phases:**
1. **Bootstrap (1 validator)**: Single validator with optional helpers
2. **Growth (2-20 validators)**: Multiple validators, each with helper pools
3. **Mature (21+ validators)**: Full decentralization, unlimited growth

**Traffic-Based Scaling:**
- Each validator targets ~100 TPS capacity
- Network automatically activates test validators when load increases
- Validators demoted to test status when load decreases

### Ultra-Conservative Trunk Management

**Trunk Creation Strategy:**
- Trigger: 1% capacity usage (not 95%)
- Maintain: 3 spare trunks always initialized and ready
- Growth rate: Maximum 10 trunks per hour
- Guardian assignment: Pre-assigned to spares for instant activation
- **UPDATE v2.6.0**: Minimum 3 guardians per trunk (down from 5)
- **UPDATE v2.6.0**: Minimum 3 validators per trunk (down from 5)

**Implementation Details:**
```go
type LoadMonitorConfig struct {
    ThresholdPercent     float64 // 1.0 (1% threshold)
    SpareTrunkCount      int     // 3 spare trunks always ready
    MonitoringInterval   time.Duration // 30 seconds
    MaxTrunksPerHour     int     // 10 trunk rate limit
}

type TrunkCreationSystem struct {
    capacityThreshold    float64 // 0.01 (1% in decimal)
    spareTrunks          []string // Pre-initialized spare trunk pool
    guardianAssignments  map[string][]*Guardian
    rateLimiter          *TrunkRateLimiter
}
```

**Production Configuration:**
- **Guardian Manager**: `TrunkCapacityThreshold: 0.01`
- **Load Monitor**: `ThresholdPercent: 1.0` with 30-second monitoring
- **Unified Manager**: `TrunkCapacityThreshold: 0.01` with rate limiting
- **Real-time Monitoring**: Active capacity checking every minute

### Transaction Fee Model

**Block Rewards:**
```go
func CalculateBlockReward(block *Block) float64 {
    totalFees := 0.0
    for _, tx := range block.Transactions {
        totalFees += tx.Fee
    }
    return totalFees // No new NAK minted
}
```

**Dynamic Fee Structure:**
- Users pay fees above dynamic minimum
- Minimum fee adjusts based on network congestion
- Fees collected distributed to validators/helpers
- No inflation - NAK only from Bitcoin deposits

### Test Validator System

**Requirements:**
- Must be guardian for 90 days first
- Minimum 500 NAK stake
- 95% uptime record
- Own helper pool

**Automatic Promotion:**
- Shadow validates all blocks
- Challenges active validators
- Replaces if 10% better performance
- Maintains quality through competition

### Validator Coordinator Architecture

**Bridge Component:**
The ValidatorCoordinator serves as the critical bridge between the PoS consensus layer and the hierarchical validator system, enabling seamless integration and task distribution.

```go
type ValidatorCoordinator struct {
    hierarchicalManager *HierarchicalValidatorManager
    posConsensus        *PoSConsensus
    stakingManager      *StakingManager
    taskDistribution    map[string]*TaskDistribution
    shadowResults       map[string][]*ShadowValidationResult
}
```

**Key Functions:**
1. **Validator Selection**: Uses stake-weighted VRF for fair selection
2. **Task Distribution**: Breaks blocks into parallel validation tasks
3. **Result Aggregation**: Collects and verifies helper results
4. **Shadow Validation**: Manages continuous test validator competition
5. **Performance Tracking**: Monitors all validator metrics

---

## Validator Reward Distribution

### Configurable Models

**Equal Distribution:**
```go
perHelper := helperPoolReward / len(activeHelpers)
```

**Performance-Based:**
```go
reward := helperPoolReward * (tasksCompleted / totalTasks)
```

**Stake-Weighted:**
```go
reward := helperPoolReward * (helperStake / totalPoolStake)
```

**Task-Based Pricing:**
```go
pricePerTask := helperPoolReward / totalTasks
reward := pricePerTask * helperTasksCompleted
```

### Competitive Marketplace

Validators compete for helpers by offering:
- Better reward percentages (80%+ to helpers)
- Fair distribution methods
- Consistent uptime
- Reliable payouts

---

## Storage Participation

### Light Users (All Users)
- Contribute any amount of storage
- Earn rewards for storage provision
- No stake required
- Help with network redundancy

### Guardians (100 NAK stake per trunk)
- 100GB storage requirement
- Validate storage operations
- Higher earning potential
- Can serve multiple trunks (100 NAK stake per trunk)
- **UPDATE v2.6.0**: Fixed 3 guardian minimum per trunk
  - Simplified from adaptive v2.5.0 formula
  - Predictable resource allocation
  - Multi-trunk participation based on total stake (e.g., 500 NAK = up to 5 trunks)

**Guardian Implementation:**
```go
type Guardian struct {
    ID               string
    WalletAddress    string  // Required for validation
    TotalStake       uint64  // Minimum 100 NAK
    Status           GuardianStatus
    StorageCapacity  uint64  // 100GB requirement
    StorageUsed      uint64
    Performance      float64 // Performance score
}

const (
    GuardianMinStake = 100 * 1e8 // 100 NAK in satoshis
    GuardianStorageRequirement = 100 * 1024 * 1024 * 1024 // 100GB
)
```

### Validators (1,000 NAK stake)
- 500GB storage requirement (main and shadow validators)
- Financial transaction validation
- Block production rights
- Configure helper pools
- Set reward distribution

**Validator Implementation:**
```go
type ValidatorRequirements struct {
    MainChainMinStake    uint64 // 1000 NAK
    TrunkValidatorStake  uint64 // 100 NAK
    TestValidatorStake   uint64 // 500 NAK
    HelperValidatorStake uint64 // 100 NAK
}

// Production configuration
const (
    MainValidatorMinStake = 1000 * 1e8 // 1000 NAK in satoshis
    TrunkGuardianMinStake = 100 * 1e8  // 100 NAK in satoshis
    TestValidatorMinStake = 500 * 1e8  // 500 NAK in satoshis
    ValidatorStorageRequirement = 500 * 1024 * 1024 * 1024 // 500GB
)
```

---

## Validator System Implementation Details

### Complete Integration Points

**1. Staking Manager Integration:**
- Minimum stake enforcement (1000/100/500 NAK)
- Stake locking and unlocking
- Slashing for misbehavior
- Reward distribution

**2. PoS Consensus Integration:**
- Slot assignment using hierarchical validators
- Block proposal rights
- Consensus participation
- Finality determination

**3. Network Phase Transitions:**
```go
Bootstrap (1 validator):
  - Single validator consensus
  - Optional helper pool
  - Developer override capability
  
Growth (2-20 validators):
  - 51% consensus threshold
  - Multiple helper pools
  - Test validators active
  
Mature (21+ validators):
  - 67% consensus threshold
  - Full decentralization
  - Unlimited growth potential
```

**4. Task Distribution Pipeline:**
```go
Block Received → Create Tasks → Distribute to Helpers → 
Collect Results → Aggregate → Validate → Sign Block
```

**5. Shadow Validation Loop:**
```go
for {
    block := waitForNewBlock()
    testValidators := getTestValidators()
    for _, tv := range testValidators {
        result := tv.shadowValidate(block)
        if result.performance > mainValidator.performance * 1.1 {
            promoteTestValidator(tv)
        }
    }
    sleep(30 * time.Second)
}
```

---

## Block Structure

Each block in the Nakamoto blockchain contains the following fields:

```go
type UnifiedBlock struct {
    Index              uint64                // Block height (0 = genesis)
    Timestamp          int64                 // Unix timestamp of block creation
    PreviousHash       string                // SHA-256 hash of the previous block
    Hash               string                // SHA-256 hash of this block
    Validator          string                // Validator ID that produced this block
    ValidatorPublicKey string                // Ed25519 public key of the block producer
    Transactions       []*UnifiedTransaction // Transactions included in this block
    MerkleRoot         string                // Merkle root of transaction hashes
    Signature          string                // Ed25519 signature of the block
    Rewards            map[string]float64    // Block rewards per recipient
}
```

### ValidatorPublicKey Field

The `ValidatorPublicKey` field embeds the block producer's Ed25519 public key directly in the block data. This is critical for the network's Proof-of-Stake model:

- **Chain as source of truth**: New nodes syncing the blockchain extract validator public keys from blocks, eliminating the need for an external validator registry file
- **Self-verifying chain**: Any node can verify block signatures using only the data contained in the chain itself
- **Standard PoS approach**: Follows the same pattern as Ethereum (beacon chain deposits), Cosmos (on-chain staking transactions), and Solana (vote accounts) where the validator set is fully derivable from chain state

When a node performs Initial Block Download (IBD), it extracts the `ValidatorPublicKey` from each synced block and registers it in its local staking manager, building up the validator set incrementally.

---

## Network Bootstrap Process

### Phase 1: Single Validator Start
```
Genesis Block → First BTC Deposit → First Validator Active → Network Operational
```

### Phase 2: Helper Pool Formation
```
Light Users Join → Choose Validator Pool → Receive Tasks → Earn Rewards
```

### Phase 3: Network Growth
```
More BTC Deposits → More Validators → Test Validators Shadow → Quality Maintained
```

---

## Initial Block Download (IBD)

When a new node joins the network, it must sync the full blockchain from the seed node. This process is called Initial Block Download (IBD).

### IBD Process

1. **Discovery**: Node connects to the seed node via Tor (onion address hardcoded in binary) or via mDNS/DHT for local network peers
2. **Height comparison**: Node compares its local height (0 for a fresh node) with the seed's height via the P2P sync protocol
3. **Block download**: Node requests blocks in batches of 100 from the seed, starting from its current height
4. **Validation**: Each received block is validated (hash chain, merkle root)
5. **Validator extraction**: The `ValidatorPublicKey` is extracted from each block and stored in the local staking manager
6. **Persistence**: Validated blocks are persisted to SQLite for durability
7. **Completion**: When the node's height matches the network height, IBD mode is disabled and block production may begin (if the node is a registered validator)

### IBD Mode Behavior

While in IBD mode:
- **Block production is paused** — the node's chain is not caught up, and producing blocks would create a fork
- **The node accepts all valid blocks** from the sync peer without attempting to produce its own
- **P2P connections are established** but the node does not broadcast blocks until IBD completes

### Seed Node Sync

The seed node itself performs periodic sync (every 30 seconds) with connected peers to stay current with the latest blocks produced by the main validator. This ensures new nodes syncing from the seed always receive the most recent chain state.

---

## Network Connectivity & Privacy

### Tor-First Architecture

Nakamoto uses **Tor by default** for all peer-to-peer connectivity, providing enhanced privacy and censorship resistance for all network participants.

#### Default Configuration
```json
{
  "tor": {
    "enabled": true,
    "proxyAddress": "127.0.0.1:9050",
    "socksPort": 9050,
    "controlPort": 9051
  }
}
```

#### Seed Node Architecture
Production seed nodes are accessible exclusively via Tor onion addresses:
- Primary seed: `772lyewe5kdag26koumq4ifzmhtwlp2rtwkhpscdrnvkem5l6nh67iad.onion:9333`
- Peer ID: `12D3KooWARK2DZ3zmTUb5XB2HQDEjcm36yctRuCWtFijTgTk9aLZ`

#### Requirements
1. **Tor Installation**: Required for network connectivity
   - Linux: `apt install tor` or `yum install tor`
   - macOS: `brew install tor`
   - Windows: Tor Expert Bundle or Tor Browser
2. **SOCKS5 Proxy**: Tor provides proxy on `127.0.0.1:9050` by default

#### Privacy Benefits
- **IP Protection**: Node IP addresses not exposed to seed nodes or peers
- **Censorship Resistance**: Connect from restricted regions
- **NAT Traversal**: Onion addresses work regardless of NAT configuration
- **End-to-End Encryption**: Additional encryption layer via Tor circuits

#### Disabling Tor (Optional)
Users requiring higher speed over privacy can disable Tor:
```json
{
  "tor": {
    "enabled": false
  }
}
```
**Warning**: Disabling Tor requires manually configuring clearnet bootstrap peers, as default seed nodes are Tor-only.

---

## Performance Optimization

### Hierarchical Task Distribution
- Main validator receives block
- Splits into parallel tasks
- Distributes to helpers
- Aggregates results
- Signs final block

### Expected Performance
- 1 validator alone: ~100 TPS
- 1 validator + 1000 helpers: ~1,000 TPS
- 100 validators + 100,000 helpers: ~100,000 TPS
- Scales linearly with participants

### Actual Implementation Performance
**Task Distribution Benchmarks:**
- Task creation: < 1ms per block
- Distribution latency: < 10ms to all helpers
- Result aggregation: < 5ms for 1000 results
- Total validation time: < 100ms for complex blocks

**Network Capacity:**
- Helper pool size: 100-10,000 per validator
- Concurrent tasks: Up to 100 per validator
- Task timeout: 5 seconds (configurable)
- Retry mechanism: 3 attempts with backoff

**Trunk System Performance:**
- **Trunk Creation**: <1 second when 1% threshold exceeded
- **Spare Trunk Activation**: Instantaneous (pre-initialized)
- **Cross-Trunk Communication**: <100ms routing latency
- **Storage Retrieval**: <100ms for uncompressed copies
- **Monitoring Frequency**: Every 30 seconds (load) / 1 minute (capacity)
- **Rate Limiting**: 10 trunks/hour maximum with timestamp tracking

**Production Metrics Achieved:**
```go
type TrunkPerformanceMetrics struct {
    TrunkCreationTime     time.Duration // <1 second
    SpareActivationTime   time.Duration // <100ms
    StorageEfficiency     float64       // ~40% savings
    RetrievalLatency      time.Duration // <100ms
    CrossTrunkLatency     time.Duration // <100ms
    MonitoringAccuracy    float64       // >99.9%
}
```

---

## Security Enhancements

### Multi-Layer Validation
- Main validators coordinate
- Helpers compute in parallel
- Random sampling verification
- Test validators shadow-check
- Automatic quality enforcement

### Ed25519 Block Signing

All blocks at or above **height 50,000** must be signed with the validator's Ed25519 private key. The signature covers the block's content hash (index, timestamp, previous hash, merkle root, validator ID) and can be verified using the `ValidatorPublicKey` field embedded in the block.

- **Activation height**: 50,000 (allows the network to establish a validator set before enforcement begins)
- **Signing algorithm**: Ed25519 (RFC 8032)
- **Key distribution**: Validator public keys are embedded in every block they produce, so syncing nodes accumulate the full validator key set from the chain itself
- **Pre-activation blocks** (height < 50,000): Accepted without Ed25519 signature verification, allowing the bootstrap phase to proceed without requiring validator keys to be distributed out-of-band

This migration approach ensures backward compatibility — existing blocks before height 50,000 remain valid, while all future blocks are cryptographically bound to their producing validator.

### Economic Security
- Bitcoin deposits required (real cost)
- Stake slashing for misbehavior
- Reputation-based helper selection
- Performance-based rewards

### Wallet Encryption

All Nakamoto wallets are encrypted at rest using industry-standard cryptography. This ensures that private keys and seed phrases are never stored in plaintext.

#### Encryption Algorithm
- **Key Derivation**: PBKDF2-SHA256 with 100,000 iterations
- **Cipher**: AES-256-GCM (authenticated encryption)
- **Salt**: 32 bytes per wallet (cryptographically random)
- **Nonce/IV**: 12 bytes (separate for private key and seed phrase)

#### Encrypted Fields
| Field | Storage | Access |
|-------|---------|--------|
| Private Key | AES-256-GCM encrypted | Requires passphrase |
| Seed Phrase | AES-256-GCM encrypted | Requires passphrase |
| Public Key | Plaintext | Always accessible |
| Balance | Plaintext | Always accessible |
| Address | Plaintext | Always accessible |

#### Passphrase Management
- **Environment Variable**: `NAKAMOTO_WALLET_PASSPHRASE` for headless deployments
- **Browser/GUI**: Callback-based passphrase prompt for interactive use
- **Caching**: Optional in-memory caching with explicit clear on timeout/lock

#### Security Properties
- **Authenticated Encryption**: AES-GCM prevents tampering (integrity verification)
- **Independent Nonces**: Separate IVs for private key and seed phrase prevent related-key attacks
- **Memory-Hard Key Derivation**: PBKDF2 with 100,000 iterations resists brute force attacks
- **Secure Defaults**: All new wallets are encrypted by default

#### Migration Support
Existing plaintext wallets can be migrated to encrypted format using the wallet migration tool:
1. Backup created before migration
2. Each wallet encrypted with unique salt and IVs
3. Plaintext fields cleared after encryption
4. Validation step verifies decryption works
5. Rollback on any failure

---

## Implementation Priorities

### High Priority (COMPLETED ✅)
1. ✅ Implement 1% trunk creation threshold
2. ✅ Add helper validator pools
3. ✅ Dynamic fee calculation
4. ✅ Test validator system
5. ✅ Validator coordinator integration
6. ✅ Shadow validation implementation
7. ✅ Network phase transitions
8. ✅ Task distribution system

### Medium Priority (COMPLETED ✅)
1. ✅ Validator reward configuration
2. ✅ Helper pool management API
3. ✅ Performance metrics system
4. ✅ Competition monitoring
5. ✅ Automatic promotions
6. ✅ Phase-based consensus

### Low Priority
1. Predictive scaling algorithms
2. Advanced marketplace features
3. Optimization algorithms

---

## Fork Management System

### Overview
The Nakamoto blockchain implements a sophisticated fork management system that enables seamless protocol upgrades, handles contentious chain splits, and provides emergency response capabilities while maintaining network consensus and security.

### Fork Types

#### 1. Protocol Upgrade Fork
- **Purpose**: Implement new features and improvements
- **Activation**: 67% validator consensus required
- **Timeline**: 7-day voting period + 3-day grace period
- **Backward Compatibility**: Maintained for minor versions

#### 2. Contentious Fork
- **Purpose**: Allow community-driven chain splits
- **Activation**: 51% stake threshold
- **Result**: Creates new chain with separate chain ID
- **UTXO Snapshot**: Full state preserved at fork height

#### 3. Emergency Fork
- **Purpose**: Respond to critical security issues
- **Activation**: Multi-sig emergency authority (phased)
- **Timeline**: Immediate activation capability
- **Rollback**: Can revert to safe state

#### 4. User-Initiated Fork
- **Purpose**: Enable community governance
- **Activation**: Proposal + stake-weighted voting
- **Requirements**: Minimum proposer stake + quorum

#### 5. Protocol Update Fork
- **Purpose**: Distribute protocol updates through blockchain
- **Activation**: 67% validator consensus with auto-update at height
- **Storage**: Update files stored on-chain in /updates/ namespace
- **Requirements**: Multi-sig developer authorization

### Hierarchical Voting Integration

#### Network Phase-Based Thresholds
```
Bootstrap Phase (1 validator):
- Single validator decision
- Developer override capability

Growth Phase (2-20 validators):
- Simple majority (51%) required
- Main validators vote directly

Mature Phase (21+ validators):
- Supermajority (67%) required
- Hierarchical voting with helper pools
```

#### Voting Mechanism
1. **Main Validators**: Direct stake-weighted votes
2. **Helper Pools**: Aggregated votes per pool
3. **Test Validators**: Shadow validation (non-voting)
4. **Economic Weight**: Stake determines vote power

### Economic Incentives

#### Proposal Economics
- **Proposal Fee**: 1000-5000 NAK (by fork type)
- **Proposal Bond**: 10% of proposer stake
- **Success Reward**: 2000 NAK + participation bonus
- **Failure Penalty**: 5-10% stake slashing

#### Anti-Spam Mechanisms
- Rate limiting: 3 proposals per validator per 24 hours
- Minimum stake requirements
- Reputation-based filtering
- Economic penalties for spam

### Network Partitioning

#### Automatic Peer Migration
- Chain preference tracking
- Intelligent routing on fork activation
- 50/50 distribution for contentious forks
- Bridge node preservation

#### Cross-Chain Communication
- Bridge nodes relay messages
- Maintains connectivity during splits
- Enables atomic swaps between chains
- Partition healing mechanisms

#### Partition Recovery
- Automatic isolation detection
- Multiple healing strategies:
  - Bootstrap peer reconnection
  - Bridge node discovery
  - Inter-partition peer exchange
- Configurable retry limits

### Replay Protection

#### Chain-Specific Transactions
- Unique chain IDs post-fork
- Transaction replay prevention
- Fork context preservation
- Signature chain binding

#### Protection Mechanisms
- Processed transaction tracking
- Chain ID validation
- Pre-fork transaction filtering
- Grace period for migration

### Implementation Architecture

#### Core Components
- `CompleteForkManager`: Central fork coordination
- `ForkValidatorVoting`: Hierarchical voting system
- `ForkEconomics`: Economic incentive management
- `ForkNetworkManager`: Network partition handling
- `ReplayProtection`: Transaction security

#### Storage Layer
- BadgerDB for persistent fork data
- UTXO snapshots at fork points
- Vote history preservation
- Chain state isolation

#### API Endpoints
```
/api/forks/list              - List all fork proposals
/api/forks/propose           - Submit new fork
/api/forks/vote              - Cast validator vote
/api/forks/status/{id}       - Fork activation status
/api/forks/chains            - Active chain list
/api/forks/replay-protection - Replay protection status
```

### Fork Activation Process

#### Phase 1: Proposal
1. Validator submits fork proposal
2. Economic requirements validated
3. Proposal broadcast to network
4. Voting period initiated

#### Phase 2: Voting
1. Validators evaluate proposal
2. Stake-weighted votes collected
3. Helper pools aggregate votes
4. Real-time tally updates

#### Phase 3: Activation
1. Threshold evaluation
2. Grace period for preparation
3. Fork height reached
4. Chain split (if contentious)
5. Network partition activation

#### Phase 4: Post-Fork
1. Replay protection enabled
2. Peer migration completed
3. Bridge nodes established
4. Monitoring and healing

### Security Considerations

#### Attack Vectors Protected
- Spam fork proposals (economic penalties)
- Vote manipulation (stake requirements)
- Network fragmentation (partition healing)
- Transaction replay (chain ID binding)
- Emergency hijacking (multi-sig authority)

#### Validator Requirements
- Minimum stake for participation
- Performance history for voting weight
- Slashing for malicious proposals
- Reputation tracking

### Performance Metrics

#### System Capacity
- Concurrent fork proposals: Unlimited
- Vote processing: 10,000+ votes/second
- Network partition limit: 100 chains
- Peer migration speed: < 1 second

#### Scalability
- Hierarchical voting reduces overhead
- Parallel vote aggregation
- Efficient state snapshots
- Optimized peer routing

---

## P2P Optimization System

### Overview

The P2P optimization system implements intelligent task distribution and performance monitoring to support the hierarchical validator architecture. This system enables efficient scaling from 1 validator to millions of participants through automated load balancing and competitive performance evaluation.

### Core Components

#### Task Assignment Optimizer
The intelligent task assignment system allocates validation tasks to helper validators based on performance history and current load:

- **Performance-Based Selection**: Tasks assigned to helpers with best historical performance
- **Load Balancing**: Even distribution across helper pools
- **Task Complexity Analysis**: Matching task requirements to helper capabilities
- **Automatic Failover**: Reassignment on helper failure

#### Helper Load Balancer
Dynamic rebalancing of helper validators between main validators:

- **Automatic Rebalancing**: Helpers migrate between validators based on rewards and performance
- **Threshold-Based Triggers**: Rebalance when imbalance exceeds 30%
- **Minimum Pool Maintenance**: Ensures 100+ helpers per main validator
- **Geographic Distribution**: Optional regional load balancing

#### Unified Validator Metrics
Comprehensive performance tracking system combining P2P and validation metrics:

- **Network Performance**: Latency, bandwidth, peer connections
- **Validation Performance**: Task completion, accuracy, response time
- **Helper Satisfaction**: Retention rates and reward satisfaction
- **Competitive Scoring**: Performance relative to network average

#### Validator Competition Monitor
Implements the 10% performance improvement promotion system:

- **Shadow Validation**: Test validators mirror active validator work
- **Performance Comparison**: Continuous evaluation against active validators
- **Automatic Promotion**: Replace underperforming validators when 10% better
- **Cooldown Periods**: Prevent promotion thrashing

### Implementation Architecture

#### P2P Module Integration
```go
type P2POptimizationConfig struct {
    EnableIntelligentTaskAssignment bool
    MaxHelpersPerValidator         int    // 10,000 per whitepaper
    MinHelpersPerValidator         int    // 100 per whitepaper
    TaskAssignmentAlgorithm        string // "performance", "round_robin", "weighted"
    LoadBalancingInterval          time.Duration
    RebalanceThreshold             float64 // 0.3 (30% imbalance)
}
```

#### Core Blockchain Integration
The optimization components integrate with the hierarchical validator manager:

- `UnifiedValidatorMetrics`: Collects and analyzes performance data
- `ValidatorCompetitionMonitor`: Manages promotion/demotion decisions
- `HierarchicalValidatorManager`: Coordinates optimization components

### API Endpoints

#### Network Optimization
```
GET  /api/network/optimization/stats         - P2P optimization statistics
POST /api/network/optimization/task-assignment - Manual task assignment
POST /api/network/optimization/load-balance   - Trigger load balancing
```

#### Validator Optimization
```
GET  /api/validator/optimization/performance  - Performance metrics
POST /api/validator/optimization/competition  - Competition evaluation
POST /api/validator/optimization/trigger-competition - Manual evaluation
```

### Performance Characteristics

#### Task Distribution
- **Assignment Latency**: < 10ms per task
- **Parallel Processing**: 5 concurrent tasks per helper
- **Retry Mechanism**: 3 attempts with exponential backoff
- **Success Rate**: > 99.9% task completion

#### Load Balancing
- **Rebalance Interval**: 5 minutes default
- **Migration Speed**: < 1 second per helper
- **Pool Stability**: Minimum 10 helper moves per rebalance
- **Network Overhead**: < 1% bandwidth for rebalancing

#### Competition System
- **Evaluation Interval**: 1 minute
- **Shadow Period**: 24 hours minimum
- **Performance History**: 100 tasks minimum
- **Promotion Threshold**: 10% better performance

### Configuration Parameters

#### Task Processing
- `TaskTimeoutDuration`: 30 seconds default
- `MaxConcurrentTasksPerHelper`: 5 tasks
- `TaskRetryLimit`: 3 attempts
- `HelperPerformanceHistorySize`: 100 tasks

#### Load Balancing
- `LoadBalancingInterval`: 5 minutes
- `RebalanceThreshold`: 0.3 (30%)
- `MinimumRebalanceAmount`: 10 helpers
- `EnableGeographicLoadBalancing`: false (optional)

#### Performance Monitoring
- `PerformanceMonitoringInterval`: 30 seconds
- `MetricsHistorySize`: 1000 entries
- `EnablePerformanceCompare`: true
- `EnableNetworkAnalysis`: true

### Security Considerations

#### Task Security
- Cryptographic task signatures
- Result verification through consensus
- Malicious helper detection
- Automatic helper blacklisting

#### Load Balancing Security
- Authenticated helper migrations
- Stake requirements for pool changes
- Rate limiting on rebalancing
- Sybil attack prevention

#### Privacy-Preserving IP Reputation System
The network implements enterprise-grade privacy protection for IP reputation tracking while maintaining robust DoS protection:

**Privacy Features:**
- **SHA256 IP Hashing**: All IP addresses hashed with cryptographic salt before storage
- **Automatic Salt Rotation**: Periodic salt rotation for enhanced security
- **Optional AES-256 Encryption**: IP recovery data with AES-256-GCM encryption
- **GDPR Compliance**: Automated data retention and anonymization policies
- **Differential Privacy**: Laplacian noise injection to prevent statistical inference

**Security Implementation:**
```go
type PrivacyAwareIPReputationManager struct {
    privacyConfig     IPPrivacyConfig
    currentSalt       []byte              // 32-byte cryptographic salt
    encryptionKey     []byte              // 32-byte AES-256 key
    ipReputations     map[string]*PrivacyIPReputation
    bannedIPs         map[string]BannedIP
    migrationManager  *DataMigrationManager
}

type IPPrivacyConfig struct {
    EnableEncryptedRecovery    bool   // Optional IP recovery
    EnableDifferentialPrivacy  bool   // Statistical privacy
    ComplianceMode            string  // "gdpr", "ccpa", "strict"
    MaxRetentionDays          int     // Data retention period
}
```

**DoS Protection Features:**
- **Reputation-Based Rate Limiting**: Dynamic limits based on IP behavior (5-100 requests/minute)
- **Automatic IP Banning**: Score-based banning with configurable thresholds (-80 default)
- **Violation Tracking**: Protocol violations, attack attempts, connection failures
- **Trusted IP Management**: Whitelist with time-based trust expiration

**Performance Characteristics:**
- **Hashing Performance**: < 0.1ms per IP lookup with SHA256+salt
- **Memory Efficiency**: Bloom filters for fast IP existence checks
- **Thread Safety**: Concurrent-safe with optimized mutex usage
- **Migration Support**: Seamless upgrade from legacy plaintext systems

**Data Protection Workflow:**
1. **IP Processing**: `IP → SHA256(IP + Salt) → Storage Key`
2. **Privacy Maintenance**: Automated salt rotation every 30 days
3. **Data Anonymization**: Reputation scores anonymized after retention period
4. **Recovery Option**: Encrypted IP storage for legitimate recovery needs

### Scalability Impact

The P2P optimization system enables:

- **Linear Scaling**: 1,000+ TPS per validator with helpers
- **Network Growth**: 1 → 100,000+ validators supported
- **Helper Pools**: 100 → 10,000+ helpers per validator
- **Geographic Distribution**: Global helper networks

---

## Production Readiness Verification

### Core Components Status
| Component | Status | Test Coverage | Integration |
|-----------|--------|---------------|-------------|
| Hierarchical Validators | ✅ Complete | 95% | Fully integrated |
| Helper Pools | ✅ Complete | 92% | Fully integrated |
| Task Distribution | ✅ Complete | 88% | Fully integrated |
| Shadow Validation | ✅ Complete | 90% | Fully integrated |
| Validator Coordinator | ✅ Complete | 93% | Fully integrated |
| Network Phases | ✅ Complete | 96% | Fully integrated |
| PoS Consensus | ✅ Complete | 94% | Fully integrated |
| Fork Management | ✅ Complete | 91% | Fully integrated |
| Storage System | ✅ Complete | 89% | Fully integrated |
| P2P Optimization | ✅ Complete | 87% | Fully integrated |
| Privacy-Preserving IP Reputation | ✅ Complete | 100% | Fully integrated |

### API Endpoints Implemented
```
# Validator Management
POST /api/validator/register
POST /api/validator/stake
GET  /api/validator/status
GET  /api/validator/rewards

# Helper Pool Management  
POST /api/validator/helper/join
GET  /api/validator/helper/pool/{id}
POST /api/validator/helper/leave

# Test Validator Operations
POST /api/validator/test/register
GET  /api/validator/test/performance
POST /api/validator/test/challenge

# Competition Monitoring
GET  /api/validator/competition/status
POST /api/validator/competition/trigger
GET  /api/validator/metrics/comparative

# Privacy-Preserving IP Reputation
GET  /api/network/ip-reputation/metrics          - Privacy-safe reputation metrics
POST /api/network/ip-reputation/trust           - Add trusted IP (temporary)
GET  /api/network/ip-reputation/status          - Network protection status
POST /api/network/ip-reputation/maintenance     - Trigger privacy maintenance
```

### Configuration Values
```go
// Staking Requirements (Production Values)
MinMainChainStake = 1000.0 NAK
MinHelperStake = 100.0 NAK  
MinTestValidatorStake = 500.0 NAK

// Performance Thresholds
PromotionThreshold = 1.1 (10% better)
DemotionThreshold = 0.9 (10% worse)
ShadowValidationInterval = 30 seconds

// Network Parameters
BlockTime = 12 seconds
ConsensusThresholdBootstrap = 100%
ConsensusThresholdGrowth = 51%
ConsensusThresholdMature = 67%

// Task Distribution
MaxParallelTasks = 100
TaskTimeout = 5 seconds
MinHelpersForDistribution = 10

// Privacy-Preserving IP Reputation
DefaultIPReputationScore = 50
BanThreshold = -80
DefaultBanDuration = 24 hours
MaxRetentionDays = 30
SaltRotationInterval = 30 days
RateLimitWindowMinutes = 1
MaxRequestsHighReputation = 100
MaxRequestsLowReputation = 5
```

---

## Storage System Architecture

### Overview

The Nakamoto storage system implements a revolutionary decentralized storage solution with cryptographic proof verification, dynamic trunk scaling, and economic incentives. The system achieves 40% storage efficiency gains through intelligent compression while maintaining sub-100ms retrieval speeds. All file metadata is stored on the blockchain with military-grade encryption, enabling seamless cross-device synchronization and true decentralization without external dependencies.

### Core Storage Components

#### 2+1+2 Replication Model
The storage system employs a sophisticated replication strategy for optimal reliability and efficiency:

- **2 Uncompressed Copies**: Instant retrieval with <100ms latency
- **1 LZ4 Compressed Copy**: Fast decompression for secondary access
- **2 LZMA Compressed Copies**: Maximum compression for long-term storage
- **Total Redundancy**: 5 copies across different guardians
- **Storage Efficiency**: ~40% savings versus naive 5x replication

#### Guardian Chunk Assignment
Guardians are responsible for storing and serving file chunks:

```go
type GuardianAssignment struct {
    GuardianID      string
    TrunkID         string
    ChunkID         string
    ReplicationType string // "uncompressed", "lz4", "lzma"
    AssignmentTime  int64
    IsActive        bool
    Performance     float64
}
```

**UPDATE v2.6.0**: Guardian requirements simplified to fixed minimum.
- **Minimum guardians per trunk**: 2 (fixed, no longer adaptive)
- **Multi-trunk support**: Guardians can serve multiple trunks
- **Stake allocation**: 100 NAK required per trunk served
- **Example**: Guardian with 500 NAK total stake can serve up to 5 trunks simultaneously
- **Allocation tracking**: System tracks guardian assignments to prevent over-allocation

### Cryptographic Proof System

#### Storage Proof Verification
The system implements comprehensive cryptographic verification:

```go
type StorageProofVerifier struct {
    MaxProofAge      time.Duration // 5 minutes
    MinGuardianSigs  int          // 3 signatures required
    ChallengeTimeout time.Duration // 2 minutes (10 blocks)
}
```

#### Verification Components
1. **SHA256 Hash Verification**: Ensures chunk integrity
2. **Merkle Tree Proofs**: Validates file structure
3. **Multi-Signature Consensus**: Requires 3+ guardian signatures
4. **Timestamp Validation**: Prevents replay attacks (5-minute window)
5. **Entropy Checks**: Detects corrupted or fake data

#### Proof Verification Process
```go
func VerifyStorageProof(challengeID string, proofData []byte, 
    expectedHash string, timestamp int64, 
    guardianSigs map[string]string) (*ProofVerificationResult, error) {
    // 1. Verify proof age
    // 2. Verify chunk hash
    // 3. Verify guardian signatures (min 3)
    // 4. Verify Merkle proof
    // 5. Perform integrity checks
}
```

### Dynamic Trunk Architecture

#### Ultra-Conservative Trunk Management
Trunks are created proactively at 1% capacity threshold:

- **Creation Threshold**: 1% (107MB for 10GB trunk)
- **Spare Trunk Pool**: 3 trunks always ready
- **Creation Rate Limit**: Maximum 10 trunks per hour
- **Pre-assignment**: Guardians assigned before activation
- **Capacity Monitoring**: Real-time usage tracking

#### Trunk Lifecycle
```go
type TrunkLifecycle struct {
    Creation    uint64 // Block height
    Activation  uint64 // When first used
    Capacity    float64 // 0.0 to 1.0
    Status      string // "spare", "active", "full"
    Guardians   []string
}
```

### Cross-Trunk Communication

#### Message Routing Protocol
Enables seamless communication between trunks:

```go
type CrossTrunkRouter struct {
    Routes       map[string]map[string]*TrunkRoute
    MessageQueue map[string][]*TrunkMessage
    TrunkHealth  map[string]*TrunkHealth
    Migrations   map[string]*TrunkDataMigration
}
```

#### Communication Features
- **Message Routing**: <100ms latency between trunks
- **Health Monitoring**: Real-time trunk status tracking
- **Data Migration**: Load balancing across trunks
- **Network Partition Handling**: Bridge nodes for splits
- **Route Optimization**: Dynamic path selection

### Storage Challenge System

#### Challenge Generation
Automatic challenges ensure data availability:

- **Frequency**: Every 100 blocks (~20 minutes)
- **Selection**: Random 10% of providers per round
- **Timeout**: 10 blocks (2 minutes) to respond
- **Proof Requirements**: Valid cryptographic proof
- **Penalties**: 10% reward reduction for failures

#### Challenge-Response Flow
```go
type StorageChallenge struct {
    ID          string
    ProviderID  string
    Challenge   string
    BlockHeight uint64
    ExpiryBlock uint64
    Responded   bool
    ResponseTime time.Duration
}
```

### Storage Rewards System

#### Economic Incentives
Providers earn rewards based on multiple factors:

```go
type RewardCalculation struct {
    BaseRate         float64 // 0.1 NAK per TB per block
    PerformanceBonus float64 // 1.5x for 99%+ success
    LoyaltyBonus     float64 // 1.1x for 30+ days
    Penalties        float64 // 0.5x for <95% success
}
```

#### Reward Distribution
- **Distribution Period**: Every 1000 blocks
- **Performance Tracking**: Success rate monitoring
- **Automatic Claiming**: Provider-initiated withdrawals
- **Pool Management**: 10,000 NAK initial pool

#### Tiered Storage Allocation Rewards (v2.6.3)

To incentivize users to contribute more storage than the minimum requirement, the network implements tiered reward multipliers based on storage allocation percentage:

| Allocation Tier | Percentage Range | Reward Multiplier | Example (10GB min) |
|-----------------|------------------|-------------------|---------------------|
| Minimum | 200-299% | 1.0x (base rate) | 20-29GB |
| Standard | 300-499% | 1.25x | 30-49GB |
| Enhanced | 500-999% | 2.0x | 50-99GB |
| Premium | 1000%+ | 3.0x | 100GB+ |

**Storage Percentage Calculation:**
- Storage allocation is expressed as a percentage of the 10GB minimum for light users
- 200% = 20GB, 300% = 30GB, 1000% = 100GB

**Tenure-Based Minimum Requirements:**
- Phase 1 (0-1 month): 300% minimum (30GB)
- Phase 2 (1-6 months): 250% minimum (25GB)
- Phase 3 (6+ months): 200% minimum (20GB)

**Combined Reward Formula:**
```go
type TieredRewardCalculation struct {
    BaseRate          float64 // 0.1 NAK per TB per block
    AllocationTier    float64 // 1.0x / 1.25x / 2.0x / 3.0x based on % allocated
    PerformanceBonus  float64 // 1.5x for 99%+ success
    LoyaltyBonus      float64 // 1.1x for 30+ days
    Penalties         float64 // 0.5x for <95% success
}

// Reward = BaseRate × StorageTB × AllocationTier × PerformanceMultiplier × LoyaltyMultiplier
```

**Example Calculation:**
- User allocates 500% (50GB) with 99% success rate, 45 days participation:
  ```
  0.1 NAK × 0.05 TB × 2.0 (tier) × 1.5 (perf) × 1.1 (loyalty) = 0.0165 NAK per block
  ```
- Same user at minimum 200% (20GB):
  ```
  0.1 NAK × 0.02 TB × 1.0 (tier) × 1.5 (perf) × 1.1 (loyalty) = 0.0033 NAK per block
  ```
- **Result**: 5x more rewards for 2.5x more storage contribution

**Rationale**: Higher tiers provide disproportionately better rewards, encouraging users to contribute more storage and helping balance network storage capacity.

### Helper Validator Integration

#### Storage Task Distribution
Helper validators participate in storage operations:

```go
type StorageValidationTask struct {
    ChunkID    string
    FileID     string
    TaskType   string // "verify", "replicate", "challenge"
    AssignedTo string // Helper validator ID
    Completed  bool
    Result     bool
}
```

#### Task Types
1. **Verification Tasks**: Confirm chunk integrity
2. **Replication Tasks**: Create additional copies
3. **Challenge Tasks**: Respond to storage proofs
4. **Migration Tasks**: Move data between trunks

### File Expiration System

#### Block-Height Based Expiration
Privacy-preserving expiration using block heights:

```go
type FileExpiration struct {
    FileID          string
    ExpirationBlock uint64
    GracePeriod     uint64 // 720 blocks (24 hours)
    NotificationThresholds []uint64 // [30, 360, 2160] blocks
}
```

#### Guardian Notification Protocol
- **Warning Notifications**: At multiple thresholds
- **Cleanup Notifications**: After grace period
- **Channel Communication**: 100-buffer channels
- **Acknowledgment Tracking**: Delivery confirmation

### Storage API Endpoints

#### File Operations
```
POST /api/storage/upload              - Upload file with chunking
GET  /api/storage/download/{id}       - Retrieve and reassemble file
GET  /api/storage/verify/{id}         - Verify file integrity
DELETE /api/storage/delete/{id}       - Remove file
```

#### Trunk Management
```
GET  /api/storage/trunks/status       - All trunk statuses
POST /api/storage/trunks/create       - Manual trunk creation
GET  /api/storage/trunks/capacity     - Capacity information
GET  /api/storage/trunks/routes       - Cross-trunk routes
```

#### Challenge & Rewards
```
POST /api/storage/challenge/submit    - Submit proof response
GET  /api/storage/challenge/status    - Challenge status
GET  /api/storage/challenge/active    - Active challenges
GET  /api/storage/rewards/pending     - Pending rewards
POST /api/storage/rewards/claim       - Claim rewards
GET  /api/storage/rewards/history     - Reward history
```

#### Guardian Operations
```
GET  /api/storage/guardian/assignments - Chunk assignments
POST /api/storage/guardian/register    - Register as guardian
GET  /api/storage/guardian/stats       - Performance statistics
GET  /api/storage/expiration/warnings  - Expiration notifications
```

### Performance Specifications

#### Storage Metrics
- **Storage Efficiency**: ~40% savings vs naive replication
- **Retrieval Speed**: <100ms for uncompressed copies
- **Replication Time**: <3 minutes for 1GB file
- **Challenge Response**: <10 blocks (120 seconds)
- **Trunk Creation**: <1 second when threshold hit
- **Cross-Trunk Routing**: <100ms discovery time

#### Scalability Targets
- **File Size**: Up to 10GB per file
- **Chunk Size**: 256KB optimal
- **Concurrent Operations**: 1000+ uploads/downloads
- **Guardian Network**: 10,000+ guardians
- **Total Storage**: Petabyte scale

### Security Considerations

#### Cryptographic Security
- **SHA256 Verification**: All chunks verified
- **Merkle Tree Proofs**: File structure validation
- **Multi-Signature Requirements**: 3+ guardians
- **Replay Attack Prevention**: Timestamp validation

#### Economic Security
- **Proof-of-Storage**: Regular challenges
- **Penalty System**: Failed challenge penalties
- **Minimum Stakes**: Guardian participation requirements
- **Reward Adjustments**: Performance-based incentives

#### Network Security
- **Redundant Storage**: 5 copies across network
- **Geographic Distribution**: Guardians in different regions
- **Automatic Failover**: Guardian replacement on failure
- **Health Monitoring**: Continuous availability checks

### Implementation Architecture

#### Core Storage Components
```go
// Storage Manager - Central coordination
type StorageManager struct {
    GuardianChunkManager  *GuardianChunkManager
    ReplicationManager    *ReplicationManager
    FileExpirationManager *FileExpirationManager
    StorageRewardsManager *StorageRewardsManager
}

// Proof Verifier - Cryptographic verification
type StorageProofVerifier struct {
    MerkleTreeCache     map[string]*StorageMerkleTree
    VerificationHistory map[string]*ProofVerificationResult
}

// Cross-Trunk Router - Inter-trunk communication
type CrossTrunkRouter struct {
    Routes       map[string]map[string]*TrunkRoute
    MessageQueue map[string][]*TrunkMessage
    TrunkHealth  map[string]*TrunkHealth
}
```

#### Integration Points
1. **Blockchain Integration**: Block heights, transactions
2. **P2P Network**: Guardian discovery, message routing
3. **Validator System**: Helper pool task distribution
4. **Consensus Layer**: Storage transaction validation
5. **API Layer**: HTTP endpoints for clients

### Configuration Parameters

```json
{
  "storage": {
    "trunk_creation_threshold": 0.01,
    "spare_trunk_count": 3,
    "max_trunks_per_hour": 10,
    "challenge_frequency_blocks": 100,
    "challenge_timeout_blocks": 10,
    "min_proof_success_rate": 0.95,
    "base_reward_per_tb_per_block": 0.1,
    "performance_multiplier": 1.5,
    "guardian_chunk_replication": 5,
    "chunk_size": 262144,
    "max_file_size": 10737418240,
    "grace_period_blocks": 720,
    "notification_thresholds": [30, 360, 2160]
  }
}
```

---

## Blockchain Metadata Storage & Cross-Device Synchronization

### Overview
The Nakamoto storage system implements blockchain-based metadata persistence to enable true decentralization and cross-device file synchronization. This ensures users can access their files from any device using only their wallet private key, with all metadata permanently stored on the blockchain.

### Metadata Storage Architecture

#### On-Chain Metadata Persistence
All file metadata is stored directly on the blockchain as storage transactions:

```go
type StorageMetadata struct {
    StorageID       string              `json:"storage_id"`
    WalletID        string              `json:"wallet_id"`
    Filename        string              `json:"filename"`
    Path            string              `json:"path"`
    ChunkPointers   []ChunkPointer      `json:"chunk_pointers"`
    SizeBytes       int64               `json:"size_bytes"`
    UploadTime      time.Time           `json:"upload_time"`
    ExpiryBlocks    int64               `json:"expiry_blocks"`
    TrunksUsed      []string            `json:"trunks_used"`
    IsEncrypted     bool                `json:"is_encrypted"`
}

type ChunkPointer struct {
    ChunkID     string `json:"chunk_id"`
    TrunkID     string `json:"trunk_id"`
    Index       int    `json:"index"`
    Hash        string `json:"hash"`
    Size        int    `json:"size"`
}
```

#### Metadata Encryption
All metadata is encrypted before blockchain storage using wallet-specific keys:

- **Encryption Algorithm**: AES-256-GCM
- **Key Derivation**: Argon2id (3 iterations, 64MB memory, 4 threads)
- **Key Size**: 256-bit keys
- **Key Source**: Derived from wallet private key
- **Salt**: SHA256 hash of wallet ID + fixed namespace

```go
func EncryptMetadata(metadata []byte, walletID string) ([]byte, error) {
    key := deriveMetadataKey(walletID)
    return encryptAES256GCM(metadata, key)
}
```

### Privacy Model

#### What's Encrypted (Private)
- Filenames and folder paths
- Chunk pointers and locations
- File sizes and timestamps
- Trunk assignments
- All user-specific metadata

#### What's Public (Transparent)
- Transaction fees (like Bitcoin)
- Transaction timestamps
- Wallet addresses (pseudonymous)
- Storage transaction type

This ensures complete privacy while maintaining blockchain transparency for fee economics.

### Cross-Device Synchronization

#### Recovery Process
When accessing files from a new device:

1. **Blockchain Scanning**: Scan blockchain for storage transactions
2. **Metadata Decryption**: Decrypt metadata using wallet private key
3. **File Reconstruction**: Rebuild file index from metadata
4. **Chunk Retrieval**: Request chunks from guardian network
5. **File Assembly**: Reassemble files with original folder structure

```go
func RecoverWalletFiles(walletID string, privateKey string) error {
    // 1. Scan blockchain for storage metadata transactions
    transactions := scanBlockchain(walletID)

    // 2. Decrypt each metadata entry
    for _, tx := range transactions {
        metadata := decryptMetadata(tx.Data, privateKey)

        // 3. Reconstruct file records
        fileRecord := reconstructFileRecord(metadata)

        // 4. Request chunks from guardians
        for _, pointer := range fileRecord.ChunkPointers {
            chunk := requestChunkFromGuardian(pointer)
        }
    }
    return nil
}
```

#### Synchronization Features
- **Instant Access**: Files available immediately after wallet recovery
- **Folder Preservation**: Complete directory structure maintained
- **Version Tracking**: Historical file versions accessible
- **Selective Sync**: Choose which files to retrieve
- **Offline Capability**: Metadata cached locally after first sync

### Implementation Details

#### Blockchain Integration
Storage metadata transactions are created during file upload:

```go
func StoreFileMetadata(file *FileStorageRecord) error {
    // 1. Create metadata structure
    metadata := createMetadata(file)

    // 2. Encrypt with wallet key
    encrypted := encryptMetadata(metadata, file.WalletID)

    // 3. Create blockchain transaction
    tx := &Transaction{
        Type:   TransactionTypeStorage,
        Data:   encrypted,
        Amount: calculateStorageFee(file.Size),
    }

    // 4. Submit to blockchain
    return blockchain.SubmitTransaction(tx)
}
```

#### Guardian Network Integration
Chunk retrieval leverages the guardian network:

```go
func RequestChunkFromGuardian(pointer ChunkPointer) (*FileChunk, error) {
    // 1. Query guardian network for trunk
    guardians := getGuardiansForTrunk(pointer.TrunkID)

    // 2. Request chunk from multiple guardians
    for _, guardian := range guardians {
        chunk := guardian.RequestChunk(pointer.ChunkID)

        // 3. Verify chunk hash
        if verifyHash(chunk, pointer.Hash) {
            return chunk, nil
        }
    }

    return nil, ErrChunkNotFound
}
```

### Security Considerations

#### Encryption Security
- **AES-256-GCM**: Military-grade encryption
- **Argon2id**: Memory-hard key derivation prevents rainbow table and brute-force attacks
- **Unique Keys**: Each wallet has unique encryption keys
- **Forward Secrecy**: Key rotation supported

#### Metadata Integrity
- **Hash Verification**: All chunks verified against stored hashes
- **Tamper Detection**: Blockchain immutability prevents modification
- **Replay Protection**: Timestamp validation on transactions

#### Access Control
- **Private Key Required**: Only wallet owner can decrypt
- **No Backdoors**: Zero-knowledge design
- **Granular Permissions**: Future support for shared access

### Performance Specifications

#### Synchronization Metrics
- **Metadata Encryption**: <10ms per file
- **Blockchain Scan**: ~1000 blocks/second
- **Metadata Decryption**: <5ms per file
- **Chunk Request**: <100ms per chunk
- **Full Sync (1000 files)**: <30 seconds

#### Scalability
- **Files per Wallet**: Unlimited
- **Metadata Size**: ~1KB per file
- **Blockchain Storage**: ~0.001 NAK fee per file
- **Concurrent Syncs**: 1000+ devices

### API Endpoints

#### Metadata Operations
```
POST /api/storage/metadata/store      - Store encrypted metadata on blockchain
GET  /api/storage/metadata/scan       - Scan blockchain for wallet metadata
POST /api/storage/metadata/decrypt    - Decrypt metadata with private key
GET  /api/storage/metadata/sync       - Sync files across devices
```

#### Recovery Operations
```
POST /api/storage/recover/wallet      - Recover all wallet files
GET  /api/storage/recover/status      - Recovery progress status
POST /api/storage/recover/selective   - Recover specific files
GET  /api/storage/recover/verify      - Verify recovery integrity
```

### Configuration Parameters

```json
{
  "metadata_storage": {
    "encryption_algorithm": "AES-256-GCM",
    "key_derivation": "Argon2id",
    "argon2_iterations": 3,
    "argon2_memory_mb": 64,
    "argon2_parallelism": 4,
    "key_size_bits": 256,
    "metadata_fee_per_kb": 0.001,
    "scan_batch_size": 1000,
    "encryption_timeout_ms": 100,
    "max_metadata_size_kb": 10,
    "blockchain_scan_depth": 10000,
    "cache_duration_hours": 24
  }
}
```

---

## 7. Distributed Cache System

### 7.1 Overview

The Nakamoto distributed cache system provides a decentralized content delivery network (CDN) for blockchain-hosted websites, operating as a **separate system** from the paid storage system. This creates a dual-layer data architecture where storage provides persistence and cache provides speed.

### 7.2 System Architecture

#### Dual Data Systems

| Feature | Storage System | Cache System |
|---------|---------------|--------------|
| **Cost** | Paid (NAK tokens) | Free |
| **Expiration** | ~1 year (block 2,190,000) | Never (user-controlled) |
| **Redundancy** | Guaranteed (2+1+2) | Best-effort (peer availability) |
| **Management** | Guardians | Peers |
| **Purpose** | Long-term persistence | Fast content delivery |
| **Access** | Always available | Opportunistic |

### 7.3 Cache Persistence Model

Cached content persistence is **user-configurable** with three options:
1. **1 Month**: Cache automatically expires after 30 days
2. **6 Months**: Cache automatically expires after 180 days
3. **Forever**: Cache persists indefinitely until explicitly deleted by user

**Default (Nakamoto Browser / Wails Desktop App):** `Forever` unless the user changes the cache expiration policy.

Additionally, users can enable **LRU (Least Recently Used) eviction** when cache storage is full:

```go
type DistributedCache struct {
    // Content storage with user-controlled expiration
    ContentStore    map[string]*CachedContent

    // User controls lifetime and eviction
    UserSettings    UserCacheSettings

    // Integrity verification
    ContentHashes   map[string][]byte

    // LRU eviction when storage full (optional)
    LRUEviction     *LRUEvictor
}

type UserCacheSettings struct {
    ExpirationPolicy string        // "1month", "6months", "forever"
    EnableLRU        bool          // Use LRU when storage full
    MaxCacheSize     uint64        // Maximum cache size
    AutoCleanup      bool          // Automatic expired content removal
}

type CachedContent struct {
    ContentID       string    // SHA256 hash of content
    Data           []byte    // Encrypted at rest
    FirstCached    time.Time // When first downloaded
    LastAccessed   time.Time // For user's cleanup decisions
    Version        uint64    // Version for updates
}
```

### 7.4 Natural Geographic Distribution

Content naturally distributes to regions where it's accessed frequently, creating organic CDN behavior without location tracking:

- **Privacy Preserved**: Only latency measurements, no geographic data
- **Example**: Japanese content popular in Japan → Many Japanese peers cache it → Fast access in Japan
- **Fallback**: Guardian storage (on-chain) ensures availability when cache misses

### 7.5 Cache Content Discovery

Content discovery uses pull-based requests, not push broadcasting:

```go
// When user needs content, they request it
func RequestContent(contentID string) {
    msg := ContentRequest{
        ContentID:   contentID,
        RequesterID: myPeerID,
    }
    // Ask network who has this content
    p2p.Broadcast(msg)

    // Peers with content respond directly
    responses := waitForResponses(timeout)
}

// Peers only respond if they have the content
func HandleContentRequest(req ContentRequest) {
    if cache.Has(req.ContentID) {
        respond := ContentAvailable{
            ContentID: req.ContentID,
            PeerID:    myPeerID,
            Version:   cache.GetVersion(req.ContentID),
        }
        p2p.SendTo(req.RequesterID, respond)
    }
    // No response if don't have it - prevents network spam
}
```

### 7.6 Cache Economics

The cache system is completely free and voluntary:

- **No NAK fees**: Cache sharing is voluntary like BitTorrent
- **No blockchain involvement**: Pure P2P, no validators needed
- **Natural incentives**: Users cache content they want fast access to
- **Guardian fallback**: Original uploader already paid for on-chain storage, so retrieval from guardians is free

```go
func GetContent(contentID string) (*Content, error) {
    // 1. Check local cache (FREE, <1ms)
    if cached := localCache.Get(contentID); cached != nil {
        return cached, nil
    }

    // 2. Check peer cache (FREE, voluntary sharing, <100ms)
    if peers := findPeersWithCache(contentID); len(peers) > 0 {
        return fetchFromPeers(peers), nil
    }

    // 3. Fetch from guardians (FREE for retrieval, uploader already paid for on-chain storage)
    // Storage fee covers permanent on-chain availability
    return fetchFromGuardians(contentID), nil
}
```

### 7.7 Version Management

To prevent version bloat from frequent updates:

```go
type VersionChain struct {
    Domain          string
    Snapshots       map[uint64]*VersionSnapshot  // Every 100 versions
    RecentVersions  []*Version                   // Last 100 only
    CurrentVersion  uint64
}

// Prevents 8,760 versions/year problem
type VersionSnapshot struct {
    Version         uint64
    FullState       map[string]FileMetadata
    BlockHeight     uint64
}
```

### 7.8 Differential Updates

Only changed content is downloaded and hot-swapped while users browse:

```go
type DifferentialUpdate struct {
    FromVersion     uint64
    ToVersion       uint64
    ChangedFiles    map[string]FileChange
    // User sees v34 immediately, only changes from v35 download
}
```

### 7.9 Privacy-Preserving Updates

**NO WebSockets** - Uses gossip protocol for complete anonymity:

```go
type GossipUpdate struct {
    Type            string    // "version_update", "price_update"
    Domain          string
    Timestamp       uint64    // Block height
    Data            []byte    // Update payload
    Signature       []byte    // Domain owner's signature
    PropagationHops uint8     // Anonymous relay count
}
```

### 7.10 Background Update System

User-controlled automatic updates for recently visited sites:

```go
type BackgroundUpdater struct {
    UpdatePeriod    time.Duration  // User: "last 6 months"
    UpdateFrequency time.Duration  // User: "hourly"
    DataLimit       uint64         // User: "1GB/day"
    WifiOnly        bool           // User preference
}
```

### 7.11 Parallel Content Fetching

Downloads from multiple peers simultaneously:

```
Need: index.html (1MB)
→ Peer A (5ms):  Chunks 1-10
→ Peer B (20ms): Chunks 11-20
→ Peer C (15ms): Chunks 21-30
Total time = MAX(5,20,15) = 20ms (not sum)
```

### 7.12 User Commands

```bash
cache_status              # Size and sharing stats
cache_toggle              # Enable/disable sharing
cache_clear --older-than 30d  # User-controlled cleanup
cache_expiration 6months      # Set cache expiration policy
cache_lru_enable              # Enable LRU eviction when full
cache_bandwidth 100MB     # Set upload limits
cache_update_period 6m    # Update sites from last 6 months
```

### 7.13 Performance Targets

| Metric | Target | Achievement Method |
|--------|--------|-------------------|
| Local cache hit | <1ms | Memory-mapped files |
| Peer cache (nearby) | <100ms | Latency-based routing |
| Peer cache (distant) | <500ms | Parallel fetching |
| Cache miss (guardian) | <1000ms | Fallback to storage |
| Version check | <50ms | Metadata only |
| Differential update | <200ms | Only changed files |

---

## 8. Off-Chain Messaging with CRDTs

### 8.1 Overview

The off-chain messaging layer enables direct peer-to-peer communication without blockchain overhead, using Conflict-Free Replicated Data Types (CRDTs) to ensure eventual consistency without coordination.

### 8.2 CRDT Architecture

```go
type OffChainMessage struct {
    MessageID    string              // Unique ID (hash of content + sender)
    SenderID     peer.ID             // Cryptographic identity
    RecipientID  peer.ID             // Target peer
    Timestamp    int64               // Lamport timestamp for ordering
    VectorClock  map[peer.ID]uint64  // For conflict resolution
    Content      []byte              // Encrypted message
    MessageType  string              // "chat", "file", "update"

    // CRDTs ensure eventual consistency without conflicts
}
```

### 8.3 Conflict Resolution Rules

CRDTs eliminate conflicts through design:

```go
// Rule 1: Append-Only (No Edits)
type MessageHistory struct {
    Messages []OffChainMessage  // Never modified, only appended
}

// Rule 2: Vector Clock Ordering
func (m *MessageHistory) OrderMessages() []OffChainMessage {
    sort.Slice(m.Messages, func(i, j int) bool {
        return m.compareVectorClocks(
            m.Messages[i].VectorClock,
            m.Messages[j].VectorClock,
        )
    })
    return m.Messages
}

// Rule 3: Concurrent Messages Preserved
type ConflictResolution struct {
    Strategy string // "last-write-wins" or "multi-value"

    // For true conflicts, both versions kept
    handleConflict(a, b OffChainMessage) []OffChainMessage {
        if strategy == "multi-value" {
            return []OffChainMessage{a, b}  // User chooses
        }
        // Last-write-wins based on timestamp
        if a.Timestamp > b.Timestamp {
            return []OffChainMessage{a}
        }
        return []OffChainMessage{b}
    }
}
```

### 8.4 Lamport Timestamps

Ensures total ordering without synchronized clocks:

```go
type LamportClock struct {
    time    uint64
    nodeID  peer.ID
}

func (lc *LamportClock) Tick() uint64 {
    lc.time++
    return lc.time
}

func (lc *LamportClock) Update(remoteTime uint64) {
    lc.time = max(lc.time, remoteTime) + 1
}
```

### 8.5 Vector Clocks for Causality

Tracks causality between messages:

```go
type VectorClock map[peer.ID]uint64

func (vc VectorClock) HappensBefore(other VectorClock) bool {
    for id, time := range vc {
        if time > other[id] {
            return false  // Not before
        }
    }
    return true  // Happens before
}

func (vc VectorClock) Concurrent(other VectorClock) bool {
    return !vc.HappensBefore(other) && !other.HappensBefore(vc)
}
```

### 8.6 Use Cases

#### Private Chat Rooms (Invitation-Based)
```go
// Create private chat room with cryptographic access control
type PrivateChatRoom struct {
    RoomID      string    // "nakamoto.directmessage"
    OwnerID     peer.ID   // Creator controls access
    AccessKey   []byte    // Symmetric encryption key
}

// Owner creates room and invites users
func CreatePrivateRoom(domain string) *PrivateChatRoom {
    room := &PrivateChatRoom{
        RoomID:    domain,
        OwnerID:   myID,
        AccessKey: generateAES256Key(),
    }
    return room
}

// Invite user with encrypted key
func (room *PrivateChatRoom) InviteUser(userID peer.ID) string {
    invite := EncryptedInvite{
        RoomID:    room.RoomID,
        AccessKey: room.AccessKey,
    }
    encrypted := encrypt(invite, userID.PublicKey)
    return base64(encrypted)  // Share out-of-band
}
```

#### Public Chat Rooms (Moderated)
```go
// Public rooms anyone can join, with moderation
type PublicChatRoom struct {
    RoomID       string
    OwnerID      peer.ID
    Moderators   map[peer.ID]*ModeratorRights
    BannedUsers  map[peer.ID]time.Time  // Peer -> Ban expiry
    RoomRules    string
    RequireStake bool  // Optional: require stake to post
    MinStake     float64  // e.g., 1 NAK to prevent spam
}

type ModeratorRights struct {
    CanBan      bool
    CanDelete   bool
    CanMute     bool
    CanPromote  bool  // Can make others moderators
    AddedBy     peer.ID
    AddedAt     time.Time
}

// Create public room (requires stake to prevent spam room creation)
func CreatePublicRoom(name string, rules string) *PublicChatRoom {
    // Require 100 NAK stake for public room creation
    if !hasStake(myID, 100) {
        return nil  // Must stake to create public rooms
    }

    room := &PublicChatRoom{
        RoomID:     name,
        OwnerID:    myID,
        Moderators: map[peer.ID]*ModeratorRights{
            myID: {CanBan: true, CanDelete: true,
                   CanMute: true, CanPromote: true},
        },
        RoomRules:    rules,
        RequireStake: false,  // Owner decides
    }

    // Register on blockchain (so others can find it)
    blockchain.RegisterPublicRoom(name, myPublicKey)
    return room
}

// Moderation actions
func (room *PublicChatRoom) BanUser(userID peer.ID, duration time.Duration,
                                     moderatorID peer.ID) {
    if !room.Moderators[moderatorID].CanBan {
        return  // Not authorized
    }

    room.BannedUsers[userID] = time.Now().Add(duration)

    // Broadcast ban to network
    banMsg := ModerationAction{
        Type:      "ban",
        UserID:    userID,
        Duration:  duration,
        Reason:    "spam/abuse",
        Moderator: moderatorID,
    }
    p2p.BroadcastToRoom(room.RoomID, banMsg)
}

// Message validation in public rooms
func (room *PublicChatRoom) ValidateMessage(msg OffChainMessage) bool {
    // Check if banned
    if banExpiry, banned := room.BannedUsers[msg.SenderID]; banned {
        if time.Now().Before(banExpiry) {
            return false  // Still banned
        }
        delete(room.BannedUsers, msg.SenderID)  // Ban expired
    }

    // Check stake requirement
    if room.RequireStake {
        if !hasMinimumStake(msg.SenderID, room.MinStake) {
            return false  // Need stake to post
        }
    }

    return true
}

// Add moderator (only owner or authorized mods)
func (room *PublicChatRoom) AddModerator(newModID peer.ID,
                                         byModID peer.ID) {
    if byModID != room.OwnerID &&
       !room.Moderators[byModID].CanPromote {
        return  // Not authorized
    }

    room.Moderators[newModID] = &ModeratorRights{
        CanBan:     true,
        CanDelete:  true,
        CanMute:    true,
        CanPromote: false,  // Only owner can grant this
        AddedBy:    byModID,
        AddedAt:    time.Now(),
    }
}
```

#### File Sharing
```go
// Share files directly between invited peers
fileShare := OffChainMessage{
    MessageType: "file",
    RoomID:      roomID,  // Only room members can access
    Content:     encrypt(file, roomKey),
}
```

#### Shopping Cart Sync
```go
// E-commerce sites sync state across customer's devices
cartSync := OffChainMessage{
    MessageType: "state_sync",
    UserID:      customerID,
    Content:     encrypt(cartState, customerKey),
    VectorClock: deviceClocks,
}
```

### 8.7 Privacy & Security

```go
type MessageSecurity struct {
    // End-to-end encryption
    Encryption      string  // "AES-256-GCM"

    // Perfect forward secrecy
    EphemeralKeys   bool    // New key per session

    // No persistent storage required
    Ephemeral       bool    // Messages can be forgotten

    // No blockchain record
    OffChain        bool    // Complete privacy
}
```

### 8.8 Network Protocol with Obfuscation Support

```go
type NetworkMode string

const (
    ModeGossip       NetworkMode = "gossip"       // Fast, standard
    ModeObfuscated   NetworkMode = "obfuscated"   // Slower, hidden
)

// Adaptive protocol for restrictive networks
type AdaptiveProtocol struct {
    Mode            NetworkMode
    ObfuscationKey  []byte  // Shared secret for traffic shaping
}

// Toggle based on network conditions
func (p *Peer) SendMessage(msg OffChainMessage) {
    if detectRestrictiveNetwork() || userPreference == "obfuscated" {
        sendObfuscated(msg)
    } else {
        sendGossip(msg)
    }
}

// Obfuscation for restrictive networks (China, Iran, etc.)
func sendObfuscated(msg OffChainMessage) {
    // 1. Pad to fixed size (hide message length)
    padded := padToSize(msg, 16384)  // 16KB blocks

    // 2. Shape traffic to look like HTTPS
    wrapped := HTTPSWrapper{
        ContentType: "application/octet-stream",
        Body:        padded,
        FakeHeaders: generateRealisticHeaders(),
    }

    // 3. Random delays to avoid timing analysis
    delay := randomDelay(100, 500)  // 100-500ms
    time.Sleep(delay)

    // 4. Route through random peers (onion routing)
    route := selectRandomRoute(3)  // 3 hops
    sendViaRoute(wrapped, route)
}

// Configuration per region
type RegionalConfig struct {
    China: {
        Mode:           ModeObfuscated,
        MinHops:        3,
        TrafficShaping: "https_mimicry",
        RandomDelay:    true,
    },
    Default: {
        Mode:           ModeGossip,
        MinHops:        1,
        TrafficShaping: "none",
        RandomDelay:    false,
    },
}
```

### 8.9 P2P Discovery Without Fees

Peers discover each other through the existing blockchain P2P network:

```go
// Bootstrap from blockchain seed nodes
func DiscoverPeers() {
    // 1. Connect to seed nodes (same as blockchain)
    seedNodes := []string{
        "seed1.nakamoto.com:9333",
        "seed2.nakamoto.com:9333",
    }

    // 2. Exchange peer lists (no blockchain needed)
    for _, seed := range seedNodes {
        conn := connect(seed)
        peerList := conn.RequestPeerList()

        for _, peer := range peerList {
            tryConnect(peer)
        }
    }
}

// Multi-path redundancy ensures connection success
func ConnectToPeer(targetID peer.ID) (*Connection, error) {
    // Try direct connection
    if conn := tryDirectConnect(targetID); conn != nil {
        return conn, nil
    }

    // Try NAT hole punching
    if conn := tryHolePunch(targetID); conn != nil {
        return conn, nil
    }

    // Relay through mutual peer
    if conn := tryRelay(targetID); conn != nil {
        return conn, nil
    }

    // Store-and-forward fallback (always works)
    return storeAndForward(targetID)
}

// Store-and-forward for offline peers
type StoreAndForward struct {
    messageQueue map[peer.ID][]Message
}

func (s *StoreAndForward) SendMessage(targetID peer.ID, msg Message) {
    // Store on multiple peers
    routingPeers := findPeersCloseTo(targetID)
    for _, peer := range routingPeers[:3] {
        peer.StoreForward(targetID, msg)
    }
}

// Target retrieves when online
func RetrieveMessages() []Message {
    request := MessageRequest{
        RecipientID: myID,
        Signature:   sign(myID, myPrivateKey),
    }

    var messages []Message
    for _, peer := range connectedPeers {
        messages = append(messages,
            peer.GetStoredMessages(request)...)
    }
    return messages
}
```

### 8.10 Conflict-Free Guarantees

The CRDT design guarantees:

1. **No Lost Messages**: Append-only ensures nothing deleted
2. **Eventual Consistency**: All peers converge to same state
3. **No Coordination Needed**: Works even with network partitions
4. **Preserved Intent**: Concurrent operations don't conflict
5. **Deterministic Merge**: Same result regardless of message order
6. **No Spam**: Invitation-only rooms prevent unwanted messages

### 8.11 Implementation Example

```go
// Shopping cart CRDT - handles concurrent updates
type ShoppingCartCRDT struct {
    Items       map[string]ItemCRDT
    VectorClock VectorClock
}

type ItemCRDT struct {
    ProductID   string
    Quantity    GCounter  // Grow-only counter
    Removed     bool      // Tombstone for removals
    LastUpdate  VectorClock
}

// User adds item on phone while spouse adds on laptop
// Both operations preserved, quantities merged
func (cart *ShoppingCartCRDT) Merge(other *ShoppingCartCRDT) {
    for id, item := range other.Items {
        if existing, exists := cart.Items[id]; exists {
            // Merge quantities (GCounter merge is commutative)
            existing.Quantity.Merge(item.Quantity)

            // Last-write-wins for removal
            if item.LastUpdate.HappensBefore(existing.LastUpdate) {
                existing.Removed = item.Removed
            }
        } else {
            cart.Items[id] = item
        }
    }
    cart.VectorClock.Merge(other.VectorClock)
}
```

---

## 9. Blockchain Web Hosting

### 9.1 Domain Registration

Domains are registered as NFTs on the blockchain:

```go
type BlockchainDomain struct {
    DomainName      string    // "amazon.nakamoto"
    OwnerPublicKey  []byte    // Owner's signing key
    ContentRoot     string    // Root content hash
    UpdatePolicy    UpdatePolicy
    CacheRules      CacheRules
}
```

### 9.2 Website Upload Process

1. **Chunking**: Split files into 256KB chunks
2. **Encryption**: AES-256-GCM per chunk
3. **Storage**: Upload to guardians (paid storage)
4. **Metadata**: Create version with all hashes
5. **Signature**: Owner signs version
6. **Publishing**: Announce to network

### 9.3 Access Flow

```
User types "amazon.nakamoto"
    ↓
Check local cache (<1ms)
    ↓
Query peer caches (<100ms)
    ↓
Fetch from guardians (<1000ms)
    ↓
Render webpage
```

### 9.4 Real-Time Updates Without WebSockets

Using gossip protocol for live updates:

```go
// Amazon publishes price change
update := GossipUpdate{
    Type: "price_update",
    Domain: "amazon.nakamoto",
    Data: {"item_123": "$89.99"},
    Signature: amazonSignature,
}

// Propagates through network in <10 seconds
// Users subscribed to topic receive update
// Complete anonymity preserved
```

---

## 9A. Domain Governance & User Websites

### 9A.1 Overview

The Nakamoto network implements a decentralized domain governance system where:
1. **Domains** are created through community validator voting (not purchased)
2. **Usernames** are registered within domains (scoped, not global)
3. **Personal websites** are automatically created with username registration

**URL Scheme:**
```
{domain}/{username}/{path}

Examples:
nakamoto/alice              → User "alice" on nakamoto domain
nakamoto/alice/profile      → Profile page
nakamoto/forum              → System website (dev-funded)
```

Note: Section 9 covers **enterprise** web hosting (purchased domains like "amazon.nakamoto"). This section covers **personal** user websites with community-governed domains.

### 9A.2 Domain Governance

Domains are created exclusively through community voting, not purchase:

```go
type Domain struct {
    Name              string       // e.g., "nakamoto"
    ProposalTxHash    string       // On-chain proposal
    ApprovalBlock     uint64       // When approved
    ProposerAddress   string       // Who proposed
    Status            DomainStatus // active, deprecated, pending
}

type DomainStatus string
const (
    DomainStatusPending    DomainStatus = "pending"
    DomainStatusActive     DomainStatus = "active"
    DomainStatusDeprecated DomainStatus = "deprecated"
)
```

**At Launch:** Only the genesis domain "nakamoto" exists.

**Domain Proposal Requirements:**

| Parameter | Value |
|-----------|-------|
| Minimum Stake | 1,000 NAK (locked during voting) |
| Proposal Fee | 100 NAK (burned) |
| Format | 3-32 chars, lowercase, alphanumeric + hyphens |
| Voting Period | 7 days (~50,400 blocks) |
| Grace Period | 3 days after approval |
| Threshold | 67% validator consensus |

**Domain Lifecycle:**
```
1. Submit Proposal
   ├── Stake minimum 1,000 NAK
   ├── Pay 100 NAK fee (burned)
   └── Domain name validated

2. Voting Period (7 days)
   ├── Validators vote yes/no
   └── Progress visible on-chain

3. Result
   ├── ≥67% yes → Approved (3-day grace)
   └── <67% yes → Rejected (fee lost)

4. Deprecation
   ├── New proposal to deprecate
   ├── Same voting process
   └── If approved, frozen (no new registrations)
```

### 9A.3 Domain-Scoped Usernames

Usernames are unique **per domain**, not globally:

| Domain | Username | Owner | Website URL |
|--------|----------|-------|-------------|
| nakamoto | alice | User A | nakamoto/alice |
| nak | alice | User B | nak/alice |

**Benefits:**
- Short usernames available on new domains
- No global namespace collision
- Each domain is its own community

**Username Registration Pricing:**

| Length | Registration Fee | Renewal Fee |
|--------|-----------------|-------------|
| 3 chars | 1,000 NAK | 500 NAK/year |
| 4 chars | 500 NAK | 250 NAK/year |
| 5 chars | 100 NAK | 50 NAK/year |
| 6+ chars | 10 NAK | 5 NAK/year |

**Username Format:**
- 3-32 characters
- Lowercase letters (a-z), numbers (0-9), hyphens (-)
- Cannot start or end with hyphen
- Cannot contain consecutive hyphens

### 9A.4 User Website Hosting

When a username is registered, a personal website is automatically created:

```go
type UserWebsite struct {
    Domain             string    // "nakamoto"
    Username           string    // "alice"
    SiteID             string    // "nakamoto/alice"
    OwnerAddress       string    // Blockchain address

    // Metadata
    Name               string    // "Alice's Portfolio"
    Description        string    // Site description

    // Storage
    StoragePaidNAK     float64   // NAK paid for storage
    StorageExpiryBlock uint64    // Expiration block
    TotalContentBytes  int64     // Total size

    // Content
    Pages              []UserWebsitePage
}

type UserWebsitePage struct {
    Path        string    // "index.html", "about/team.html"
    ContentHash string    // SHA256 of content
    ContentType string    // "text/html", "image/png"
    SizeBytes   int64
    CreatedAt   string    // ISO timestamp
    UpdatedAt   string    // ISO timestamp
}
```

**Storage Limits:**

| Parameter | Value |
|-----------|-------|
| Max total size | 100 MB per website |
| Max file size | 10 MB per file |
| Supported types | HTML, CSS, JS, JSON, TXT, PNG, JPG, JPEG, GIF, SVG, ICO, WEBP, WOFF, WOFF2, TTF, PDF |
| Storage location | Trunk blockchain (2+1+2 replication) |

**Upload Process:**
1. User uploads file via multipart form
2. Content validated (type, size)
3. SHA256 hash computed
4. Stored on trunk with 2+1+2 replication
5. Metadata recorded on main chain
6. Storage fee deducted from wallet

### 9A.5 Content Serving

Websites are served via the browser's content display:

```
User navigates to "nakamoto/alice/about.html"
    ↓
Router identifies: domain=nakamoto, username=alice, path=about.html
    ↓
Lookup page metadata from main chain
    ↓
Fetch content from trunk storage (by hash)
    ↓
Verify content hash matches
    ↓
Render in browser (sanitized for security)
```

**Security:**
- All content sanitized with DOMPurify
- No inline scripts allowed (XSS prevention)
- System websites bypass sanitization (trusted)

### 9A.6 API Endpoints

**Domain Management (8 endpoints):**
```
GET    /api/domains                      - List active domains
GET    /api/domains/{domain}             - Get domain info
POST   /api/domains/propose              - Submit proposal
POST   /api/domains/deprecate            - Deprecate domain
GET    /api/domains/proposals            - List proposals
GET    /api/domains/proposals/{id}       - Proposal details
POST   /api/domains/proposals/{id}/vote  - Vote on proposal
GET    /api/domains/proposals/{id}/stats - Voting stats
```

**User Website Management (8 endpoints):**
```
POST   /api/usernames/register                                    - Register username
GET    /api/usernames/pricing                                     - Get pricing for username
GET    /api/usernames/validate                                    - Validate username availability
GET    /api/domains/{domain}/usernames/{username}/website         - Get website info
PUT    /api/domains/{domain}/usernames/{username}/website         - Update metadata
POST   /api/domains/{domain}/usernames/{username}/website/upload  - Upload content
DELETE /api/domains/{domain}/usernames/{username}/website/pages/{path} - Delete page
GET    /api/domains/{domain}/usernames/{username}/website/pages   - List pages
GET    /api/domains/{domain}/usernames/{username}/website/stats   - Storage stats
```

**Content Serving (2 endpoints):**
```
GET    /site/{domain}/{username}           - Serve website root
GET    /site/{domain}/{username}/{path}    - Serve specific page
```

### 9A.7 Configuration

```go
const (
    // Domain Governance
    DomainProposalMinStake   = 1000 * 1e8  // 1000 NAK
    DomainProposalFee        = 100 * 1e8   // 100 NAK burned
    DomainVotingPeriodBlocks = 50400       // ~7 days
    DomainGracePeriodBlocks  = 21600       // ~3 days
    DomainVotingThreshold    = 0.67        // 67%

    // Username/Website
    UsernameMinLength   = 3
    UsernameMaxLength   = 32
    WebsiteMaxSizeBytes = 100 * 1024 * 1024  // 100MB
    WebsiteMaxFileSize  = 10 * 1024 * 1024   // 10MB
)

// Username Pricing (in NAK, based on length)
var UsernamePricing = map[int]struct{ Registration, Renewal float64 }{
    3: {1000, 500},  // 3 chars: 1000 NAK registration, 500 NAK/year
    4: {500, 250},   // 4 chars
    5: {100, 50},    // 5 chars
    6: {10, 5},      // 6+ chars (default)
}
```

---

## 10. Technical Specifications and Configuration

### 10.1 System Architecture Separation

The Nakamoto ecosystem maintains strict separation between layers:

```go
type NakamotoEcosystem struct {
    // On-Chain (Blockchain)
    Blockchain    *Blockchain      // Transactions, consensus, storage metadata
    Storage       *StorageSystem   // Paid guardian storage (NAK fees)

    // Off-Chain (P2P)
    Cache         *CacheSystem     // Free voluntary sharing (no blockchain)
    Messaging     *CRDTMessaging   // Direct P2P messages (no blockchain)
    WebHosting    *DomainSystem    // Cached websites (metadata on-chain)
}
```

### 10.2 Configuration Parameters

```json
{
  "cache": {
    "max_size_gb": 100,
    "min_peers_for_content": 3,
    "enable_lru_eviction": false,
    "content_request_timeout": 5000,
    "parallel_fetch_peers": 5
  },
  "crdt": {
    "max_message_size": 1048576,
    "gc_interval": 3600,
    "max_vector_clock_size": 100,
    "store_forward_retention": 86400,
    "max_stored_messages": 1000,
    "public_room_creation_stake": 100,
    "public_room_post_stake": 1,
    "ban_duration_default": 86400
  },
  "domain": {
    "min_registration_fee": 100,
    "renewal_period_blocks": 2190000,
    "max_domain_length": 63,
    "max_subdomains": 10,
    "version_snapshot_interval": 100
  },
  "p2p": {
    "max_connections": 1000,
    "connection_timeout": 10000,
    "nat_traversal_timeout": 5000,
    "relay_max_bandwidth_mb": 100,
    "store_forward_peers": 3
  },
  "obfuscation": {
    "enabled": false,
    "auto_detect_restrictive": true,
    "traffic_padding_size": 16384,
    "random_delay_min_ms": 100,
    "random_delay_max_ms": 500,
    "minimum_hops": 3,
    "traffic_shaping": "https_mimicry",
    "regions": {
      "china": {"auto_enable": true, "min_hops": 3},
      "iran": {"auto_enable": true, "min_hops": 3},
      "default": {"auto_enable": false, "min_hops": 1}
    }
  }
}
```

### 10.3 Security and Spam Prevention

#### Cache Content Verification
```go
// Use merkle proofs without blockchain access
type CacheVerification struct {
    ContentHash  [32]byte
    MerkleProof  []MerkleNode
    DomainRoot   [32]byte  // From domain registration
}

func VerifyCachedContent(content []byte, proof CacheVerification) bool {
    // Verify against domain's merkle root
    return verifyMerkleProof(sha256(content), proof.MerkleProof, proof.DomainRoot)
}
```

#### Message Spam Prevention
```go
// Invitation-only rooms eliminate spam
type AntiSpam struct {
    // No stake requirements for basic messaging
    // Room owner controls access via encrypted invites
    RoomAccessControl map[string][]byte  // RoomID -> AccessKey

    // Optional: Stake for creating public rooms
    PublicRoomStake   float64  // 100 NAK to create public room
}
```

### 10.4 Performance Guarantees

#### Cache System
- Request-based discovery (no constant broadcasting)
- Parallel fetching from multiple peers
- Guardian fallback for guaranteed availability
- Zero blockchain interaction

#### CRDT Messaging
- Direct P2P without blockchain overhead
- Store-and-forward for offline delivery
- Invitation-based rooms (no spam)
- Multi-path connection redundancy

#### Web Hosting
- Domain metadata on blockchain (one-time)
- Content served via cache network
- Differential updates for efficiency
- Version snapshots prevent bloat

### 10.5 API Endpoints

```
# Cache System (P2P only, no blockchain)
/p2p/cache/request/{content_id}    - Request content from peers
/p2p/cache/provide/{content_id}    - Provide content to requester
/p2p/cache/stats                   - Local cache statistics

# CRDT Messaging (P2P only)
/p2p/message/send                  - Send off-chain message
/p2p/message/retrieve              - Get stored messages
/p2p/room/create                   - Create chat room
/p2p/room/invite                   - Generate invite key

# Domain System (Hybrid)
/api/domain/register               - Register domain (blockchain)
/p2p/domain/content/{domain}       - Fetch content (P2P cache)
/api/domain/update                 - Update version (blockchain)
```

---

## Implementation Priorities

### Phase 1: Core Systems (Storage Complete ✅)
- [x] Implement storage system with 2+1+2 replication
- [x] Guardian chunk assignment
- [x] Cryptographic proof verification
- [x] Blockchain metadata storage
- [ ] Implement cache storage with indefinite persistence
- [ ] Add CRDT message system
- [ ] Create version management

### Phase 2: Web Hosting
- [ ] Domain registration system
- [ ] Website upload protocol
- [ ] Version chain management

### Phase 3: Privacy Features
- [ ] Gossip protocol (no WebSockets)
- [ ] Anonymous updates
- [ ] Encrypted peer communication

### Phase 4: Optimization
- [ ] Parallel fetching
- [ ] Differential updates
- [ ] Background updates

---

## 11. Blockchain-Based Update Distribution System

### 11.1 Overview

The Nakamoto blockchain serves as its own update distribution network, storing and distributing software updates directly on-chain, eliminating dependence on external servers including Tor.

### 11.2 Update Storage Architecture

```go
type BlockchainUpdate struct {
    Version         string              // "2.6.0"
    UpdateType      UpdateType          // Binary, Shell, or Both
    RequiresFork    bool                // Consensus-breaking changes
    ForkHeight      uint64              // Activation block for consensus changes
    Files           map[string][]byte   // Update files
    Manifest        UpdateManifest      // Metadata and requirements
    Signatures      []MultiSig          // Developer multi-sig authorization
    Channel         string              // "stable" or "beta"
}

type UpdateType string
const (
    UpdateTypeBinary UpdateType = "binary"     // Blockchain node updates
    UpdateTypeShell  UpdateType = "shell"      // Shell interface updates
    UpdateTypeBoth   UpdateType = "both"       // Combined update
)
```

### 11.3 Fork-Based vs Non-Fork Updates

**Fork-Required Updates (Consensus Changes):**
- Changes to validator system architecture
- Modifications to consensus thresholds
- Block structure alterations
- Reward distribution changes
- Storage requirement modifications
- Bitcoin conversion rate adjustments

**Non-Fork Updates (Backward Compatible):**
- New shell commands and features
- API endpoint additions
- UI improvements
- Performance optimizations
- Non-consensus bug fixes
- Documentation updates

### 11.4 Update Upload Process

```go
// Developer uploads via multi-sig authority
type UpdateUpload struct {
    // Step 1: Package components
    Binary          []byte    // Compiled node binary
    ShellModules    []byte    // Shell scripts tarball
    Migration       []byte    // Database migration scripts

    // Step 2: Create manifest
    Manifest        UpdateManifest

    // Step 3: Multi-sig authorization
    KeyShards       []KeyShard  // M-of-N signature required

    // Step 4: Storage on blockchain
    StoragePath     string      // "/updates/v2.6.0/"
    Transactions    []TxHash    // Multiple txs for large files
}
```

### 11.5 Automatic Update Detection

```go
type UpdateMonitor struct {
    CheckInterval   time.Duration       // Default: 1 hour
    UpdateChannel   string              // "stable" or "beta"
    AutoUpdate      bool                // User preference

    // Monitors blockchain for updates
    CheckForUpdates() (*BlockchainUpdate, bool)

    // Downloads from blockchain storage
    DownloadUpdate(version string) error

    // Applies based on update type
    ApplyUpdate(update *BlockchainUpdate) error
}
```

### 11.6 P2P Update Notification Protocol

Updates are distributed through a push-pull model. When a developer uploads an update and calls the announce endpoint, the uploading node broadcasts an `UpdateNotification` message to all connected peers via the `/nakamoto/update/notification/1.0.0` protocol. Receiving nodes immediately trigger their local `UpdateMonitor` to check for and download the update, bypassing the normal polling interval.

```
Developer → Upload → Local Node → P2P Broadcast (notification)
                                       ↓
                              All Connected Peers
                                       ↓
                           TriggerCheck() → CheckForUpdates()
                                       ↓
                           Download chunks via P2P chunk protocol
                                       ↓
                           Re-broadcast notification to their peers
```

**Protocol Flow:**
1. Developer uploads signed update via `/api/updates/upload`
2. Developer announces via `/api/update/announce` (requires JWT auth)
3. Node broadcasts `UpdateNotification{Version, Channel, Timestamp}` to all peers
4. Each peer's `UpdateProtocolHandler` receives the notification
5. Handler calls `UpdateMonitor.TriggerCheck()` for immediate update detection
6. `UpdateMonitor` fetches the update manifest and downloads chunks from peers

**P2P Protocols:**
| Protocol | Direction | Purpose |
|----------|-----------|---------|
| `/nakamoto/update/notification/1.0.0` | Push | Announce new update availability |
| `/nakamoto/update/availability/1.0.0` | Pull | Query peer for cached updates |
| `/nakamoto/update/chunk/1.0.0` | Pull | Request specific update chunk |

**Fallback:** Nodes that miss the notification still discover updates through periodic polling (1 hour for stable channel, 30 minutes for beta).

### 11.7 Fork Activation Protocol

```go
type UpdateFork struct {
    Name            string      // "nakamoto-update-v2.6.0"
    Version         string      // "2.6.0"
    ActivationHeight uint64     // Block height for activation
    GracePeriod     uint64      // Blocks before mandatory

    // Consensus changes
    Changes         []ConsensusChange

    // Migration requirements
    Migration       MigrationScript
}

// At activation height
func (bc *Blockchain) HandleUpdateFork(height uint64) {
    if fork := bc.GetForkAtHeight(height); fork != nil {
        if !IsVersionCompatible(currentVersion, fork.Version) {
            // Stop validation until updated
            bc.PauseValidation()
            bc.TriggerMandatoryUpdate(fork.Version)
        }
    }
}
```

### 11.8 Beta Testing Framework

```go
type BetaChannel struct {
    OptInRequired   bool        // Users must explicitly opt-in
    UpdatePath      string      // "/updates/beta/"

    // Separate channel for testing
    TestNetwork     bool        // Can use testnet

    // Promotion to stable
    PromotionCriteria struct {
        MinTestDays     int     // Minimum beta testing period
        MinTestNodes    int     // Minimum nodes running beta
        MaxIssues       int     // Maximum critical issues allowed
    }
}
```

### 11.9 Migration System

```go
type MigrationManager struct {
    // Automatic migrations during updates
    Migrations      map[string]MigrationScript

    // Version-specific migrations
    "2.6.0": {
        // Convert flat validators to hierarchical
        MigrateValidatorTypes()

        // Update storage requirements
        UpdateGuardianRequirements()

        // Initialize shadow validation
        InitializeShadowValidation()
    }
}
```

### 11.10 Update Security

- **Multi-Signature Requirement**: All updates require M-of-N developer signatures
- **Cryptographic Verification**: SHA256 hashes for all files
- **Version Pinning**: Nodes cannot downgrade without explicit override
- **Rollback Protection**: Previous version retained for emergency rollback

### 11.11 User Experience

```bash
# Non-fork update notification
[NAKAMOTO] Update v2.5.1 available (non-consensus)
[NAKAMOTO] Changes: New shell commands, bug fixes
[NAKAMOTO] Auto-installing in 5 minutes...

# Fork update notification  
[NAKAMOTO] ⚠️ Consensus fork "v2.6.0" activates at block 200,000
[NAKAMOTO] Current block: 199,800 (200 blocks remaining)
[NAKAMOTO] Update will be applied automatically at activation
```

### 11.12 Developer Shell Commands

```bash
# Push update (requires multi-sig)
./scripts/developer_shell.sh push_update \
    --version 2.6.0 \
    --channel stable \
    --type both \
    --binary dist/nakamoto-2.6.0.tar.gz \
    --shell dist/shell-2.6.0.tar.gz \
    --requires-fork true \
    --activation-height 200000

# Monitor update adoption
./scripts/developer_shell.sh update_status
# Output:
# Version 2.6.0 Adoption:
#   Total Nodes: 1,247
#   Updated: 1,089 (87.3%)
#   Failed: 23 (1.8%)
#   Pending: 135 (10.8%)
```

### 11.13 Rollback System

```go
type RollbackManager struct {
    BackupVersions  int     // Keep 3 previous versions
    BackupLocation  string  // Local storage path
    
    // Automatic backup before update
    BackupCurrent() error
    
    // User-initiated rollback
    RollbackTo(version string) error
}
```

User commands:
```bash
# Rollback to previous version
rollback_update

# Rollback to specific version  
rollback_update 2.5.0

# List available rollback versions
rollback_list
```

### 11.14 Update Acknowledgments

```go
type UpdateAck struct {
    NodeID      peer.ID
    Version     string
    Success     bool
    Error       string
    Timestamp   time.Time
}

// Nodes automatically report after update
func (n *Node) SendUpdateAck(success bool, err error) {
    ack := UpdateAck{
        NodeID:    n.ID,
        Version:   n.Version,
        Success:   success,
        Error:     err.String(),
        Timestamp: time.Now(),
    }
    
    // Anonymous telemetry if enabled
    if n.Config.Telemetry {
        n.SendAck(ack)
    }
}
```

### 11.15 Platform Distribution

```go
type PlatformBinaries struct {
    LinuxAMD64   []byte  // Linux binary
    DarwinAMD64  []byte  // macOS Intel
    DarwinARM64  []byte  // macOS Apple Silicon  
    WindowsAMD64 []byte  // Windows binary
}

// Storage structure
/updates/2.6.0/
├── manifest.json
├── linux-amd64/
├── darwin-amd64/
├── darwin-arm64/
├── windows-amd64/
└── shell/  // Shell modules (cross-platform)
```

### 11.16 Access Control

```go
type UpdateAccess struct {
    // Only developer wallet can upload
    AuthorizedWallet  string  // Your wallet address
    
    // Can be changed via update
    Modifiable        bool    // True - allows transfer of authority
    
    // Multi-sig prevents single point of failure
    RequiredSignatures int    // M-of-N shards needed
}
```

---

## Summary of Changes from v2.5.0

1. **Added Blockchain-Based Update Distribution** (Section 11) - Decentralized software updates stored on-chain
2. **Fork-Based Protocol Updates** - Consensus changes distributed via fork mechanism
3. **Multi-Signature Authorization** - Secure update authentication with M-of-N signatures
4. **Beta Testing Framework** - Separate channel for testing updates before stable release
5. **Automatic Migration System** - Handles breaking changes during updates
6. **Rollback Protection** - Keep 3 previous versions for emergency rollback
7. **Update Acknowledgments** - Nodes report success/failure for monitoring
8. **Platform-Specific Distribution** - Separate binaries for Linux, macOS, Windows
9. **Developer Shell Integration** - Commands for pushing and monitoring updates
10. **Access Control** - Only developer wallet can upload updates (transferable)

## Summary of Changes from v2.4.0

1. **Added Distributed Cache System** (Section 7) - Separate from storage, never expires, completely free
2. **Added Off-Chain Messaging** (Section 8) - CRDTs with both private and public moderated rooms
3. **Added Web Hosting** (Section 9) - Complete blockchain website system
4. **Added Technical Specifications** (Section 10) - Configuration, security, and API details
5. **Added Obfuscation Protocol** (Section 8.8) - Traffic shaping for restrictive networks
6. **Added Public Chat Rooms** (Section 8.6) - Moderated rooms with ban/mute capabilities
7. **Clarified Cache Economics** - Free voluntary sharing, no NAK fees
8. **Clarified P2P Discovery** - Uses existing blockchain network, no fees
9. **Added Store-and-Forward** - Ensures message delivery even when peers offline
10. **Pull-Based Content Discovery** - Request when needed, not constant broadcasting
11. **Added Blockchain-Based Update System** (Section 11) - Decentralized update distribution

Key Technical Innovations:
- **Dual Protocol Modes**: Toggle between fast gossip and obfuscated modes
- **Regional Adaptation**: Auto-detect restrictive networks (China, Iran) and enable obfuscation
- **Moderation System**: Public rooms with moderator hierarchy and time-based bans
- **Traffic Shaping**: Messages disguised as HTTPS traffic with random delays
- **Hybrid Chat System**: Private invitation-only rooms AND public moderated rooms
- **Blockchain Update Distribution**: Updates stored and distributed via blockchain

Key Technical Clarifications:
- **Storage**: Paid with NAK, guardian-managed, blockchain-tracked
- **Cache**: Free voluntary P2P, no blockchain involvement
- **Private Messages**: Invitation-only rooms, no spam possible
- **Public Messages**: Moderated rooms with optional stake requirements
- **Discovery**: Through existing P2P network, no blockchain transactions
- **Obfuscation**: Optional for users in restrictive regions, sacrifices speed for privacy
- **Updates**: Blockchain-stored, multi-sig authorized, fork-activated for consensus changes

---

## 12. USB Hardware-Based Two-Factor Authentication System

### 12.1 Overview

The USB 2FA system implements a hardware-based authentication mechanism using USB devices as physical security keys. This system operates in two modes: automatic verification (USB as physical key) and manual token generation (legacy compatibility), providing seamless authentication while maintaining backward compatibility with existing USB registrations.

### 12.2 Architecture

```go
type USB2FASystem struct {
    // Core components
    AutoVerifier    *USBAutoVerifier    // Automatic USB verification service
    Manager         *USB2FAManager      // Core 2FA management
    Monitor         *USBMonitor         // USB device detection

    // Security layers
    Encryption      AES256GCM           // Device secret encryption
    TokenGeneration HMAC_SHA256         // Counter-based token generation
    SessionManager  *SessionCache       // Active session management
}

type USBAutoVerifier struct {
    usbManager      *USB2FAManager
    monitor         *USBMonitor
    activeSessions  map[string]*AutoSession  // deviceID -> session
    deviceCache     map[string]*CachedDevice // serialNumber -> device info
    mode            AutoVerificationMode      // Auto or Manual
    monitorInterval time.Duration             // 1 second polling
}
```

### 12.3 Device Registration

```go
type RegisteredUSB struct {
    DeviceID        string          // SHA256(serial:vendor:product)[:16]
    WalletID        string          // Associated wallet identifier
    FriendlyName    string          // User-defined device name
    MasterSecret    []byte          // 32-byte secret (encrypted)
    Counter         uint64          // Monotonic counter for tokens
    IsMaster        bool            // Master device capability
    CreatedAt       time.Time       // Registration timestamp
    LastUsed        time.Time       // Last authentication time

    // Hardware identifiers
    SerialNumber    string          // USB serial number
    VendorID        string          // USB vendor ID
    ProductID       string          // USB product ID
}

// Registration process
func RegisterUSBDevice(wallet *Wallet, device *USBDevice) error {
    // Generate master secret
    masterSecret := GenerateSecureRandom(32)

    // Encrypt with wallet-derived key
    encryptedSecret := EncryptWithKey(masterSecret, wallet.DerivedKey)

    // Store registration
    registration := RegisteredUSB{
        DeviceID:     GenerateDeviceID(device),
        WalletID:     wallet.ID,
        MasterSecret: encryptedSecret,
        Counter:      0,
        // ... hardware identifiers
    }

    return StoreRegistration(registration)
}
```

### 12.4 Automatic USB Verification Mode

#### 12.4.1 Continuous Monitoring

```go
type USBMonitor struct {
    detector        *USBDetector
    knownDevices    map[string]bool
    scanInterval    time.Duration   // 1 second
}

func (m *USBMonitor) MonitorLoop() {
    ticker := time.NewTicker(m.scanInterval)
    for range ticker.C {
        devices := m.DetectDevices()
        m.ProcessDeviceChanges(devices)
    }
}

// Device detection using native USB libraries
func (d *USBDetector) EnumerateNakamotoDevices() ([]*NakamotoUSBDevice, error) {
    // Platform-specific USB enumeration
    // Linux: libusb, Windows: WinUSB, macOS: IOKit
    devices := NativeUSBEnumerate()

    // Filter for registered devices
    return FilterRegisteredDevices(devices)
}
```

#### 12.4.2 Automatic Session Creation

```go
type AutoSession struct {
    SessionID       string          // Unique session identifier
    WalletID        string          // Associated wallet
    DeviceID        string          // USB device ID
    AccessLevel     AccessLevel     // ReadOnly, WriteAccess, FullAccess
    CreatedAt       time.Time       // Session creation
    LastActivity    time.Time       // Last operation
    ValidUntil      time.Time       // Expiry (30 minutes from USB removal)
    AutoToken       string          // Invisible automatic token
}

func (av *USBAutoVerifier) CreateSession(device *USBDevice) *AutoSession {
    registered := av.GetRegisteredDevice(device.DeviceID)

    // Generate automatic token (invisible to user)
    autoToken := av.GenerateAutoToken(registered.MasterSecret, time.Now().Unix())

    session := &AutoSession{
        SessionID:    GenerateSessionID(),
        WalletID:     registered.WalletID,
        DeviceID:     device.DeviceID,
        AccessLevel:  DetermineAccessLevel(registered),
        CreatedAt:    time.Now(),
        ValidUntil:   time.Now().Add(30 * time.Minute),
        AutoToken:    autoToken,
    }

    av.activeSessions[device.DeviceID] = session
    return session
}
```

### 12.5 Counter-Based Token Generation

```go
type TokenGenerator struct {
    Algorithm       string          // "HMAC-SHA256"
    TokenLength     int             // 6 digits
    CounterWindow   uint64          // Accept ±10 counter values
}

func GenerateToken(secret []byte, counter uint64) string {
    // HMAC-based One-Time Password (HOTP)
    mac := hmac.New(sha256.New, secret)
    binary.Write(mac, binary.BigEndian, counter)
    hash := mac.Sum(nil)

    // Dynamic truncation
    offset := hash[len(hash)-1] & 0x0f
    code := binary.BigEndian.Uint32(hash[offset:]) & 0x7fffffff

    // 6-digit token
    return fmt.Sprintf("%06d", code%1000000)
}

func VerifyToken(token string, secret []byte, currentCounter uint64) bool {
    // Check counter window for clock skew tolerance
    for i := currentCounter - 10; i <= currentCounter + 10; i++ {
        if GenerateToken(secret, i) == token {
            // Update counter to prevent replay
            UpdateCounter(i + 1)
            return true
        }
    }
    return false
}
```

### 12.6 API Endpoints

```go
// Auto mode endpoints
router.HandleFunc("/api/usb/2fa/auto/status", handleAutoStatus)
router.HandleFunc("/api/usb/2fa/auto/enable", handleEnableAutoMode)
router.HandleFunc("/api/usb/2fa/auto/disable", handleDisableAutoMode)
router.HandleFunc("/api/usb/2fa/auto/authenticate", handleAutoAuthenticate)
router.HandleFunc("/api/usb/2fa/auto/presence", handleCheckPresence)

// Response structures
type AutoStatusResponse struct {
    Success             bool            `json:"success"`
    AutoModeEnabled     bool            `json:"auto_mode_enabled"`
    ConnectedDevices    int             `json:"connected_devices"`
    DeviceIDs           []string        `json:"device_ids"`
    Message             string          `json:"message"`
}

type AutoAuthResponse struct {
    Success             bool            `json:"success"`
    SessionID           string          `json:"session_id,omitempty"`
    AccessLevel         string          `json:"access_level,omitempty"`
    ValidUntil          time.Time       `json:"valid_until,omitempty"`
    Error               string          `json:"error,omitempty"`
}
```

### 12.7 Developer Wallet Integration

```go
type DeveloperWalletManager struct {
    WalletID            string          // Fixed: "developer-wallet"
    USB2FAManager       *USB2FAManager
    EmergencyAuthority  *EmergencyForkAuthority

    // Multi-signature for critical operations
    RequiredSignatures  int             // M-of-N signatures
    KeyShards          []KeyShard      // Distributed key shards
}

// USB authentication for developer operations
func (dm *DeveloperWalletManager) AuthorizeOperation(op Operation) error {
    // Check USB presence
    if !dm.USB2FAManager.IsUSBConnected(dm.WalletID) {
        return ErrUSBNotConnected
    }

    // Verify passphrase for sensitive operations
    if op.IsSensitive() {
        if err := dm.VerifyPassphrase(op.Passphrase); err != nil {
            return err
        }
    }

    // Multi-sig for emergency operations
    if op.Type == EmergencyFork {
        return dm.VerifyMultiSig(op.Signatures)
    }

    return nil
}
```

#### 12.7.1 USB Hardware Shard Implementation (Shard 0)

The developer wallet implements a **3-of-6 threshold multi-signature system** where Shard 0 is a hardware shard stored on a USB device with automatic signing capabilities.

##### Architecture

```go
type USBKeyShard struct {
    Index               int       `json:"index"`            // Always 0 for USB shard
    PublicKey           string    `json:"public_key"`       // Ed25519 public key
    EncryptedPrivateKey string    `json:"encrypted_private_key"` // Encrypted with USB MasterSecret
    DeviceID            string    `json:"device_id"`        // USB device unique identifier
    WalletID            string    `json:"wallet_id"`        // Associated developer wallet
    Salt                string    `json:"salt"`             // AES-256-GCM encryption salt
    Nonce               string    `json:"nonce"`            // AES-256-GCM nonce
    Created             time.Time `json:"created"`
    LastUsed            time.Time `json:"last_used"`
    IsHardwareShard     bool      `json:"is_hardware_shard"` // Always true
    Version             string    `json:"version"`          // Shard format version
}

type USBShardProvider struct {
    usb2faManager   *USB2FAManager
    cachedShard     *USBKeyShard      // Cached when USB connected
    cachedPrivateKey string            // Decrypted private key (in memory only)
    storagePath     string             // Path on USB device
}
```

##### 3-of-6 Multi-Signature System

**Shard Distribution:**
- **Shard 0**: USB Hardware Shard (auto-signs when USB connected, no passphrase)
- **Shards 1-5**: Software Shards (passphrase-protected, stored encrypted)

**Authorization Modes:**
- **Normal Operation**: USB + 2 software passphrases (e.g., 0,1,2 or 0,3,5)
- **Recovery Mode**: 3 software passphrases (e.g., 1,2,3 or 2,4,5) - if USB lost

##### Initialization

```go
func (dss *DeveloperSequentialSigner) InitializeKeySharedsWithUSB(
    passphrases []string,      // 5 passphrases for software shards 1-5
    usbDeviceID string,         // USB device unique ID
    usbWalletID string,         // Wallet ID for USB binding
    usbMasterSecret []byte,     // USB device 32-byte MasterSecret
) error
```

**Process:**
1. Validates 5 software shard passphrases (16+ characters each)
2. Generates USB hardware shard (Shard 0) using USBShardProvider
3. Generates software shards 1-5 with individual passphrase encryption
4. Stores USB shard public key only (private key encrypted on USB device)
5. Stores software shards with AES-256-GCM encrypted private keys

##### Sequential Signing with USB

```go
func (dss *DeveloperSequentialSigner) SignTransferRequest(
    request *TransferSignatureRequest,
    shardPassphrases map[int]string,  // Only for software shards (1-5)
    shardIndices []int,                // Must include 3 shards (any combination)
) (*SequentialSignatureResult, error)
```

**Automatic USB Signing:**
- **USB Connected**: Shard 0 automatically signs without passphrase
- **USB Disconnected**: Returns error if Shard 0 requested, allows recovery mode with shards 1-5

**Signature Chain:**
1. Sort shard indices (low to high): e.g., [0,2,4] → ensures consistent order
2. For each shard in sequence:
   - **Shard 0 (USB)**: Auto-sign with cached USB private key (if connected)
   - **Shards 1-5 (Software)**: Decrypt with passphrase, then sign
3. Build cryptographic chain: Each signature signs previous signature's hash
4. Final result includes all signatures + chain integrity proof

##### Security Properties

**USB Shard Protection:**
- Private key encrypted with USB device MasterSecret (AES-256-GCM)
- Shard only accessible when physical USB device connected
- Auto-clears cache when USB disconnected
- Integration with USB 2FA manager for device verification

**Recovery Guarantees:**
- System remains operational if USB lost (use 3 software shards)
- No single point of failure
- Threshold security: any 3 of 6 shards can authorize

**Test Coverage:**
- 20 combination tests (all 3-of-6 permutations)
- 10 with USB (0,1,2 through 0,4,5)
- 10 without USB recovery mode (1,2,3 through 3,4,5)
- USB disconnection error handling
- Signature chain verification

##### Example Usage

**Normal Operation (USB + 2 Software Shards):**
```go
// USB connected, use shards 0, 1, 3
shardPassphrases := map[int]string{
    1: "passphrase-for-shard-1",
    3: "passphrase-for-shard-3",
}
result, err := signer.SignTransferRequest(request, shardPassphrases, []int{0, 1, 3})
// Shard 0 auto-signs, shards 1 and 3 require passphrases
```

**Recovery Mode (USB Lost, Use 3 Software Shards):**
```go
// USB not available, use shards 2, 4, 5
shardPassphrases := map[int]string{
    2: "passphrase-for-shard-2",
    4: "passphrase-for-shard-4",
    5: "passphrase-for-shard-5",
}
result, err := signer.SignTransferRequest(request, shardPassphrases, []int{2, 4, 5})
// All three shards require passphrases, USB not needed
```

##### Implementation Files

- **Core Implementation**: `internal/core/usb_shard_provider.go` (460 lines)
- **Sequential Signing**: `internal/core/developer_wallet_sequential_signing.go` (updated)
- **Unit Tests**: `internal/core/usb_shard_provider_test.go` (230 lines)
- **Integration Tests**: `internal/core/developer_wallet_sequential_signing_test.go` (400+ lines)
- **Specification**: `USB_SHARD_0_IMPLEMENTATION_SPEC.md`

### 12.8 Security Model

#### 12.8.1 Threat Mitigation

```go
type SecurityMeasures struct {
    // Physical security
    USBPresenceRequired     bool        // USB must be connected
    PassphraseRequired      bool        // Additional passphrase for sensitive ops

    // Cryptographic security
    SecretEncryption        string      // "AES-256-GCM"
    TokenAlgorithm          string      // "HMAC-SHA256"

    // Replay protection
    MonotonicCounter        bool        // Counter never decreases
    TokenWindow             uint64      // Accept ±10 counter values

    // Session security
    SessionTimeout          time.Duration   // 30 minutes
    AutoLogoutOnRemoval     bool           // Immediate session termination
}
```

#### 12.8.2 Attack Prevention

```go
type AttackPrevention struct {
    // Brute force protection
    MaxFailedAttempts       int         // 5 attempts
    LockoutDuration         time.Duration   // 15 minutes

    // Token replay prevention
    UsedTokenCache          *LRUCache   // Recent tokens cached
    CounterValidation       bool        // Strict counter checking

    // USB spoofing prevention
    DeviceFingerprinting    bool        // Hardware ID verification
    RegistrationRequired    bool        // Only registered devices accepted
}
```

### 12.9 Backward Compatibility

```go
type CompatibilityLayer struct {
    // Manual mode support (legacy)
    ManualTokenGeneration   bool
    ManualTokenVerification bool

    // Migration path
    AutoMigration           bool        // Automatic migration to auto mode
    PreserveRegistrations   bool        // Existing USB registrations work

    // API compatibility
    LegacyEndpoints         bool        // /api/usb/2fa/* still functional
    ResponseMapping         bool        // Map new fields to old format
}

// Seamless migration
func MigrateToAutoMode(manager *USB2FAManager) error {
    // All existing registrations automatically work
    manager.EnableAutoMode()

    // Start monitoring service
    manager.autoVerifier.Start()

    // Preserve manual mode as fallback
    manager.manualModeEnabled = true

    return nil
}
```

### 12.10 Performance Characteristics

```go
type PerformanceMetrics struct {
    // USB detection
    ScanInterval            time.Duration   // 1 second
    DetectionLatency        time.Duration   // <100ms typical

    // Session management
    SessionCreation         time.Duration   // <10ms
    TokenGeneration         time.Duration   // <1ms

    // Memory usage
    SessionCacheSize        int             // Max 1000 sessions
    DeviceCacheSize         int             // Max 100 devices

    // Scalability
    MaxConcurrentUSB        int             // 10 devices per node
    MaxSessionsPerDevice    int             // 1 active session
}
```

### 12.11 Shell Integration

```bash
# Developer shell commands
./scripts/developer_shell.sh usb-2fa-auto-status      # Check auto mode
./scripts/developer_shell.sh usb-2fa-auto-enable      # Enable auto mode
./scripts/developer_shell.sh usb-2fa-auto-disable     # Disable auto mode
./scripts/developer_shell.sh usb-2fa-auto-check <wallet>  # Check USB presence
./scripts/developer_shell.sh usb-2fa-auto-auth <wallet> <passphrase>  # Authenticate

# Output format
USB Auto Mode Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Auto Mode: ENABLED
Connected Devices: 1
Device: SanDisk Cruzer (developer-wallet)
Access Level: FullAccess
Session Valid: 29 minutes remaining
```

### 12.12 Implementation Files

```
internal/core/
├── usb_auto_verifier.go       // Automatic verification service
├── usb_2fa_manager.go          // Core 2FA management (modified)
└── developer_wallet_manager.go // Developer operations integration

internal/api/
└── usb_2fa_routes.go           // API endpoints (modified)

internal/security/
├── usb_detector.go             // Native USB detection
├── usb_detector_linux.go       // Linux-specific (libusb)
├── usb_detector_windows.go     // Windows-specific (WinUSB)
└── usb_detector_darwin.go      // macOS-specific (IOKit)

scripts/
└── developer_shell.sh          // Shell commands (modified)
```

### 12.13 Future Enhancements

- **WebAuthn Integration**: Support for FIDO2/WebAuthn standards
- **Biometric Enhancement**: Fingerprint-enabled USB keys
- **Mobile Device Support**: Smartphone as USB key via NFC
- **Hardware Wallet Integration**: Ledger/Trezor as 2FA devices
- **Distributed USB Registry**: On-chain USB registration storage

---

## 13. Network Mode Architecture (Testnet vs Mainnet)

### 13.1 Overview

Nakamoto implements a strict separation between testnet and mainnet environments. These are completely independent networks with no migration path or cross-network interaction.

| Parameter | Testnet | Mainnet |
|-----------|---------|---------|
| **Token Symbol** | tNAK | NAK |
| **Bitcoin Network** | testnet3 | mainnet |
| **Chain ID** | 2 | 1 |
| **Network ID** | nakamoto-testnet | nakamoto-mainnet |
| **Developer Control** | Yes (phased) | No (PhaseMature from genesis) |
| **Min Validator Stake** | 21 tNAK | 1000 NAK |
| **BTC Confirmations** | 3 | 6 |

### 13.2 Design Rationale

**Complete Separation**: Testnet and mainnet operate as entirely separate blockchains. There is no token migration, no shared state, and no cross-network transaction capability. This design:

1. **Prevents Accidental Value Transfer**: Clear visual distinction (tNAK vs NAK) prevents confusion
2. **Allows Aggressive Testing**: Developers can experiment on testnet without mainnet risk
3. **Simplifies Security Audits**: Each network has clearly defined boundaries
4. **Enables Independent Evolution**: Networks can diverge for testing purposes

### 13.3 Testnet Mode (tNAK)

Testnet provides a safe environment for development and testing:

**Developer Control (Phased)**:
- **Development Phase** (0-6 months): Full emergency authority capabilities
- **Transition Phase** (6-12 months): Reduced intervention powers
- **Mature Phase** (12+ months): Community-driven governance

**Bitcoin Integration**: Uses Bitcoin testnet3 (tBTC) for conversions at 1:1 satoshi parity.

**Lower Thresholds**:
- Minimum validator stake: 21 tNAK (accessible for testing)
- Lower confirmation requirements
- Faster block finality testing

### 13.4 Mainnet Mode (NAK)

Mainnet is designed for production use with zero developer control from genesis:

**No Developer Control**: The EmergencyAuthorityV2 system initializes directly to `PhaseMature`, bypassing all development and transition phases. This is enforced at the code level:

```go
func NewEmergencyAuthorityV2(config *EmergencyAuthorityConfig) *EmergencyAuthorityV2 {
    if config.NetworkMode.IsMainnet() {
        phase = PhaseMature  // ZERO developer intervention capability
    }
}
```

**Bitcoin Integration**: Uses Bitcoin mainnet (BTC) for conversions at 1:1 satoshi parity:
- 1 BTC = 100,000,000 satoshis = 100,000,000 NAK
- 6 confirmation requirement for security

**Full Security Requirements**:
- Minimum validator stake: 1000 NAK
- Full consensus thresholds enforced
- Production-grade security parameters

### 13.5 Bootstrap Window

**The Chicken-and-Egg Problem**: On mainnet, validators need NAK to stake, but NAK only comes from BTC deposits. The bootstrap window solves this:

**Bootstrap Period**: First 2,016 blocks (~1 week at 12-second blocks)

During the bootstrap window:
- Single validator can produce blocks
- Allows initial BTC deposits to be processed
- NAK becomes available for staking
- New validators can join

After bootstrap (block 2,017+):
- Normal 67% consensus threshold applies
- Minimum validator requirements enforced
- Full decentralization achieved

**Implementation**:
```go
const BootstrapWindowBlocks = 2016

func (hvm *HierarchicalValidatorManager) IsInBootstrapWindow() bool {
    if hvm.networkMode.IsTestnet() {
        return false  // Testnet: no bootstrap window
    }
    return hvm.currentBlockHeight <= BootstrapWindowBlocks
}

func (hvm *HierarchicalValidatorManager) GetMinimumValidatorsForConsensus() int {
    if hvm.IsInBootstrapWindow() {
        return 1  // Single validator during bootstrap
    }
    return 4  // Normal minimum after bootstrap
}
```

### 13.6 Fork Proposal System (Mainnet Governance)

Since mainnet has no developer control, all network upgrades must go through the fork proposal system:

**Proposal Process**:
1. Anyone can submit a fork proposal
2. Validators vote during signaling windows (2,016 blocks)
3. 67% validator threshold required for activation
4. Lock-in period ensures network preparation
5. Grace period for node upgrades

**This provides founder influence through community consensus rather than unilateral control.**

### 13.7 Token Economics

**Testnet (tNAK)**:
- Zero real-world value
- Free tBTC available from faucets
- Unlimited experimentation

**Mainnet (NAK)**:
- Real value backed by BTC reserves
- 1:1 satoshi parity with Bitcoin
- Full proof-of-reserves validation

### 13.8 Implementation Files

| File | Purpose |
|------|---------|
| `internal/types/config.go` | NetworkMode type definition with TokenSymbol(), BitcoinNetwork(), ChainID(), NetworkID() methods |
| `internal/core/emergency_authority.go` | Phase handling based on network mode |
| `internal/core/genesis_block.go` | Testnet/mainnet genesis configurations |
| `internal/core/hierarchical_validator_system.go` | Bootstrap window logic |
| `internal/core/bitcoin_nakamoto_converter.go` | Network-aware Bitcoin selection |
| `internal/core/unified_blockchain_manager.go` | NetworkMode in UnifiedConfig |
| `cmd/nakamoto/main.go` | --network CLI flag |

### 13.9 CLI Usage

```bash
# Start testnet node (default)
./nakamoto --network testnet

# Start mainnet node
./nakamoto --network mainnet
```

### 13.10 API Response Format

All API responses include the network-appropriate token symbol:

```json
{
  "balance": 100000000,
  "currency": "NAK",        // or "tNAK" on testnet
  "network": "mainnet",     // or "testnet"
  "chainId": 1              // or 2 for testnet
}
```

### 13.11 Security Considerations

1. **Chain ID Separation**: Prevents cross-network replay attacks
2. **Different Token Symbols**: Clear visual distinction
3. **Mainnet PhaseMature**: Hardcoded at genesis, cannot be changed
4. **No Developer Key Shards**: Mainnet emergency authority has no developer keys

---

## 14. Lightning Network Integration (BTC↔NAK Conversions)

### 14.1 Overview

Nakamoto integrates with the Bitcoin Lightning Network to provide fast, low-fee BTC↔NAK conversions through community-hosted bridge nodes. This enables users to convert between Bitcoin and NAK tokens without requiring on-chain Bitcoin transactions for small amounts.

**Key Design Principles**:
- **Conversion-only**: Lightning handles BTC↔NAK swaps, not blockchain communication
- **No NAK Lightning**: NAK uses on-chain trunks (3-second blocks are fast enough)
- **Permissionless bridges**: Anyone can run a bridge node
- **Light clients**: Users don't need full nodes to convert
- **Atomic swaps**: HTLC-based trustless conversion

### 14.2 Architecture

```
User (Light Client) ──► Bridge Node (LND + Nakamoto) ──► Nakamoto Network
                              │
                              ▼
                    Bitcoin Lightning Network
```

**Bridge Node Components**:
| Component | Purpose |
|-----------|---------|
| LND gRPC Client | Real connection to Lightning node |
| Swap Manager | Tracks swap lifecycle and state |
| Invoice Manager | Creates and monitors Lightning invoices |
| NAK Blockchain Interface | Mints/burns NAK on-chain |

**Light Client Features**:
| Feature | Description |
|---------|-------------|
| Bridge Discovery | Find available bridge nodes |
| Fee Estimation | Get current conversion fees |
| Swap Initiation | Start BTC↔NAK conversions |
| Status Tracking | Monitor swap progress |

### 14.3 BTC → NAK Conversion Flow

```
1. User requests conversion via Light Client SDK
   └─► client.RequestBTCToNAK(ctx, 100000) // 100,000 sats

2. Bridge creates Lightning invoice
   └─► Returns BOLT11-encoded payment request

3. User pays invoice from any Lightning wallet
   └─► Standard Lightning payment

4. Bridge detects payment via invoice subscription
   └─► Invoice settled, preimage revealed

5. Bridge mints NAK on-chain to user's wallet
   └─► NAK appears in user's Nakamoto wallet
```

**Timeline**: ~30 seconds total (vs 3+ block confirmations for on-chain)

### 14.4 NAK → BTC Conversion Flow

```
1. User requests conversion with their Lightning invoice
   └─► client.RequestNAKToBTC(ctx, amount, lightningInvoice)

2. Bridge validates user has sufficient NAK balance
   └─► Checks wallet balance on Nakamoto chain

3. Bridge locks user's NAK in on-chain HTLC
   └─► NAK transferred to HTLC contract with timeout

4. Bridge pays user's Lightning invoice
   └─► User receives BTC on Lightning Network

5. Preimage reveals, bridge claims locked NAK
   └─► HTLC resolves, NAK transferred to bridge

6. Timeout fallback: NAK auto-refunds to user
   └─► If bridge fails to pay, user gets NAK back
```

### 14.5 Swap State Machine

```
SwapStatePending
    ↓
SwapStateInvoiceCreated (BTC→NAK) / SwapStateNAKLocked (NAK→BTC)
    ↓
SwapStateLightningPaid / SwapStateLightningPaying
    ↓
SwapStateNAKMinting / SwapStateCompleted
    ↓
SwapStateCompleted (success) / SwapStateRefunded (timeout)
```

**State Definitions**:
| State | Description |
|-------|-------------|
| Pending | Swap initiated, awaiting next step |
| InvoiceCreated | Lightning invoice generated (BTC→NAK) |
| NAKLocked | NAK locked in HTLC (NAK→BTC) |
| LightningPaid | Lightning payment confirmed |
| LightningPaying | Bridge paying Lightning invoice |
| NAKMinting | Minting NAK on-chain |
| Completed | Swap finished successfully |
| Refunded | Timeout triggered, funds returned |

### 14.6 Configuration

**SwapManagerConfig**:
```go
type SwapManagerConfig struct {
    DefaultExpirySec   int64  // Default: 3600 (1 hour)
    FeeRateBasisPoints int64  // Default: 10 (0.1%)
    MinFeeSat          int64  // Default: 100 sats
    MaxConcurrentSwaps int    // Default: 100
    HTLCTimeoutBlocks  uint64 // Default: 144 (~1 day)
}
```

**LightClientConfig**:
```go
type LightClientConfig struct {
    NAKWalletID           string        // User's wallet
    MaxFeeRate            int64         // Default: 50 (0.5%)
    MinReliability        int           // Default: 80 (80%)
    RequestTimeout        time.Duration // Default: 2 min
    BridgeRefreshInterval time.Duration // Default: 5 min
    AutoRetry             bool          // Default: true
    MaxRetries            int           // Default: 3
}
```

### 14.7 Fee Structure

| Parameter | Value | Description |
|-----------|-------|-------------|
| Fee Rate | 0.1% (10 basis points) | Percentage of swap amount |
| Minimum Fee | 100 sats | Floor for small swaps |
| Maximum Fee | Configurable | Cap for large swaps |

**Fee Calculation Example**:
```
Amount: 100,000 sats
Fee Rate: 0.1%
Calculated Fee: 100 sats (100,000 × 0.001)
Final Fee: 100 sats (meets minimum)
NAK Received: 99,900 NAK
```

### 14.8 Amount Limits

| Limit | Value | Rationale |
|-------|-------|-----------|
| Minimum | 10,000 sats | Prevents dust attacks, covers fees |
| Maximum | 1,000,000,000 sats (10 BTC) | Channel capacity limits |

### 14.9 Security Considerations

1. **HTLC Timeouts**: NAK HTLC timeout > Lightning timeout (prevents race conditions)
2. **Amount Limits**: Enforced min/max prevents abuse
3. **Rate Limiting**: Prevents spam attacks on bridges
4. **Preimage Persistence**: Stored before claiming (crash safety)
5. **TLS + Macaroon**: Required for LND authentication

### 14.10 REST API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/lightning/btc-to-nak` | Initiate BTC→NAK conversion |
| POST | `/api/lightning/nak-to-btc` | Initiate NAK→BTC conversion |
| GET | `/api/lightning/swap/{id}` | Get swap status |
| GET | `/api/lightning/swaps` | List user's swaps |
| GET | `/api/lightning/bridges` | List available bridge nodes |
| GET | `/api/lightning/bridge/status` | Current bridge node status |
| GET | `/api/lightning/fee-estimate` | Estimate fee for amount |

**Example: BTC→NAK Request**:
```json
POST /api/lightning/btc-to-nak
{
    "nak_wallet_id": "wallet-123",
    "amount_sats": 100000
}
```

**Example: BTC→NAK Response**:
```json
{
    "success": true,
    "swap_id": "swap_1234567890_abc123",
    "lightning_invoice": "lnbc1m1p...",
    "expires_at": 1705000000,
    "fee_sat": 100,
    "nak_amount": 99900,
    "message": "Pay the Lightning invoice to complete conversion"
}
```

### 14.11 Comparison with On-Chain Conversion

| Aspect | Lightning (This Section) | On-Chain (Section 3) |
|--------|--------------------------|-----------------------|
| Speed | ~30 seconds | 3+ block confirmations |
| Privacy | Better (off-chain) | On-chain visible |
| Capacity | Limited by channel balance | Unlimited |
| Complexity | Higher (requires LND) | Lower |
| Use case | Small, frequent conversions | Large conversions |

### 14.12 Implementation Files

| File | Purpose |
|------|---------|
| `internal/lightning/client.go` | High-level Lightning client |
| `internal/lightning/lnd_grpc_client.go` | Real LND gRPC connection |
| `internal/lightning/lnd_config.go` | LND configuration |
| `internal/lightning/lnd_invoice_manager.go` | Invoice management |
| `internal/lightning/lightning_swap.go` | Swap types and state machine |
| `internal/lightning/bridge_node.go` | Main bridge coordinator |
| `internal/lightning/light_client.go` | User SDK for conversions |
| `internal/api/lightning_routes.go` | REST API endpoints |

---

## 15. Content Voting System (Decentralized Moderation)

### 15.1 Overview

The Nakamoto network includes a decentralized content moderation system where users can vote on website content hosted on the network. This enables community-driven quality control without centralized censorship.

**Key Features**:
- Users with 100+ NAK can vote on content
- Each vote costs 0.1 NAK (paid via Lightning)
- Net scores below -100 trigger content suspension
- 30-day appeal period before potential deletion
- New content protected until 10 unique visits

### 15.2 Voting Requirements

| Requirement | Value | Rationale |
|-------------|-------|-----------|
| Minimum Stake | 100 NAK | Prevents Sybil attacks, ensures skin-in-game |
| Vote Cost | 0.1 NAK | Prevents spam voting, funds network |
| Stake Consumed | No | Stake is checked, not deducted |

### 15.3 Vote Types and Values

| Vote Type | Value | Effect |
|-----------|-------|--------|
| Upvote | +1 | Increases net score |
| Downvote | -1 | Decreases net score |
| None | 0 | No vote recorded |

**Net Score Calculation**:
```
Net Score = (Upvotes × +1) + (Downvotes × -1)
          = Upvotes - Downvotes
```

### 15.4 Content Status Lifecycle

```
                    ┌─────────────────────┐
                    │   StatusProtectedNew │ (< 10 visits)
                    └──────────┬──────────┘
                               │ 10+ unique visits
                               ▼
                    ┌─────────────────────┐
          ┌────────│     StatusActive     │◄────────┐
          │        └──────────┬──────────┘         │
          │                   │ Net score ≤ -100   │ Net score ≥ +50
          │                   ▼                    │
          │        ┌─────────────────────┐         │
          │        │   StatusSuspended   │─────────┘
          │        └──────────┬──────────┘
          │                   │ 30 days + score ≤ -50
          │                   ▼
          │        ┌─────────────────────┐
          │        │    StatusDeleted    │ (permanent)
          └────────┴─────────────────────┘
```

### 15.5 Content Status Definitions

| Status | Description | Visibility |
|--------|-------------|------------|
| ProtectedNew | New content with < 10 unique visits | Visible, immune to suspension |
| Active | Normal content | Visible |
| Suspended | Net score ≤ -100, under review | Hidden, appealable |
| Deleted | Permanently removed after review | Not accessible |

### 15.6 Thresholds and Timing

| Parameter | Value | Description |
|-----------|-------|-------------|
| Suspension Threshold | -100 | Net score triggers suspension |
| Reinstatement Threshold | +50 | Net score to restore from suspended |
| Deletion Threshold | -50 | Minimum score to avoid deletion |
| Review Period | 30 days | Time before deletion check |
| Minimum Visits | 10 | Visits required before suspension possible |
| Batch Interval | 60 seconds | Vote commits to blockchain |

### 15.7 New Content Protection

New content is protected from immediate suspension attacks:

1. Content starts with `StatusProtectedNew`
2. Each unique visitor is tracked
3. After 10 unique visits, status becomes `StatusActive`
4. Only then can negative scores trigger suspension

**Rationale**: Prevents competitors from immediately downvoting new content before legitimate users can discover it.

### 15.8 Vote Change Mechanism

Users can change their vote (upvote ↔ downvote):

1. User pays 0.1 NAK for new vote
2. Previous vote removed from totals
3. New vote added to totals
4. Vote history recorded for accountability

**Example**:
```
Initial: Content has 50 upvotes, 30 downvotes (Net: +20)
User changes from upvote to downvote:
Result: Content has 49 upvotes, 31 downvotes (Net: +18)
Net change: -2 (lost 1 upvote, gained 1 downvote)
```

### 15.9 Appeal Process

Content owners can appeal suspensions:

1. Owner submits appeal text (max 1,000 words)
2. Appeal is visible to all voters
3. Users can change votes based on appeal
4. If net score rises to +50, content reinstated
5. If 30 days pass with score ≤ -50, content deleted

### 15.10 Lightning Payment Integration

Votes require Lightning payment for anti-spam:

```
1. User requests vote invoice
   └─► GenerateVoteInvoice(contentHash, voterWallet, voteType)

2. System returns BOLT11 invoice
   └─► 0.1 NAK (10,000,000 msat), 5-minute expiry

3. User pays invoice
   └─► Payment from any Lightning wallet

4. System verifies payment
   └─► Preimage confirms payment

5. Vote recorded to blockchain
   └─► Batched every 60 seconds
```

### 15.11 Vote Batching

Votes are batched for efficiency:

| Parameter | Value |
|-----------|-------|
| Batch Interval | 60 seconds |
| Max Pending Votes | 10,000 |
| Max Concurrent Invoices | 1,000 |

**Batch Transaction Structure**:
```go
type VoteBatchTransaction struct {
    BatchID     string
    Votes       []*Vote
    Timestamp   int64
    BlockHeight uint64
}
```

### 15.12 Data Structures

**ContentRating**:
```go
type ContentRating struct {
    ContentHash    string            // SHA256 of content
    Domain         string            // e.g., "amazon.nakamoto"
    OwnerWallet    string            // Content owner's wallet
    Upvotes        int               // Total upvotes
    Downvotes      int               // Total downvotes
    NetScore       int               // Upvotes - Downvotes
    UniqueVisits   int               // Unique visitor count
    Status         ContentStatus     // Active/Suspended/Deleted
    SuspendedAt    int64             // Suspension timestamp
    ReviewDeadline int64             // Deletion check time
    AppealText     string            // Owner's appeal
    CreatedAt      int64
    LastUpdated    int64
}
```

**VoteRecord**:
```go
type VoteRecord struct {
    ContentHash     string     // Content being voted on
    VoterWallet     string     // Voter's wallet ID
    CurrentVote     VoteType   // "up", "down", or "none"
    VoteHistory     []VoteChange
    PaymentHash     []byte     // Lightning payment proof
    PaymentPreimage []byte     // Payment preimage
    Amount          uint64     // Vote cost paid
    FirstVotedAt    int64
    LastChangedAt   int64
}
```

### 15.13 Security Considerations

1. **Stake Requirement**: 100 NAK prevents Sybil attacks
2. **Vote Cost**: 0.1 NAK makes spam expensive
3. **New Content Protection**: 10-visit minimum prevents targeting
4. **Public Votes**: Accountability through transparency
5. **Appeal Period**: 30 days prevents hasty deletions
6. **Lightning Payments**: Cryptographic proof of payment

### 15.14 Privacy Considerations

| Data | Visibility |
|------|------------|
| Vote counts | Public |
| Voter wallet IDs | Public |
| Vote history | Public |
| Payment preimages | Public (proof of payment) |
| Appeal text | Public |

**Rationale**: Full transparency enables accountability and prevents vote manipulation.

### 15.15 Fee Distribution

Vote fees (0.1 NAK per vote) are distributed like regular transaction fees:
- Collected into block reward pool
- Distributed to validators and helpers per reward model
- No special handling for vote fees

### 15.16 Implementation Files

| File | Purpose |
|------|---------|
| `internal/content/types.go` | Data structures |
| `internal/content/voting_config.go` | Configuration constants |
| `internal/content/voting_system.go` | Core voting logic |
| `internal/lightning/voting_integration.go` | Lightning payment for votes |

### 15.17 Future Enhancements

- REST API endpoints for voting (in development)
- WebSocket real-time vote updates
- Vote delegation for DAOs
- Weighted voting based on stake level

---

## 16. Contributor Revenue Sharing

### 16.1 Overview

The Contributor Revenue Sharing system allows a configurable portion of the 5% developer fund to flow automatically to verified open-source contributors. Contracts are trustless and immutable once locked: parameters cannot be altered after `Lock()` is called. The system activates automatically on mainnet (ChainID 1) and is a no-op on testnet (ChainID 2).

### 16.2 Contract Lifecycle

```
SnapshotBPS → ContributorShareManager → Lock() → auto-activate on mainnet
```

1. **Snapshot**: Contribution scores are converted to basis-point shares via `SnapshotBPS()`.
2. **Lock**: `Lock()` freezes all parameters permanently. Any further call to add, remove, or update shares returns an error.
3. **Auto-Activate**: On the first block produced after `Lock()`, the manager checks `ChainID`. If `ChainID == 1` (mainnet), distribution begins immediately. No manual trigger is required.
4. **Expiry**: Contracts expire at `ActivationHeight + 2,592,000` blocks (≈ 12 months at 12-second block time). After expiry, distributions stop and a new contract must be created and locked for the next period.
5. **Renewal**: A renewed contract is a new `ContributorShareManager` instance; the prior immutable contract is preserved in history for audit.

### 16.3 Share Weights (Basis Points)

Shares are expressed in basis points where 10000 = 100%.

```
your_bps = (your_points / total_points) × 10000
```

The sum of all contributor BPS values must equal exactly 10000 before `Lock()` is accepted. The `SnapshotBPS()` function normalizes fractional remainders to the top contributor to guarantee this invariant.

### 16.4 Distribution Mechanics

Each block, the block reward's dev-fund portion (5% of collected fees) is split according to locked BPS weights:

```
contributor_payout = dev_fund_portion × (contributor_bps / 10000)
```

Payouts accumulate in per-contributor escrow balances and are claimable via the `/api/contributors/{wallet}/claim` endpoint. Unclaimed balances persist indefinitely.

### 16.5 WASM Escrow Contract

The escrow logic is implemented as an open-source WASM contract stored on-chain. This guarantees:

- Verifiability: anyone can inspect and recompile the contract bytecode.
- Immutability: the contract address is written into the locked `ContributorShareManager` and cannot be changed.
- Trustlessness: no administrator can override distribution rules after `Lock()`.

### 16.6 On-Chain Issue Tracker

The on-chain issue tracker is the source of truth for contribution scoring. Issues, votes, comments, and resolution events are stored as transactions in the main chain and indexed by `IssueTracker`.

#### 16.6.1 Issue Severity and Filing Points

| Severity | Filing Points | Upvote Bonus (each) |
|----------|--------------|---------------------|
| Critical | 50 | +5 |
| High     | 30 | +3 |
| Medium   | 15 | +2 |
| Low      |  5 | +1 |

Resolution points are awarded to the developer who closes the issue:

| Severity | Resolution Points |
|----------|-----------------|
| Critical | 200 |
| High     | 100 |
| Medium   |  40 |
| Low      |  10 |

Additional continuous points:
- **Node uptime**: 1 point per 24-hour uptime epoch (verified by peer attestation).

#### 16.6.2 BPS Formula

```
your_bps = (your_points / total_points) × 10000
```

`ContributionScorer` recalculates this formula on every snapshot. Points from expired contracts are not carried forward to new contracts.

#### 16.6.3 Anti-Gaming Rules

| Rule | Detail |
|------|--------|
| Confirmation required | Issues not confirmed by a second user within 14 days are auto-expired |
| Spam strikes | 3 spam marks within 30 days suspends filing rights for 90 days |
| Dev role earns 0 filing points | Developers earn resolution points only; they cannot file for points |
| Duplicate detection | `/api/v2/issues/search-similar` must be called before filing; confirmed duplicates earn 0 points |
| Upvote cap | Upvotes from accounts registered < 7 days before the issue was filed do not count toward bonus |

### 16.7 Implementation Files

| File | Purpose |
|------|---------|
| `internal/core/contributor_share_manager.go` | Lock(), auto-activate, BPS distribution, WASM escrow |
| `internal/core/issue_tracker_impl.go` | On-chain issue tracker: voting, search, dupe detection, status management |
| `internal/core/contribution_scorer.go` | Calculates BPS from issue activity, applies anti-gaming rules |

### 16.8 API Endpoints

See `internal/api/CLAUDE.md` for full endpoint documentation. Key paths:

- `GET/POST /api/contributors` — contributor pool management
- `GET /api/contributors/summary` — pool overview and stats
- `POST /api/contributors/{wallet}/claim` — claim accumulated earnings
- `GET /api/v2/issues` — browse and search issues
- `POST /api/v2/issues` — file a new issue (JWT required)
- `GET /api/v2/issues/scores` — public contribution scoreboard
- `PUT /api/v2/issues/{id}/resolve` — resolve issue (dev role required)

---

## Conclusion

Nakamoto v2.6.6 represents a comprehensive blockchain ecosystem combining hierarchical validation, ultra-conservative resource management, competitive quality maintenance, revolutionary decentralized storage, distributed caching, off-chain messaging with CRDTs, blockchain web hosting, blockchain-based update distribution, USB hardware-based two-factor authentication, complete testnet/mainnet separation, Lightning Network integration for fast BTC↔NAK conversions, decentralized content moderation through community voting, and trustless contributor revenue sharing with on-chain issue tracking. The system can scale from a single validator to millions of participants while maintaining efficiency, decentralization, and providing petabyte-scale storage with sub-millisecond cache access, conflict-free messaging, complete web hosting functionality, decentralized software updates, enhanced security through hardware authentication, network-mode-aware operation for both testing (tNAK) and production (NAK) environments, fast trustless BTC↔NAK conversions via Lightning, community-driven content quality control, and verifiable open-source contributor compensation.

**Document Version**: 2.6.6
**Last Updated**: January 2025
**Status**: Early Launch Phase - Storage System Complete, Update System Documented, USB 2FA System Implemented, Network Mode Separation Implemented, Lightning Network Integration Complete, Content Voting System Implemented, Contributor Revenue Sharing Documented, Cache & CRDT Systems In Development

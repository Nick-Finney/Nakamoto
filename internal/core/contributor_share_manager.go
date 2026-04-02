////////////////////////////////////////////////////////////////////////////////
// File: internal/core/contributor_share_manager.go
//
// Summary:
//   Trustless, immutable contributor revenue sharing system for Nakamoto blockchain.
//   Distributes a configurable percentage of the 5% dev fund to contributors
//   proportional to their share weights (basis points).
//
//   TRUSTLESS DESIGN — once a contract is locked, NOBODY can change it:
//   - Pool percentage, contributor list, and share weights become immutable
//   - Activation is automatic: mainnet detection (ChainID 1) or explicit block height
//   - Expiry is deterministic: activation_block + earning_duration_blocks
//   - Block heights are on-chain and tamper-proof (not timestamps)
//   - All state persisted to auditable JSON files
//
//   RENEWABLE CONTRACTS — annual earning periods:
//   - Each contract covers a fixed earning period (e.g., 12 months of blocks)
//   - When a contract expires, a NEW contract can be deployed for the next period
//   - Old contracts remain readable forever (audit trail)
//   - Contributors from expired contracts can still claim unclaimed earnings
//   - New contracts have independent contributor lists and share weights
//
//   BLOCK HEIGHT MATH (12-second block times):
//   - 1 hour   =    300 blocks
//   - 1 day    =  7,200 blocks
//   - 1 month  = 216,000 blocks (30 days)
//   - 12 months = 2,592,000 blocks
//
// How it works:
//   1. Admin creates a contract with contributors and share weights
//   2. Admin calls Lock() — contract becomes IMMUTABLE (no changes possible)
//   3. Contract auto-activates when mainnet is detected OR at a set block height
//   4. DistributeFees() is called each block — fees flow to contributors
//   5. After earning_duration_blocks, fees stop flowing (contract expired)
//   6. Contributors claim earnings at any time (even after expiry)
//   7. Admin can deploy a NEW contract for the next earning period
//
// Dependencies:
//   - developer_wallet_manager.go: Source of dev fund fees
//   - economic_distribution_manager.go: 5% dev fund from 4-way fee split
//
// Associated Test Files:
//   - internal/core/contributor_share_manager_test.go
////////////////////////////////////////////////////////////////////////////////

package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ===========================
// Constants
// ===========================

const (
	// BlocksPerMonth is the number of blocks in ~30 days at 12-second block times
	BlocksPerMonth uint64 = 216000

	// DefaultEarningDurationBlocks is 12 months of blocks (the default earning period)
	DefaultEarningDurationBlocks uint64 = 2592000

	// MainnetChainID is the chain ID that triggers automatic contract activation
	MainnetChainID = 1

	// TestnetChainID is the testnet chain ID
	TestnetChainID = 2
)

// ===========================
// Configuration
// ===========================

// ContributorShareConfig holds configuration for the contributor revenue sharing system
type ContributorShareConfig struct {
	// PoolPercentage is the percentage of dev fund fees routed to the contributor pool.
	// Range: 0.01-100.0. Example: 50.0 means 50% of dev fund goes to contributors.
	// IMMUTABLE once contract is locked.
	PoolPercentage float64 `json:"pool_percentage"`

	// StoragePath is the directory where contract state files are persisted
	StoragePath string `json:"storage_path"`

	// MinClaimAmount is the minimum NAK that can be claimed in a single transaction.
	// Prevents dust claims that waste blockchain space.
	MinClaimAmount float64 `json:"min_claim_amount"`

	// EarningDurationBlocks is how long the earning period lasts (in blocks).
	// At 12-second block times: 2,592,000 = ~12 months.
	// IMMUTABLE once contract is locked.
	EarningDurationBlocks uint64 `json:"earning_duration_blocks"`

	// ActivationBlockHeight is the block height at which earnings start accruing.
	// If 0, the contract auto-activates when mainnet is detected (ChainID 1).
	// IMMUTABLE once contract is locked.
	ActivationBlockHeight uint64 `json:"activation_block_height"`

	// AutoActivateOnMainnet enables automatic activation when mainnet is detected.
	// When true, ActivationBlockHeight is set to the first mainnet block seen.
	AutoActivateOnMainnet bool `json:"auto_activate_on_mainnet"`
}

// ===========================
// Core Types
// ===========================

// Contributor represents a project contributor who earns a share of dev fund revenue
type Contributor struct {
	// WalletID is the Nakamoto wallet address that receives earnings
	WalletID string `json:"wallet_id"`

	// Alias is a human-readable name (e.g., Discord handle, GitHub username)
	Alias string `json:"alias"`

	// ShareBPS is the contributor's share in basis points (10000 = 100% of pool).
	// Example: 3000 = 30% of the contributor pool.
	// IMMUTABLE once contract is locked.
	ShareBPS int `json:"share_bps"`

	// Contributions describes what this person contributed (for transparency)
	Contributions string `json:"contributions,omitempty"`

	// JoinedAt is when this contributor was added to the system
	JoinedAt time.Time `json:"joined_at"`

	// Active indicates whether this contributor is currently earning shares
	Active bool `json:"active"`
}

// ContributorEarnings tracks a contributor's accumulated and claimed earnings
type ContributorEarnings struct {
	WalletID         string    `json:"wallet_id"`
	TotalEarned      float64   `json:"total_earned"`
	TotalClaimed     float64   `json:"total_claimed"`
	UnclaimedBalance float64   `json:"unclaimed_balance"`
	LastEarnedAt     time.Time `json:"last_earned_at,omitempty"`
	LastClaimedAt    time.Time `json:"last_claimed_at,omitempty"`
}

// DistributionResult represents the result of a fee distribution round
type DistributionResult struct {
	TotalFeeInput     float64            `json:"total_fee_input"`
	TotalDistributed  float64            `json:"total_distributed"`
	RetainedByDevFund float64            `json:"retained_by_dev_fund"`
	PerContributor    map[string]float64 `json:"per_contributor"`
	BlockHeight       uint64             `json:"block_height"`
	Timestamp         time.Time          `json:"timestamp"`
}

// ClaimResult represents the result of a contributor claiming earnings
type ClaimResult struct {
	WalletID   string    `json:"wallet_id"`
	Amount     float64   `json:"amount"`
	Remaining  float64   `json:"remaining"`
	ContractID string    `json:"contract_id"`
	ClaimedAt  time.Time `json:"claimed_at"`
}

// DistributionEvent is an audit trail entry for distributions and claims
type DistributionEvent struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details"`
}

// ContributorShareSummary provides an overview of the contributor sharing system
type ContributorShareSummary struct {
	ContractID            string  `json:"contract_id"`
	TotalContributors     int     `json:"total_contributors"`
	PoolPercentage        float64 `json:"pool_percentage"`
	TotalAllocatedBPS     int     `json:"total_allocated_bps"`
	TotalDistributed      float64 `json:"total_distributed"`
	TotalClaimed          float64 `json:"total_claimed"`
	TotalUnclaimed        float64 `json:"total_unclaimed"`
	Locked                bool    `json:"locked"`
	Activated             bool    `json:"activated"`
	Expired               bool    `json:"expired"`
	ActivationBlockHeight uint64  `json:"activation_block_height"`
	ExpiryBlockHeight     uint64  `json:"expiry_block_height"`
	EarningDurationBlocks uint64  `json:"earning_duration_blocks"`
	CurrentBlockHeight    uint64  `json:"current_block_height,omitempty"`
}

// ContributorShareState is the full persisted state (for JSON serialization).
// This file IS the contract — it is the single source of truth, auditable by anyone.
type ContributorShareState struct {
	// Contract identity
	ContractID string `json:"contract_id"`
	Version    int    `json:"version"` // Increments with each new contract (1, 2, 3...)

	// Immutable parameters (cannot change after Locked=true)
	PoolPercentage        float64        `json:"pool_percentage"`
	EarningDurationBlocks uint64         `json:"earning_duration_blocks"`
	Contributors          []*Contributor `json:"contributors"`

	// Activation state
	Locked                bool   `json:"locked"`                  // Once true, NOTHING can change
	ActivationBlockHeight uint64 `json:"activation_block_height"` // 0 = not yet activated
	ExpiryBlockHeight     uint64 `json:"expiry_block_height"`     // 0 = not yet calculated
	AutoActivateOnMainnet bool   `json:"auto_activate_on_mainnet"`

	// Mutable state (earnings accrue and claims happen)
	MinClaimAmount   float64                         `json:"min_claim_amount"`
	Earnings         map[string]*ContributorEarnings `json:"earnings"`
	History          []*DistributionEvent            `json:"history"`
	TotalDistributed float64                         `json:"total_distributed"`
	TotalClaimed     float64                         `json:"total_claimed"`

	// Metadata
	CreatedAt   time.Time `json:"created_at"`
	LockedAt    time.Time `json:"locked_at,omitempty"`
	ActivatedAt time.Time `json:"activated_at,omitempty"`
	LastUpdated time.Time `json:"last_updated"`
}

// ===========================
// Manager
// ===========================

// ContributorShareManager manages the trustless contributor revenue sharing system.
// Each instance represents ONE contract (one earning period). To renew, create a
// new manager with a new contract ID.
//
// Lifecycle:
//   1. NewContributorShareManager() — create contract
//   2. AddContributor() — add contributors (only before lock)
//   3. Lock() — PERMANENTLY freeze all parameters
//   4. DistributeFeesAtBlock() — called each block, auto-activates on mainnet
//   5. ClaimEarnings() — contributors withdraw their earnings
//   6. IsExpired() — check if earning period is over
//   7. Deploy new contract for next period
type ContributorShareManager struct {
	config       *ContributorShareConfig
	contractID   string
	version      int
	contributors map[string]*Contributor
	earnings     map[string]*ContributorEarnings
	history      []*DistributionEvent
	logger       *zap.Logger
	mutex        sync.RWMutex

	// Immutability
	locked    bool
	lockedAt  time.Time
	createdAt time.Time

	// Activation state
	activationBlockHeight uint64
	expiryBlockHeight     uint64
	activated             bool

	// Aggregate tracking
	totalDistributed float64
	totalClaimed     float64
}

// NewContributorShareManager creates a new contributor share manager (a new contract).
// Loads existing state from disk if a contract with this ID exists.
func NewContributorShareManager(config *ContributorShareConfig, logger *zap.Logger) (*ContributorShareManager, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required")
	}

	if config.PoolPercentage <= 0 || config.PoolPercentage > 100 {
		return nil, fmt.Errorf("pool_percentage must be between 0 (exclusive) and 100 (inclusive), got %f", config.PoolPercentage)
	}

	if config.StoragePath == "" {
		return nil, fmt.Errorf("storage_path is required")
	}

	if config.EarningDurationBlocks == 0 {
		config.EarningDurationBlocks = DefaultEarningDurationBlocks
	}

	if err := os.MkdirAll(config.StoragePath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	mgr := &ContributorShareManager{
		config:       config,
		contributors: make(map[string]*Contributor),
		earnings:     make(map[string]*ContributorEarnings),
		history:      make([]*DistributionEvent, 0),
		logger:       logger,
		createdAt:    time.Now(),
	}

	// Try to load existing state from disk
	if err := mgr.load(); err != nil {
		if !os.IsNotExist(err) {
			logger.Warn("failed to load contributor shares state, starting fresh",
				zap.Error(err))
		}
	}

	return mgr, nil
}

// ===========================
// Immutability Controls
// ===========================

// Lock permanently freezes the contract. After this call:
//   - No contributors can be added, removed, or modified
//   - Pool percentage cannot change
//   - Earning duration cannot change
//   - The ONLY mutable state is earnings accrual and claims
//
// This is the core trustless guarantee: once locked, code enforces everything.
func (csm *ContributorShareManager) Lock() error {
	csm.mutex.Lock()
	defer csm.mutex.Unlock()

	if csm.locked {
		return fmt.Errorf("contract is already locked")
	}

	if len(csm.contributors) == 0 {
		return fmt.Errorf("cannot lock contract with no contributors")
	}

	csm.locked = true
	csm.lockedAt = time.Now()

	// If an explicit activation height is set, calculate expiry now
	if csm.config.ActivationBlockHeight > 0 {
		csm.activationBlockHeight = csm.config.ActivationBlockHeight
		csm.expiryBlockHeight = csm.activationBlockHeight + csm.config.EarningDurationBlocks
		csm.activated = true
	}

	csm.history = append(csm.history, &DistributionEvent{
		Type:      "contract_locked",
		Timestamp: csm.lockedAt,
		Details: map[string]interface{}{
			"pool_percentage":         csm.config.PoolPercentage,
			"earning_duration_blocks": csm.config.EarningDurationBlocks,
			"contributor_count":       len(csm.contributors),
			"total_allocated_bps":     csm.totalAllocatedBPSUnsafe(),
			"activation_block_height": csm.activationBlockHeight,
			"auto_activate_mainnet":   csm.config.AutoActivateOnMainnet,
		},
	})

	csm.logger.Info("CONTRACT LOCKED — parameters are now immutable",
		zap.Float64("pool_percentage", csm.config.PoolPercentage),
		zap.Uint64("earning_duration_blocks", csm.config.EarningDurationBlocks),
		zap.Int("contributors", len(csm.contributors)),
		zap.Bool("auto_activate_mainnet", csm.config.AutoActivateOnMainnet),
	)

	// Persist immediately — the locked state must survive restarts
	return csm.saveUnsafe()
}

// IsLocked returns whether the contract is locked (immutable)
func (csm *ContributorShareManager) IsLocked() bool {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()
	return csm.locked
}

// IsActivated returns whether the contract is currently active (earning period started)
func (csm *ContributorShareManager) IsActivated() bool {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()
	return csm.activated
}

// IsExpired returns whether the earning period has ended
func (csm *ContributorShareManager) IsExpired(currentBlockHeight uint64) bool {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()
	return csm.isExpiredUnsafe(currentBlockHeight)
}

func (csm *ContributorShareManager) isExpiredUnsafe(currentBlockHeight uint64) bool {
	if !csm.activated || csm.expiryBlockHeight == 0 {
		return false
	}
	return currentBlockHeight >= csm.expiryBlockHeight
}

// IsEarning returns true if the contract is locked, activated, and not yet expired.
// This is the window during which fees are distributed to contributors.
func (csm *ContributorShareManager) IsEarning(currentBlockHeight uint64) bool {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()
	return csm.locked && csm.activated && !csm.isExpiredUnsafe(currentBlockHeight)
}

// ===========================
// Contributor Management
// ===========================

// AddContributor adds a new contributor to the revenue sharing pool.
// Returns error if contract is locked, wallet_id already exists, or total BPS exceeds 10000.
func (csm *ContributorShareManager) AddContributor(c *Contributor) error {
	csm.mutex.Lock()
	defer csm.mutex.Unlock()

	// IMMUTABILITY CHECK — the core trustless guarantee
	if csm.locked {
		return fmt.Errorf("contract is locked — contributors cannot be modified")
	}

	if c == nil {
		return fmt.Errorf("contributor cannot be nil")
	}
	if c.WalletID == "" {
		return fmt.Errorf("wallet_id is required")
	}
	if c.ShareBPS <= 0 || c.ShareBPS > 10000 {
		return fmt.Errorf("share_bps must be between 1 and 10000, got %d", c.ShareBPS)
	}

	if _, exists := csm.contributors[c.WalletID]; exists {
		return fmt.Errorf("contributor with wallet_id %q already exists", c.WalletID)
	}

	totalBPS := csm.totalAllocatedBPSUnsafe()
	if totalBPS+c.ShareBPS > 10000 {
		return fmt.Errorf("adding %d BPS would exceed 10000 total (currently %d allocated)", c.ShareBPS, totalBPS)
	}

	c.JoinedAt = time.Now()
	c.Active = true

	csm.contributors[c.WalletID] = c
	csm.earnings[c.WalletID] = &ContributorEarnings{
		WalletID: c.WalletID,
	}

	csm.history = append(csm.history, &DistributionEvent{
		Type:      "add_contributor",
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"wallet_id": c.WalletID,
			"alias":     c.Alias,
			"share_bps": c.ShareBPS,
		},
	})

	csm.logger.Info("contributor added",
		zap.String("wallet_id", c.WalletID),
		zap.String("alias", c.Alias),
		zap.Int("share_bps", c.ShareBPS),
	)

	return nil
}

// RemoveContributor removes a contributor from the revenue sharing pool.
// Returns error if contract is locked. Unclaimed earnings remain available.
func (csm *ContributorShareManager) RemoveContributor(walletID string) error {
	csm.mutex.Lock()
	defer csm.mutex.Unlock()

	if csm.locked {
		return fmt.Errorf("contract is locked — contributors cannot be modified")
	}

	c, exists := csm.contributors[walletID]
	if !exists {
		return fmt.Errorf("contributor with wallet_id %q not found", walletID)
	}

	c.Active = false
	delete(csm.contributors, walletID)

	csm.history = append(csm.history, &DistributionEvent{
		Type:      "remove_contributor",
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"wallet_id": walletID,
			"alias":     c.Alias,
		},
	})

	csm.logger.Info("contributor removed",
		zap.String("wallet_id", walletID),
		zap.String("alias", c.Alias),
	)

	return nil
}

// UpdateContributorShare updates a contributor's share weight.
// Returns error if contract is locked.
func (csm *ContributorShareManager) UpdateContributorShare(walletID string, newShareBPS int) error {
	csm.mutex.Lock()
	defer csm.mutex.Unlock()

	if csm.locked {
		return fmt.Errorf("contract is locked — share weights cannot be modified")
	}

	c, exists := csm.contributors[walletID]
	if !exists {
		return fmt.Errorf("contributor with wallet_id %q not found", walletID)
	}

	if newShareBPS <= 0 || newShareBPS > 10000 {
		return fmt.Errorf("share_bps must be between 1 and 10000, got %d", newShareBPS)
	}

	totalBPS := csm.totalAllocatedBPSUnsafe() - c.ShareBPS + newShareBPS
	if totalBPS > 10000 {
		return fmt.Errorf("updating to %d BPS would exceed 10000 total", newShareBPS)
	}

	oldBPS := c.ShareBPS
	c.ShareBPS = newShareBPS

	csm.history = append(csm.history, &DistributionEvent{
		Type:      "update_share",
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"wallet_id": walletID,
			"old_bps":   oldBPS,
			"new_bps":   newShareBPS,
		},
	})

	return nil
}

// ListContributors returns all active contributors
func (csm *ContributorShareManager) ListContributors() []*Contributor {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()

	result := make([]*Contributor, 0, len(csm.contributors))
	for _, c := range csm.contributors {
		if c.Active {
			result = append(result, c)
		}
	}
	return result
}

// ===========================
// Mainnet Auto-Activation
// ===========================

// NotifyBlockHeight is called each block to check for auto-activation.
// If the contract is locked with AutoActivateOnMainnet=true and the chain ID
// indicates mainnet, the contract activates at this block height.
// This is the trustless trigger — no human action required.
func (csm *ContributorShareManager) NotifyBlockHeight(blockHeight uint64, chainID int) {
	csm.mutex.Lock()
	defer csm.mutex.Unlock()

	if !csm.locked || csm.activated {
		return
	}

	// Auto-activate on mainnet detection
	if csm.config.AutoActivateOnMainnet && chainID == MainnetChainID {
		csm.activationBlockHeight = blockHeight
		csm.expiryBlockHeight = blockHeight + csm.config.EarningDurationBlocks
		csm.activated = true

		csm.history = append(csm.history, &DistributionEvent{
			Type:      "contract_activated",
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"trigger":              "mainnet_detected",
				"chain_id":             chainID,
				"activation_block":     blockHeight,
				"expiry_block":         csm.expiryBlockHeight,
				"earning_duration":     csm.config.EarningDurationBlocks,
				"earning_months_approx": float64(csm.config.EarningDurationBlocks) / float64(BlocksPerMonth),
			},
		})

		csm.logger.Info("CONTRACT ACTIVATED — mainnet detected, earnings now accruing",
			zap.Uint64("activation_block", blockHeight),
			zap.Uint64("expiry_block", csm.expiryBlockHeight),
			zap.Uint64("duration_blocks", csm.config.EarningDurationBlocks),
		)

		// Persist activation state immediately
		_ = csm.saveUnsafe()
	}
}

// ===========================
// Fee Distribution
// ===========================

// DistributeFees distributes dev fund fees to the contributor pool.
// This is the backwards-compatible version (no block height). Calls
// DistributeFeesAtBlock with block height 0.
func (csm *ContributorShareManager) DistributeFees(devFundFee float64) (*DistributionResult, error) {
	return csm.DistributeFeesAtBlock(devFundFee, 0)
}

// DistributeFeesAtBlock distributes dev fund fees to the contributor pool,
// checking the current block height against the activation/expiry window.
//
// Rules enforced by code (not trust):
//   - If contract is not locked: no distribution (setup phase)
//   - If contract is not activated: no distribution (waiting for mainnet)
//   - If contract is expired: no distribution (earning period over)
//   - If all checks pass: fees are distributed proportionally by BPS
func (csm *ContributorShareManager) DistributeFeesAtBlock(devFundFee float64, blockHeight uint64) (*DistributionResult, error) {
	csm.mutex.Lock()
	defer csm.mutex.Unlock()

	result := &DistributionResult{
		TotalFeeInput:  devFundFee,
		PerContributor: make(map[string]float64),
		BlockHeight:    blockHeight,
		Timestamp:      time.Now(),
	}

	// Rule 1: Contract must be locked
	if !csm.locked {
		result.RetainedByDevFund = devFundFee
		return result, nil
	}

	// Rule 2: Contract must be activated
	if !csm.activated {
		result.RetainedByDevFund = devFundFee
		return result, nil
	}

	// Rule 3: Contract must not be expired (block height check)
	if blockHeight > 0 && csm.isExpiredUnsafe(blockHeight) {
		result.RetainedByDevFund = devFundFee
		return result, nil
	}

	// Rule 4: Must have fees and contributors
	if devFundFee <= 0 || len(csm.contributors) == 0 {
		result.RetainedByDevFund = devFundFee
		return result, nil
	}

	// Calculate pool amount
	poolAmount := devFundFee * (csm.config.PoolPercentage / 100.0)
	totalAllocatedBPS := csm.totalAllocatedBPSUnsafe()

	// Distribute to each active contributor
	totalDistributed := 0.0
	for walletID, c := range csm.contributors {
		if !c.Active {
			continue
		}

		share := poolAmount * (float64(c.ShareBPS) / 10000.0)
		if share <= 0 {
			continue
		}

		earnings, exists := csm.earnings[walletID]
		if !exists {
			earnings = &ContributorEarnings{WalletID: walletID}
			csm.earnings[walletID] = earnings
		}
		earnings.TotalEarned += share
		earnings.UnclaimedBalance += share
		earnings.LastEarnedAt = result.Timestamp

		result.PerContributor[walletID] = share
		totalDistributed += share
	}

	result.TotalDistributed = totalDistributed
	unallocatedPoolBPS := 10000 - totalAllocatedBPS
	unallocatedPool := poolAmount * (float64(unallocatedPoolBPS) / 10000.0)
	result.RetainedByDevFund = (devFundFee - poolAmount) + unallocatedPool

	csm.totalDistributed += totalDistributed

	csm.history = append(csm.history, &DistributionEvent{
		Type:      "distribution",
		Timestamp: result.Timestamp,
		Details: map[string]interface{}{
			"fee_input":         devFundFee,
			"pool_amount":       poolAmount,
			"total_distributed": totalDistributed,
			"retained":          result.RetainedByDevFund,
			"block_height":      blockHeight,
			"contributor_count": len(result.PerContributor),
		},
	})

	return result, nil
}

// ===========================
// Claiming
// ===========================

// ClaimEarnings allows a contributor to claim a specified amount of their unclaimed earnings.
// Claims are allowed even after the contract expires — earned NAK never disappears.
// Claims require the contract to be locked and activated (mainnet must be live).
func (csm *ContributorShareManager) ClaimEarnings(walletID string, amount float64) (*ClaimResult, error) {
	csm.mutex.Lock()
	defer csm.mutex.Unlock()

	// Contract must be locked and activated before claims are possible
	if !csm.locked {
		return nil, fmt.Errorf("contract is not locked — claims not available during setup")
	}
	if !csm.activated {
		return nil, fmt.Errorf("contract is not yet activated — earnings accrue after mainnet launch")
	}

	earnings, exists := csm.earnings[walletID]
	if !exists {
		return nil, fmt.Errorf("no earnings found for wallet_id %q", walletID)
	}

	if amount < csm.config.MinClaimAmount {
		return nil, fmt.Errorf("claim amount %.8f is below minimum %.8f NAK", amount, csm.config.MinClaimAmount)
	}

	if amount > earnings.UnclaimedBalance {
		return nil, fmt.Errorf("claim amount %.8f exceeds unclaimed balance %.8f NAK", amount, earnings.UnclaimedBalance)
	}

	earnings.UnclaimedBalance -= amount
	earnings.TotalClaimed += amount
	earnings.LastClaimedAt = time.Now()

	csm.totalClaimed += amount

	result := &ClaimResult{
		WalletID:   walletID,
		Amount:     amount,
		Remaining:  earnings.UnclaimedBalance,
		ContractID: csm.contractID,
		ClaimedAt:  earnings.LastClaimedAt,
	}

	csm.history = append(csm.history, &DistributionEvent{
		Type:      "claim",
		Timestamp: result.ClaimedAt,
		Details: map[string]interface{}{
			"wallet_id": walletID,
			"amount":    amount,
			"remaining": earnings.UnclaimedBalance,
		},
	})

	csm.logger.Info("contributor claimed earnings",
		zap.String("wallet_id", walletID),
		zap.Float64("amount", amount),
		zap.Float64("remaining", earnings.UnclaimedBalance),
	)

	return result, nil
}

// ===========================
// Queries
// ===========================

// GetContributorEarnings returns the earnings summary for a specific contributor.
// Returns zero-value earnings for nonexistent contributors.
func (csm *ContributorShareManager) GetContributorEarnings(walletID string) ContributorEarnings {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()

	if earnings, exists := csm.earnings[walletID]; exists {
		return *earnings
	}
	return ContributorEarnings{WalletID: walletID}
}

// GetDistributionHistory returns the last N distribution events
func (csm *ContributorShareManager) GetDistributionHistory(limit int) []*DistributionEvent {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()

	if limit <= 0 || limit > len(csm.history) {
		limit = len(csm.history)
	}

	start := len(csm.history) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*DistributionEvent, limit)
	copy(result, csm.history[start:])
	return result
}

// GetSummary returns an overview of the contributor sharing system
func (csm *ContributorShareManager) GetSummary() ContributorShareSummary {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()

	totalUnclaimed := 0.0
	for _, e := range csm.earnings {
		totalUnclaimed += e.UnclaimedBalance
	}

	return ContributorShareSummary{
		ContractID:            csm.contractID,
		TotalContributors:     len(csm.contributors),
		PoolPercentage:        csm.config.PoolPercentage,
		TotalAllocatedBPS:     csm.totalAllocatedBPSUnsafe(),
		TotalDistributed:      csm.totalDistributed,
		TotalClaimed:          csm.totalClaimed,
		TotalUnclaimed:        totalUnclaimed,
		Locked:                csm.locked,
		Activated:             csm.activated,
		Expired:               csm.expiryBlockHeight > 0 && csm.activated,
		ActivationBlockHeight: csm.activationBlockHeight,
		ExpiryBlockHeight:     csm.expiryBlockHeight,
		EarningDurationBlocks: csm.config.EarningDurationBlocks,
	}
}

// GetContractID returns the contract identifier
func (csm *ContributorShareManager) GetContractID() string {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()
	return csm.contractID
}

// SetContractID sets the contract identifier (only before lock)
func (csm *ContributorShareManager) SetContractID(id string) error {
	csm.mutex.Lock()
	defer csm.mutex.Unlock()
	if csm.locked {
		return fmt.Errorf("contract is locked — ID cannot be changed")
	}
	csm.contractID = id
	return nil
}

// GetActivationBlockHeight returns the block at which earning started
func (csm *ContributorShareManager) GetActivationBlockHeight() uint64 {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()
	return csm.activationBlockHeight
}

// GetExpiryBlockHeight returns the block at which earning stops
func (csm *ContributorShareManager) GetExpiryBlockHeight() uint64 {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()
	return csm.expiryBlockHeight
}

// ===========================
// Persistence
// ===========================

// Save persists the current state to disk
func (csm *ContributorShareManager) Save() error {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()
	return csm.saveUnsafe()
}

// saveUnsafe persists state without acquiring lock (caller must hold lock)
func (csm *ContributorShareManager) saveUnsafe() error {
	contributors := make([]*Contributor, 0, len(csm.contributors))
	for _, c := range csm.contributors {
		contributors = append(contributors, c)
	}

	state := &ContributorShareState{
		ContractID:            csm.contractID,
		Version:               csm.version,
		PoolPercentage:        csm.config.PoolPercentage,
		EarningDurationBlocks: csm.config.EarningDurationBlocks,
		Contributors:          contributors,
		Locked:                csm.locked,
		ActivationBlockHeight: csm.activationBlockHeight,
		ExpiryBlockHeight:     csm.expiryBlockHeight,
		AutoActivateOnMainnet: csm.config.AutoActivateOnMainnet,
		MinClaimAmount:        csm.config.MinClaimAmount,
		Earnings:              csm.earnings,
		History:               csm.history,
		TotalDistributed:      csm.totalDistributed,
		TotalClaimed:          csm.totalClaimed,
		CreatedAt:             csm.createdAt,
		LockedAt:              csm.lockedAt,
		LastUpdated:           time.Now(),
	}
	if csm.activated {
		state.ActivatedAt = csm.history[len(csm.history)-1].Timestamp
		for _, h := range csm.history {
			if h.Type == "contract_activated" {
				state.ActivatedAt = h.Timestamp
				break
			}
		}
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	path := filepath.Join(csm.config.StoragePath, "contributor_shares.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	csm.logger.Debug("contributor shares state saved", zap.String("path", path))
	return nil
}

// load reads persisted state from disk
func (csm *ContributorShareManager) load() error {
	path := filepath.Join(csm.config.StoragePath, "contributor_shares.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var state ContributorShareState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Restore contributors
	for _, c := range state.Contributors {
		csm.contributors[c.WalletID] = c
	}

	// Restore earnings
	for walletID, e := range state.Earnings {
		csm.earnings[walletID] = e
	}

	// Restore contract identity and state
	csm.contractID = state.ContractID
	csm.version = state.Version
	csm.locked = state.Locked
	csm.lockedAt = state.LockedAt
	csm.createdAt = state.CreatedAt
	csm.activationBlockHeight = state.ActivationBlockHeight
	csm.expiryBlockHeight = state.ExpiryBlockHeight
	csm.activated = state.ActivationBlockHeight > 0
	csm.history = state.History
	csm.totalDistributed = state.TotalDistributed
	csm.totalClaimed = state.TotalClaimed

	csm.logger.Info("contributor shares contract loaded",
		zap.String("contract_id", csm.contractID),
		zap.Bool("locked", csm.locked),
		zap.Bool("activated", csm.activated),
		zap.Int("contributors", len(csm.contributors)),
		zap.Float64("total_distributed", csm.totalDistributed),
	)

	return nil
}

// Close saves state and releases resources
func (csm *ContributorShareManager) Close() {
	_ = csm.Save()
}

// ===========================
// Internal Helpers
// ===========================

// totalAllocatedBPSUnsafe returns total BPS allocated across all active contributors.
// Must be called with at least a read lock held.
func (csm *ContributorShareManager) totalAllocatedBPSUnsafe() int {
	total := 0
	for _, c := range csm.contributors {
		if c.Active {
			total += c.ShareBPS
		}
	}
	return total
}

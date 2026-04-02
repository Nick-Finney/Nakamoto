////////////////////////////////////////////////////////////////////////////////
// File: internal/core/issue_tracker_impl.go
//
// Summary:
//   Core community issue tracker for the Nakamoto blockchain.
//   Users file bug reports / feature requests, vote on them, confirm reproduction,
//   and developers triage/resolve them. Spam detection via UserRegistry strike system.
//
//   DESIGN PRINCIPLES:
//   - Decentralized moderation: spam detection via strike system (UserRegistry)
//   - Self-confirming: community confirms issues, not just the reporter
//   - Duplicate detection: two independent flags OR one dev flag auto-merges votes
//   - Auto-expiry: unconfirmed open issues expire after AutoExpireBlocks (14 days)
//   - Persistence: JSON-backed, same pattern as contributor_share_manager.go
//
// Roles:
//   - contributor (default): can file, vote, comment, confirm, flag duplicates
//   - dev: can triage (set status), resolve, close-as-invalid, mark-as-spam
//
// Vote rules:
//   - One vote per wallet per issue (can switch direction)
//   - Creator cannot vote on their own issue (conflict of interest)
//
// Duplicate merging:
//   - When duplicate is confirmed, votes from the dupe are merged into the target
//   - Votes already cast on the target are not overridden
//
// Dependencies:
//   - user_registry.go: Spam strike tracking and role checks
//
// Associated Test Files:
//   - internal/core/issue_tracker_test.go
////////////////////////////////////////////////////////////////////////////////

package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ===========================
// Constants
// ===========================

// trackerValidSeverities lists the allowed issue severity levels.
// Ordered from most to least critical.
var trackerValidSeverities = map[string]bool{
	"critical":    true,
	"major":       true,
	"minor":       true,
	"enhancement": true,
}

// trackerValidStatuses lists the allowed issue status transitions.
// Lifecycle: open → confirmed/in-progress/cannot-reproduce → resolved/closed/invalid/spam
var trackerValidStatuses = map[string]bool{
	"open":             true,
	"confirmed":        true,
	"in-progress":      true,
	"resolved":         true,
	"closed":           true,
	"cannot-reproduce": true,
	"invalid":          true,
	"spam":             true,
}

// trackerValidNetworks lists the allowed network scopes for issues.
var trackerValidNetworks = map[string]bool{
	"testnet": true,
	"mainnet": true,
	"both":    true,
}

// trackerDupeFlagThreshold is the number of community duplicate flags required to
// auto-confirm a duplicate without a dev decision. Two independent users
// flagging the same target provides enough signal to merge.
const trackerDupeFlagThreshold = 2

// ===========================
// Configuration
// ===========================

// IssueTrackerConfig holds configuration for the issue tracker.
type IssueTrackerConfig struct {
	// StoragePath is the directory where issues.json is persisted.
	StoragePath string `json:"storage_path"`

	// AutoExpireBlocks is how many blocks an open issue can survive without
	// at least one confirmation before being auto-closed as "cannot-reproduce".
	// Default: 100800 blocks (~14 days at 12-second block times).
	AutoExpireBlocks uint64 `json:"auto_expire_blocks"`

	// MaxSpamStrikes is the number of spam strikes before a user is blocked
	// from filing new issues. Default: 3.
	MaxSpamStrikes int `json:"max_spam_strikes"`
}

// ===========================
// Core Types
// ===========================

// TrackerIssue represents a bug report, feature request, or enhancement proposal.
// Issues are community-driven: anyone can file, vote, and confirm reproduction.
//
// Note: named TrackerIssue to coexist with the legacy bounty-based Issue type
// in issue_tracker.go. The IssueTracker API surfaces this type.
type TrackerIssue struct {
	// IssueID is the unique identifier, auto-assigned as "ISS-1", "ISS-2", etc.
	IssueID string `json:"issue_id"`

	// ProjectID scopes the issue to a project namespace (e.g., "nakamoto").
	ProjectID string `json:"project_id"`

	// Title is the short summary (required).
	Title string `json:"title"`

	// Body is the detailed description in Markdown format.
	Body string `json:"body"`

	// Severity indicates impact: "critical", "major", "minor", "enhancement".
	// Critical = data loss or security vulnerability.
	// Major = significant functionality broken.
	// Minor = small bugs, cosmetic issues.
	// Enhancement = new feature requests.
	Severity string `json:"severity"`

	// Status tracks the issue lifecycle. Starts at "open".
	// Only dev-role users can change status.
	Status string `json:"status"`

	// Network indicates which network is affected: "testnet", "mainnet", "both".
	Network string `json:"network"`

	// Tags are optional freeform labels for filtering (e.g., "p2p", "wallet").
	Tags []string `json:"tags,omitempty"`

	// CreatorWallet is the wallet address of the issue filer.
	// Used for self-vote prevention.
	CreatorWallet string `json:"creator_wallet"`

	// CreatorUsername is the display name of the filer.
	CreatorUsername string `json:"creator_username"`

	// ResolverWallet is the dev wallet that marked this issue resolved.
	ResolverWallet string `json:"resolver_wallet,omitempty"`

	// ResolutionNote explains how the issue was fixed.
	ResolutionNote string `json:"resolution_note,omitempty"`

	// InvalidReason is required when closing as "invalid" to explain why.
	InvalidReason string `json:"invalid_reason,omitempty"`

	// Score is the net vote count (upvotes - downvotes).
	// Higher score = more community support.
	Score int `json:"score"`

	// Upvotes is the total number of upvotes received.
	Upvotes int `json:"upvotes"`

	// Downvotes is the total number of downvotes received.
	Downvotes int `json:"downvotes"`

	// VoterWallets maps each voter's wallet to their vote direction: "up" or "down".
	// Prevents double-voting and allows vote switching.
	VoterWallets map[string]string `json:"voter_wallets"`

	// DuplicateOf is the IssueID that this issue is considered a duplicate of,
	// once DupeConfirmed becomes true.
	DuplicateOf string `json:"duplicate_of,omitempty"`

	// DupeFlags maps flagger wallet to the target IssueID they flagged.
	// When 2+ wallets flag the same target, or a dev flags it, DupeConfirmed is set.
	DupeFlags map[string]string `json:"dupe_flags"`

	// DupeConfirmed indicates that this issue has been officially confirmed as a
	// duplicate, and its votes have been merged into the target issue.
	DupeConfirmed bool `json:"dupe_confirmed"`

	// Confirmations is the set of wallet addresses that have confirmed they can
	// reproduce the issue. Issues with 0 confirmations expire after AutoExpireBlocks.
	Confirmations map[string]bool `json:"confirmations"`

	// CreatedAt is the wall-clock timestamp when the issue was filed.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the wall-clock timestamp of the last modification.
	UpdatedAt time.Time `json:"updated_at"`

	// CreatedAtBlock is the blockchain height when the issue was filed.
	// Used for auto-expiry calculations.
	CreatedAtBlock uint64 `json:"created_at_block"`

	// ResolvedAtBlock is the blockchain height when the issue was resolved.
	ResolvedAtBlock uint64 `json:"resolved_at_block,omitempty"`
}

// TrackerComment is a reply attached to a TrackerIssue.
// Comments are append-only (no editing/deletion).
//
// Note: named TrackerComment to coexist with the legacy IssueComment type.
type TrackerComment struct {
	// CommentID is the unique identifier, auto-assigned as "{IssueID}-C{n}".
	CommentID string `json:"comment_id"`

	// IssueID links this comment to its parent issue.
	IssueID string `json:"issue_id"`

	// AuthorWallet is the commenter's wallet address.
	AuthorWallet string `json:"author_wallet"`

	// AuthorUsername is the commenter's display name.
	AuthorUsername string `json:"author_username"`

	// Body is the comment text (Markdown supported).
	Body string `json:"body"`

	// CreatedAt is the wall-clock timestamp of the comment.
	CreatedAt time.Time `json:"created_at"`

	// CreatedAtBlock is the blockchain height when the comment was filed.
	CreatedAtBlock uint64 `json:"created_at_block"`
}

// IssueFilters allows filtering and sorting the issue list.
// All filters are optional; an empty IssueFilters returns all issues.
type IssueFilters struct {
	// Query is a keyword search string. Matches against title and body.
	// Multiple words are OR-combined (any match counts).
	Query string

	// Severity filters to a specific severity level.
	Severity string

	// Status filters to a specific status.
	Status string

	// Network filters to a specific network scope.
	Network string

	// Creator filters to issues filed by a specific wallet.
	Creator string

	// SortBy controls ordering: "votes" (highest score first),
	// "newest" (most recently created), "oldest", "relevance" (keyword score).
	SortBy string
}

// IssueSearchResult pairs an issue with its keyword relevance score.
// Used by SearchSimilar() to surface near-duplicate titles.
type IssueSearchResult struct {
	Issue     *TrackerIssue `json:"issue"`
	Relevance float64       `json:"relevance"`
}

// ===========================
// Persistence State
// ===========================

// issueTrackerState is the serialized form persisted to issues.json.
type issueTrackerState struct {
	Issues   []*TrackerIssue              `json:"issues"`
	Comments map[string][]*TrackerComment `json:"comments"`
	NextID   int                          `json:"next_id"`
}

// scoredTrackerIssue is an internal type pairing an issue with its relevance score
// during filtering/sorting. Defined at package level to avoid anonymous struct type mismatches.
type scoredTrackerIssue struct {
	issue *TrackerIssue
	score int
}

// ===========================
// Manager
// ===========================

// IssueTracker manages the complete lifecycle of issues for the Nakamoto project.
// State is kept in memory and persisted to JSON on each mutation.
//
// Thread safety: all public methods acquire the RWMutex appropriately.
// Blockchain persistence: when SetBlockchainStore is called, every mutation
// also stores an event in the trunk blockchain so events propagate via P2P.
type IssueTracker struct {
	config   *IssueTrackerConfig
	issues   map[string]*TrackerIssue         // issueID -> TrackerIssue
	comments map[string][]*TrackerComment     // issueID -> ordered comments
	nextID   int                              // auto-increment counter for issue IDs
	logger   *zap.Logger
	mutex    sync.RWMutex

	// blockchainStore is an optional callback wired by server.go after the
	// issue tracker is initialized.  When set, every mutation fires it
	// outside the mutex to persist the event to the trunk blockchain without
	// risking a deadlock.  Non-nil only when a UnifiedBlockchainManager is
	// available (i.e., production mode).
	blockchainStore func(issueID, eventType, authorID string, eventData []byte) error
}

// NewIssueTracker creates a new IssueTracker and loads any persisted state.
// If no persisted state exists, starts fresh with nextID = 1.
func NewIssueTracker(config *IssueTrackerConfig, logger *zap.Logger) (*IssueTracker, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration is required")
	}
	if config.StoragePath == "" {
		return nil, fmt.Errorf("storage_path is required")
	}
	if config.AutoExpireBlocks == 0 {
		// Default: 100800 blocks ≈ 14 days at 12-second block times
		config.AutoExpireBlocks = 100800
	}
	if config.MaxSpamStrikes == 0 {
		config.MaxSpamStrikes = 3
	}

	// Ensure the storage directory exists before loading
	if err := os.MkdirAll(config.StoragePath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	it := &IssueTracker{
		config:   config,
		issues:   make(map[string]*TrackerIssue),
		comments: make(map[string][]*TrackerComment),
		nextID:   1,
		logger:   logger,
	}

	// Attempt to restore previously persisted state
	if err := it.load(); err != nil {
		if !os.IsNotExist(err) {
			logger.Warn("failed to load issue tracker state, starting fresh",
				zap.Error(err))
		}
	}

	return it, nil
}

// ===========================
// Blockchain Storage Wiring
// ===========================

// SetBlockchainStore registers a callback that the tracker calls after every
// successful mutation to persist the event in the trunk blockchain.
//
// The callback must be goroutine-safe; the tracker invokes it outside the
// internal mutex to avoid deadlocks when the blockchain store acquires its
// own locks.  Errors from the store are logged at WARN level and do not fail
// the mutation — JSON persistence (issues.json) is the source of truth for
// this node; on-chain storage is best-effort replication.
func (it *IssueTracker) SetBlockchainStore(store func(issueID, eventType, authorID string, eventData []byte) error) {
	it.mutex.Lock()
	defer it.mutex.Unlock()
	it.blockchainStore = store
}

// storeOnChain fires the blockchainStore callback if one has been registered.
// It must be called OUTSIDE any held mutex to prevent deadlocks.
// Errors are logged as warnings; they never propagate to the caller.
func (it *IssueTracker) storeOnChain(issueID, eventType, authorID string, payload interface{}) {
	// Read the callback pointer under a read lock — avoids a race without
	// blocking concurrent readers longer than necessary.
	it.mutex.RLock()
	store := it.blockchainStore
	it.mutex.RUnlock()

	if store == nil {
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		it.logger.Warn("failed to marshal issue event for chain storage",
			zap.String("issue_id", issueID),
			zap.String("event_type", eventType),
			zap.Error(err),
		)
		return
	}

	if err := store(issueID, eventType, authorID, data); err != nil {
		it.logger.Warn("failed to store issue event on chain",
			zap.String("issue_id", issueID),
			zap.String("event_type", eventType),
			zap.Error(err),
		)
	}
}

// ===========================
// Issue CRUD
// ===========================

// CreateIssue validates and persists a new issue.
// Assigns the next sequential ID ("ISS-1", "ISS-2", ...).
// Required fields: Title, Severity (valid), Network (valid).
// Sets Status to "open" and initializes all maps.
func (it *IssueTracker) CreateIssue(issue *TrackerIssue) (*TrackerIssue, error) {
	if issue == nil {
		return nil, fmt.Errorf("issue is required")
	}

	// Title is mandatory — without it, the issue cannot be searched or displayed.
	if strings.TrimSpace(issue.Title) == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Severity determines triage priority — must be one of the defined levels.
	if !trackerValidSeverities[issue.Severity] {
		return nil, fmt.Errorf("invalid severity %q: must be one of critical, major, minor, enhancement", issue.Severity)
	}

	// Network scope must be specified so users know which environment to reproduce on.
	if !trackerValidNetworks[issue.Network] {
		return nil, fmt.Errorf("invalid network %q: must be one of testnet, mainnet, both", issue.Network)
	}

	it.mutex.Lock()
	defer it.mutex.Unlock()

	// Assign the next sequential ID and increment the counter atomically
	id := fmt.Sprintf("ISS-%d", it.nextID)
	it.nextID++

	now := time.Now()

	// Initialize all maps so callers never encounter nil map panics
	if issue.VoterWallets == nil {
		issue.VoterWallets = make(map[string]string)
	}
	if issue.DupeFlags == nil {
		issue.DupeFlags = make(map[string]string)
	}
	if issue.Confirmations == nil {
		issue.Confirmations = make(map[string]bool)
	}
	if issue.Tags == nil {
		issue.Tags = []string{}
	}

	issue.IssueID = id
	issue.Status = "open" // All issues start as open; devs triage from here
	issue.Score = 0
	issue.Upvotes = 0
	issue.Downvotes = 0
	issue.CreatedAt = now
	issue.UpdatedAt = now

	it.issues[id] = issue

	if err := it.saveUnsafe(); err != nil {
		it.logger.Error("failed to save after CreateIssue", zap.String("issue_id", id), zap.Error(err))
	}

	it.logger.Info("issue created",
		zap.String("issue_id", id),
		zap.String("title", issue.Title),
		zap.String("severity", issue.Severity),
	)

	// Capture values before releasing the lock so the goroutine has a stable snapshot.
	// The goroutine runs outside the mutex to prevent deadlock with blockchain locks.
	issueCopy := *issue // shallow copy is safe: maps inside are not mutated by storeOnChain
	go it.storeOnChain(id, "created", issue.CreatorWallet, &issueCopy)

	return issue, nil
}

// GetIssue returns the issue with the given ID, or nil if not found.
func (it *IssueTracker) GetIssue(issueID string) *TrackerIssue {
	it.mutex.RLock()
	defer it.mutex.RUnlock()
	return it.issues[issueID]
}

// ListIssues returns issues matching the given filters.
// If no filters are set, all issues are returned.
// When Query is set, only issues matching at least one keyword are returned,
// and SortBy "relevance" orders by match score.
func (it *IssueTracker) ListIssues(filters IssueFilters) []*TrackerIssue {
	it.mutex.RLock()
	defer it.mutex.RUnlock()

	// Build keyword list for relevance scoring
	var keywords []string
	if filters.Query != "" {
		for _, w := range strings.Fields(strings.ToLower(filters.Query)) {
			if w != "" {
				keywords = append(keywords, w)
			}
		}
	}

	var results []scoredTrackerIssue

	for _, issue := range it.issues {
		// Apply exact-match filters first (fast path)
		if filters.Severity != "" && issue.Severity != filters.Severity {
			continue
		}
		if filters.Status != "" && issue.Status != filters.Status {
			continue
		}
		if filters.Network != "" && issue.Network != filters.Network {
			continue
		}
		if filters.Creator != "" && issue.CreatorWallet != filters.Creator {
			continue
		}

		// Keyword relevance scoring:
		// Title matches are worth 3x body matches because titles are more concise
		// and a title match is a stronger signal of relevance.
		relevanceScore := 0
		if len(keywords) > 0 {
			titleLower := strings.ToLower(issue.Title)
			bodyLower := strings.ToLower(issue.Body)
			for _, kw := range keywords {
				if strings.Contains(titleLower, kw) {
					relevanceScore += 3
				}
				if strings.Contains(bodyLower, kw) {
					relevanceScore += 1
				}
			}
			// If keyword search is active, exclude issues with no match
			if relevanceScore == 0 {
				continue
			}
		}

		results = append(results, scoredTrackerIssue{issue: issue, score: relevanceScore})
	}

	// Sort results based on the requested strategy
	sortTrackerIssues(results, filters.SortBy)

	// Extract just the issue pointers for the return value
	out := make([]*TrackerIssue, len(results))
	for i, s := range results {
		out[i] = s.issue
	}
	return out
}

// sortTrackerIssues performs an in-place sort on the scored issue slice.
// Supported strategies: "votes", "newest", "oldest", "relevance".
// Default (empty string) falls through to "newest".
func sortTrackerIssues(issues []scoredTrackerIssue, sortBy string) {
	// Simple insertion sort — issue lists are expected to be small (<10k)
	for i := 1; i < len(issues); i++ {
		for j := i; j > 0; j-- {
			a, b := issues[j], issues[j-1]
			var less bool
			switch sortBy {
			case "votes":
				less = a.issue.Score > b.issue.Score
			case "oldest":
				less = a.issue.CreatedAt.Before(b.issue.CreatedAt)
			case "relevance":
				less = a.score > b.score
			default: // "newest" or unspecified
				less = a.issue.CreatedAt.After(b.issue.CreatedAt)
			}
			if less {
				issues[j], issues[j-1] = issues[j-1], issues[j]
			} else {
				break
			}
		}
	}
}

// ===========================
// Voting
// ===========================

// VoteOnIssue records a vote on an issue from voterWallet.
// voteType must be "up" or "down".
// Each wallet gets one vote, but can switch direction (score adjusts by 2).
// The creator cannot vote on their own issue to prevent self-promotion.
func (it *IssueTracker) VoteOnIssue(issueID, voterWallet, voteType string) error {
	if voteType != "up" && voteType != "down" {
		return fmt.Errorf("invalid vote type %q: must be 'up' or 'down'", voteType)
	}

	it.mutex.Lock()
	defer it.mutex.Unlock()

	issue, ok := it.issues[issueID]
	if !ok {
		return fmt.Errorf("issue %s not found", issueID)
	}

	// Prevent self-voting: creators have an obvious conflict of interest
	if voterWallet == issue.CreatorWallet {
		return fmt.Errorf("creator cannot vote on their own issue")
	}

	existing, alreadyVoted := issue.VoterWallets[voterWallet]
	if alreadyVoted {
		if existing == voteType {
			// Same direction again — idempotent rejection
			return fmt.Errorf("wallet %s already voted %s on issue %s", voterWallet, voteType, issueID)
		}

		// Switching vote: adjust all counters and score atomically.
		// When switching from up→down: remove +1 upvote, add -1 downvote → net change = -2
		// When switching from down→up: remove -1 downvote, add +1 upvote → net change = +2
		if existing == "up" {
			issue.Upvotes--
			issue.Downvotes++
			issue.Score -= 2
		} else {
			issue.Downvotes--
			issue.Upvotes++
			issue.Score += 2
		}
		// Update direction record only — counters already adjusted above
		issue.VoterWallets[voterWallet] = voteType
	} else {
		// First-time vote: record and adjust counters
		issue.VoterWallets[voterWallet] = voteType
		if voteType == "up" {
			issue.Upvotes++
			issue.Score++
		} else {
			issue.Downvotes++
			issue.Score--
		}
	}
	issue.UpdatedAt = time.Now()

	if err := it.saveUnsafe(); err != nil {
		it.logger.Error("failed to save after VoteOnIssue", zap.Error(err))
	}

	// Capture stable values before the mutex is released.
	voteIssueID := issueID
	voteVoterWallet := voterWallet
	voteType_ := voteType
	go it.storeOnChain(voteIssueID, "voted", voteVoterWallet, map[string]interface{}{
		"issue_id": voteIssueID,
		"voter":    voteVoterWallet,
		"vote":     voteType_,
	})

	return nil
}

// ===========================
// Confirmation
// ===========================

// ConfirmIssue records that confirmerWallet has successfully reproduced the issue.
// Issues with 0 confirmations are auto-expired after AutoExpireBlocks.
// A wallet can only confirm once; the creator cannot confirm their own issue.
func (it *IssueTracker) ConfirmIssue(issueID, confirmerWallet string) error {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	issue, ok := it.issues[issueID]
	if !ok {
		return fmt.Errorf("issue %s not found", issueID)
	}

	// Creators confirming their own issue adds no independent verification signal
	if confirmerWallet == issue.CreatorWallet {
		return fmt.Errorf("creator cannot confirm their own issue")
	}

	if issue.Confirmations[confirmerWallet] {
		return fmt.Errorf("wallet %s already confirmed issue %s", confirmerWallet, issueID)
	}

	issue.Confirmations[confirmerWallet] = true
	issue.UpdatedAt = time.Now()

	if err := it.saveUnsafe(); err != nil {
		it.logger.Error("failed to save after ConfirmIssue", zap.Error(err))
	}

	// Capture stable values for the goroutine outside the mutex.
	confirmIssueID := issueID
	confirmWallet := confirmerWallet
	go it.storeOnChain(confirmIssueID, "confirmed", confirmWallet, map[string]interface{}{
		"issue_id":   confirmIssueID,
		"confirmer":  confirmWallet,
	})

	return nil
}

// ===========================
// Status Management
// ===========================

// SetStatus changes the status of an issue.
// Only users with the "dev" role can change status — contributors cannot triage.
// The new status must be one of the valid status values.
func (it *IssueTracker) SetStatus(issueID, newStatus, actorWallet, actorRole string) error {
	// Only developers can change issue status — this prevents contributors from
	// self-closing issues they disagree with
	if actorRole != "dev" {
		return fmt.Errorf("only dev-role users can change issue status")
	}

	if !trackerValidStatuses[newStatus] {
		return fmt.Errorf("invalid status %q", newStatus)
	}

	it.mutex.Lock()
	defer it.mutex.Unlock()

	issue, ok := it.issues[issueID]
	if !ok {
		return fmt.Errorf("issue %s not found", issueID)
	}

	issue.Status = newStatus
	issue.UpdatedAt = time.Now()

	if err := it.saveUnsafe(); err != nil {
		it.logger.Error("failed to save after SetStatus", zap.Error(err))
	}

	// Capture stable values for the goroutine outside the mutex.
	statusIssueID := issueID
	statusNewStatus := newStatus
	statusActor := actorWallet
	go it.storeOnChain(statusIssueID, "status_changed", statusActor, map[string]interface{}{
		"issue_id":   statusIssueID,
		"new_status": statusNewStatus,
		"actor":      statusActor,
	})

	return nil
}

// ResolveIssue marks an issue as resolved with attribution and a resolution note.
// Sets status to "resolved", records the resolver's wallet.
// Only devs can resolve issues.
func (it *IssueTracker) ResolveIssue(issueID, resolverWallet, resolverRole, note string) error {
	if resolverRole != "dev" {
		return fmt.Errorf("only dev-role users can resolve issues")
	}

	it.mutex.Lock()
	defer it.mutex.Unlock()

	issue, ok := it.issues[issueID]
	if !ok {
		return fmt.Errorf("issue %s not found", issueID)
	}

	issue.Status = "resolved"
	issue.ResolverWallet = resolverWallet
	issue.ResolutionNote = note
	issue.UpdatedAt = time.Now()

	if err := it.saveUnsafe(); err != nil {
		it.logger.Error("failed to save after ResolveIssue", zap.Error(err))
	}

	// Capture stable values for the goroutine outside the mutex.
	resolveIssueID := issueID
	resolveWallet := resolverWallet
	resolveNote := note
	go it.storeOnChain(resolveIssueID, "resolved", resolveWallet, map[string]interface{}{
		"issue_id": resolveIssueID,
		"resolver": resolveWallet,
		"note":     resolveNote,
	})

	return nil
}

// CloseAsInvalid marks an issue as invalid with a mandatory reason.
// Used when the issue describes expected behavior, is a user error,
// or cannot be reproduced on any known configuration.
// Reason is required to provide transparency to the filer.
func (it *IssueTracker) CloseAsInvalid(issueID, reason, actorRole string) error {
	if actorRole != "dev" {
		return fmt.Errorf("only dev-role users can close issues as invalid")
	}
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("reason is required when closing an issue as invalid")
	}

	it.mutex.Lock()
	defer it.mutex.Unlock()

	issue, ok := it.issues[issueID]
	if !ok {
		return fmt.Errorf("issue %s not found", issueID)
	}

	issue.Status = "invalid"
	issue.InvalidReason = reason
	issue.UpdatedAt = time.Now()

	if err := it.saveUnsafe(); err != nil {
		it.logger.Error("failed to save after CloseAsInvalid", zap.Error(err))
	}

	// Capture stable values for the goroutine outside the mutex.
	closedIssueID := issueID
	closedReason := reason
	go it.storeOnChain(closedIssueID, "closed_invalid", "dev", map[string]interface{}{
		"issue_id": closedIssueID,
		"reason":   closedReason,
	})

	return nil
}

// MarkAsSpam marks an issue as spam and adds a spam strike to the creator's account.
// After MaxSpamStrikes, the user is blocked from filing new issues.
// userRegistry is required to record the strike against the creator's username.
func (it *IssueTracker) MarkAsSpam(issueID, actorRole string, userRegistry *UserRegistry) error {
	if actorRole != "dev" {
		return fmt.Errorf("only dev-role users can mark issues as spam")
	}

	it.mutex.Lock()
	defer it.mutex.Unlock()

	issue, ok := it.issues[issueID]
	if !ok {
		return fmt.Errorf("issue %s not found", issueID)
	}

	issue.Status = "spam"
	issue.UpdatedAt = time.Now()

	// Record the spam strike so the user's filing privileges can be revoked
	// after MaxSpamStrikes. This call is outside any UserRegistry lock so we
	// must accept that the registry update and issue update are not atomic.
	if userRegistry != nil && issue.CreatorUsername != "" {
		if _, err := userRegistry.AddSpamStrike(issue.CreatorUsername); err != nil {
			it.logger.Warn("failed to add spam strike",
				zap.String("username", issue.CreatorUsername),
				zap.Error(err),
			)
		}
	}

	if err := it.saveUnsafe(); err != nil {
		it.logger.Error("failed to save after MarkAsSpam", zap.Error(err))
	}

	// Capture stable value for the goroutine outside the mutex.
	spamIssueID := issueID
	go it.storeOnChain(spamIssueID, "marked_spam", "dev", map[string]interface{}{
		"issue_id": spamIssueID,
	})

	return nil
}

// ===========================
// Duplicate Management
// ===========================

// FlagDuplicate records that flaggerWallet believes issueID is a duplicate of targetIssueID.
// Auto-confirms the duplicate if:
//   - 2 or more community members flag the same target, OR
//   - The flagger has the "dev" role (trusted signal)
//
// When a duplicate is confirmed, votes from the duplicate are merged into the target
// to preserve community signal without double-counting.
func (it *IssueTracker) FlagDuplicate(issueID, targetIssueID, flaggerWallet, flaggerRole string) error {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	issue, ok := it.issues[issueID]
	if !ok {
		return fmt.Errorf("issue %s not found", issueID)
	}

	target, ok := it.issues[targetIssueID]
	if !ok {
		return fmt.Errorf("target issue %s not found", targetIssueID)
	}

	// Record this wallet's duplicate flag (overwrites previous flag from same wallet)
	issue.DupeFlags[flaggerWallet] = targetIssueID

	// Count how many wallets have flagged the same target
	targetFlagCount := 0
	for _, flaggedTarget := range issue.DupeFlags {
		if flaggedTarget == targetIssueID {
			targetFlagCount++
		}
	}

	// Auto-confirm if threshold is met or a dev flags it
	shouldConfirm := targetFlagCount >= trackerDupeFlagThreshold || flaggerRole == "dev"

	if shouldConfirm && !issue.DupeConfirmed {
		issue.DupeConfirmed = true
		issue.DuplicateOf = targetIssueID
		issue.UpdatedAt = time.Now()

		// Merge votes from the duplicate into the target issue.
		// Only add votes that don't conflict with existing votes on the target.
		// This preserves existing target votes and adds new community signal.
		it.mergeVotesUnsafe(issue, target)

		it.logger.Info("duplicate confirmed and votes merged",
			zap.String("dupe_id", issueID),
			zap.String("target_id", targetIssueID),
		)
	} else {
		issue.UpdatedAt = time.Now()
	}

	if err := it.saveUnsafe(); err != nil {
		it.logger.Error("failed to save after FlagDuplicate", zap.Error(err))
	}

	// Capture stable values for the goroutine outside the mutex.
	flagIssueID := issueID
	flagTargetID := targetIssueID
	flagWallet := flaggerWallet
	go it.storeOnChain(flagIssueID, "flagged_duplicate", flagWallet, map[string]interface{}{
		"issue_id":     flagIssueID,
		"duplicate_of": flagTargetID,
		"flagger":      flagWallet,
	})

	return nil
}

// mergeVotesUnsafe transfers votes from the duplicate issue to the target issue.
// Votes that conflict with existing target votes are skipped.
// Caller must hold the write lock.
func (it *IssueTracker) mergeVotesUnsafe(dupe, target *TrackerIssue) {
	for wallet, voteType := range dupe.VoterWallets {
		// Skip if this wallet has already voted on the target
		if _, exists := target.VoterWallets[wallet]; exists {
			continue
		}

		// Add the vote to the target
		target.VoterWallets[wallet] = voteType
		if voteType == "up" {
			target.Upvotes++
			target.Score++
		} else {
			target.Downvotes++
			target.Score--
		}
	}
	target.UpdatedAt = time.Now()
}

// ===========================
// Search
// ===========================

// SearchSimilar finds issues with titles similar to the given title.
// Used during issue filing to surface potential duplicates.
// Returns up to 5 results with relevance > 0, ordered by relevance descending.
//
// Relevance is the number of title words from the query that appear in each issue's title.
// This is intentionally simple — exact substring match per word is sufficient for
// duplicate detection where full text search would be over-engineered.
func (it *IssueTracker) SearchSimilar(title string) []*IssueSearchResult {
	it.mutex.RLock()
	defer it.mutex.RUnlock()

	words := strings.Fields(strings.ToLower(title))
	if len(words) == 0 {
		return nil
	}

	var results []*IssueSearchResult

	for _, issue := range it.issues {
		titleLower := strings.ToLower(issue.Title)
		score := 0
		for _, w := range words {
			if strings.Contains(titleLower, w) {
				score++
			}
		}
		if score > 0 {
			results = append(results, &IssueSearchResult{
				Issue:     issue,
				Relevance: float64(score),
			})
		}
	}

	// Sort by relevance descending (insertion sort — small N)
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].Relevance > results[j-1].Relevance; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	// Return top 5 matches only to avoid overwhelming the UI
	if len(results) > 5 {
		results = results[:5]
	}

	return results
}

// ===========================
// Comments
// ===========================

// AddComment attaches a comment to an existing issue.
// Validates that the parent issue exists and assigns a unique CommentID.
func (it *IssueTracker) AddComment(comment *TrackerComment) error {
	if comment == nil {
		return fmt.Errorf("comment is required")
	}

	it.mutex.Lock()
	defer it.mutex.Unlock()

	if _, ok := it.issues[comment.IssueID]; !ok {
		return fmt.Errorf("issue %s not found", comment.IssueID)
	}

	// Generate a stable CommentID: "{IssueID}-C{n}" where n is 1-indexed
	existing := it.comments[comment.IssueID]
	comment.CommentID = fmt.Sprintf("%s-C%d", comment.IssueID, len(existing)+1)
	comment.CreatedAt = time.Now()

	it.comments[comment.IssueID] = append(existing, comment)

	if err := it.saveUnsafe(); err != nil {
		it.logger.Error("failed to save after AddComment", zap.Error(err))
	}

	// Capture a snapshot of the comment for the goroutine outside the mutex.
	commentCopy := *comment
	go it.storeOnChain(commentCopy.IssueID, "commented", commentCopy.AuthorWallet, &commentCopy)

	return nil
}

// ListComments returns all comments for the given issue, in order of creation.
// Returns an empty slice if the issue has no comments.
func (it *IssueTracker) ListComments(issueID string) []*TrackerComment {
	it.mutex.RLock()
	defer it.mutex.RUnlock()
	return it.comments[issueID]
}

// ===========================
// Auto-Expiry
// ===========================

// ExpireUnconfirmedIssues scans all open issues and closes those that have:
//   - Status "open" (not yet triaged)
//   - Zero confirmations (nobody could reproduce it)
//   - Age >= AutoExpireBlocks blocks
//
// This garbage-collects stale unconfirmed issues to keep the tracker clean.
// Returns the number of issues that were expired in this pass.
func (it *IssueTracker) ExpireUnconfirmedIssues(currentBlock uint64) int {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	expired := 0

	for _, issue := range it.issues {
		// Only expire issues that are still open and unconfirmed
		if issue.Status != "open" {
			continue
		}
		if len(issue.Confirmations) > 0 {
			continue
		}

		// Check if the issue has lived past its expiry window.
		// We use block height arithmetic to avoid wall-clock drift.
		age := uint64(0)
		if currentBlock >= issue.CreatedAtBlock {
			age = currentBlock - issue.CreatedAtBlock
		}

		if age >= it.config.AutoExpireBlocks {
			issue.Status = "cannot-reproduce"
			issue.UpdatedAt = time.Now()
			expired++

			it.logger.Info("issue auto-expired (no confirmations)",
				zap.String("issue_id", issue.IssueID),
				zap.Uint64("age_blocks", age),
			)
		}
	}

	if expired > 0 {
		if err := it.saveUnsafe(); err != nil {
			it.logger.Error("failed to save after ExpireUnconfirmedIssues", zap.Error(err))
		}
	}

	return expired
}

// ===========================
// Persistence
// ===========================

// Save persists the current in-memory state to disk.
// Called automatically after each mutation; can also be called explicitly.
func (it *IssueTracker) Save() error {
	it.mutex.Lock()
	defer it.mutex.Unlock()
	return it.saveUnsafe()
}

// saveUnsafe is the internal (lock-free) save implementation.
// Caller must hold the write lock.
func (it *IssueTracker) saveUnsafe() error {
	issues := make([]*TrackerIssue, 0, len(it.issues))
	for _, issue := range it.issues {
		issues = append(issues, issue)
	}

	state := &issueTrackerState{
		Issues:   issues,
		Comments: it.comments,
		NextID:   it.nextID,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal issue tracker state: %w", err)
	}

	path := filepath.Join(it.config.StoragePath, "issues.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write issues.json: %w", err)
	}

	it.logger.Debug("issue tracker state saved", zap.String("path", path))
	return nil
}

// load reads persisted state from issues.json and restores in-memory maps.
// Called once during NewIssueTracker. Returns os.ErrNotExist if no file yet.
func (it *IssueTracker) load() error {
	path := filepath.Join(it.config.StoragePath, "issues.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err // caller handles os.IsNotExist
	}

	var state issueTrackerState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal issues.json: %w", err)
	}

	// Restore issues map
	for _, issue := range state.Issues {
		// Ensure maps are never nil after deserialization to prevent panics
		if issue.VoterWallets == nil {
			issue.VoterWallets = make(map[string]string)
		}
		if issue.DupeFlags == nil {
			issue.DupeFlags = make(map[string]string)
		}
		if issue.Confirmations == nil {
			issue.Confirmations = make(map[string]bool)
		}
		it.issues[issue.IssueID] = issue
	}

	// Restore comments map
	if state.Comments != nil {
		for issueID, comments := range state.Comments {
			it.comments[issueID] = comments
		}
	}

	// Restore the ID counter so new issues don't collide with persisted ones
	it.nextID = state.NextID
	if it.nextID < 1 {
		it.nextID = 1
	}

	it.logger.Info("issue tracker state loaded",
		zap.Int("issues", len(it.issues)),
		zap.Int("next_id", it.nextID),
	)

	return nil
}

// Close flushes any pending state and releases resources.
// Should be called when the tracker is no longer needed (e.g., in tests via t.Cleanup).
func (it *IssueTracker) Close() {
	// Attempt a final save on Close. Errors are logged but not returned
	// since Close() is typically called in deferred cleanup paths.
	if err := it.Save(); err != nil {
		it.logger.Warn("failed to save on Close", zap.Error(err))
	}
}

////////////////////////////////////////////////////////////////////////////////
// File: internal/core/contribution_scorer.go
//
// Summary:
//   Auto-calculates contribution points and BPS (basis points) for wallets
//   based on their activity in the IssueTracker. Points flow from confirmed
//   issues, community votes, resolutions, and node uptime.
//
//   SCORING RULES (all points are 0 for "dev" role — devs are staff, not
//   community contributors):
//   - Issue filed (confirmed/in-progress/resolved/closed): SeverityBasePoints[severity]
//   - Vote bonus: int(base * netUpvotes * 0.10), clamped to >= 0
//   - Resolution: SeverityBasePoints[severity] * 2  (to the resolver)
//   - Uptime: 5 pts/day, capped to monthCap per AddNodeUptime call
//
//   BPS CALCULATION:
//   - bps = (walletPoints / totalAllPoints) * 10000
//   - Sum of all BPS may differ from 10000 by ±1 due to integer rounding
//
// Dependencies:
//   - issue_tracker_impl.go: IssueTracker, TrackerIssue
//   - contributor_share_manager.go: ContributorShareManager, Contributor
//   - user_registry.go: UserRegistry, GetUserRole
//
// Associated Test Files:
//   - internal/core/contribution_scorer_test.go
////////////////////////////////////////////////////////////////////////////////

package core

import (
	"fmt"

	"go.uber.org/zap"
)

// ===========================
// Constants
// ===========================

// SeverityBasePoints maps issue severity levels to their base contribution points.
// Points are awarded when an issue advances past "open" to a confirmed state.
// Values chosen to give meaningful weight differences between severity tiers.
var SeverityBasePoints = map[string]int{
	"critical":    100,
	"major":       50,
	"minor":       20,
	"enhancement": 30,
}

// ===========================
// Types
// ===========================

// ContributionScorer calculates contribution points from IssueTracker activity
// and uptime data, then converts them to BPS for the ContributorShareManager.
type ContributionScorer struct {
	issueTracker *IssueTracker
	// uptimePoints maps walletID → accumulated uptime points earned via AddNodeUptime.
	// Node uptime is tracked outside the issue system, so it is stored separately here.
	uptimePoints map[string]int
	logger       *zap.Logger
}

// ContributionScore holds the full breakdown of points for a single wallet,
// plus the derived BPS value (populated by SnapshotBPS).
type ContributionScore struct {
	WalletID string `json:"wallet_id"`
	Username string `json:"username"`
	// Role is the wallet owner's role ("contributor" or "dev").
	// Dev-role wallets always have zero points.
	Role              string `json:"role"`
	TotalPoints       int    `json:"total_points"`
	IssuesFiledPoints int    `json:"issues_filed_points"`
	VotePoints        int    `json:"vote_points"`
	ResolutionPoints  int    `json:"resolution_points"`
	UptimePoints      int    `json:"uptime_points"`
	// CalculatedBPS is filled by SnapshotBPS; zero in CalculateScore output.
	CalculatedBPS int `json:"calculated_bps"`
}

// ===========================
// Constructor
// ===========================

// NewContributionScorer creates a ContributionScorer backed by the given IssueTracker.
// The scorer is stateless except for the uptime map, which is built up via AddNodeUptime.
func NewContributionScorer(tracker *IssueTracker, logger *zap.Logger) *ContributionScorer {
	return &ContributionScorer{
		issueTracker: tracker,
		uptimePoints: make(map[string]int),
		logger:       logger,
	}
}

// ===========================
// Scoring
// ===========================

// CalculateScore computes the full contribution breakdown for a single wallet.
//
// Scoring rules:
//   - Dev-role wallets always receive 0 points (devs are staff, not community).
//   - Only issues with status "confirmed", "in-progress", "resolved", or "closed"
//     earn points. Open, cannot-reproduce, invalid, and spam statuses are skipped.
//   - Confirmed duplicates (DupeConfirmed=true) earn 0 points — the original issue
//     already captures the contribution.
//   - Upvote bonus = int(base * max(score, 0) * 0.10). Clamped to 0 to prevent
//     downvote campaigns from removing earned base points.
//   - Resolution points go to the ResolverWallet at 2× base points.
func (cs *ContributionScorer) CalculateScore(walletID, role string) *ContributionScore {
	score := &ContributionScore{
		WalletID: walletID,
		Role:     role,
	}

	// Devs earn zero contribution points by design (Whitepaper Section: Contributor Rewards).
	// This ensures the community — not paid developers — drives the contribution metric.
	if role == "dev" {
		return score
	}

	issues := cs.issueTracker.ListIssues(IssueFilters{})

	for _, issue := range issues {
		// Skip issues that haven't progressed past initial filing.
		// Only "confirmed", "in-progress", "resolved", "closed" represent genuine
		// validated contributions worth rewarding.
		if issue.Status == "open" ||
			issue.Status == "cannot-reproduce" ||
			issue.Status == "invalid" ||
			issue.Status == "spam" {
			continue
		}

		// Confirmed duplicates are merged into the original issue.
		// The original creator gets credit; the duplicate filer gets none.
		if issue.DupeConfirmed {
			continue
		}

		base, ok := SeverityBasePoints[issue.Severity]
		if !ok {
			// Unknown severity — skip rather than guess a value.
			cs.logger.Warn("unknown severity in issue, skipping points",
				zap.String("issue_id", issue.IssueID),
				zap.String("severity", issue.Severity))
			continue
		}

		// Award filing points to the creator.
		// The vote bonus reflects community endorsement: each net upvote adds 10% of
		// the base value, incentivising quality filing and discouraging noise.
		if issue.CreatorWallet == walletID {
			score.IssuesFiledPoints += base

			netScore := issue.Score
			if netScore < 0 {
				// Clamp negative scores to 0 — downvote campaigns should not erase
				// legitimate filing points already awarded.
				netScore = 0
			}
			upvoteBonus := int(float64(base) * float64(netScore) * 0.10)
			score.VotePoints += upvoteBonus
		}

		// Award resolution points to the resolver.
		// Resolvers earn 2× base to reflect the extra effort of fixing vs. filing.
		if issue.ResolverWallet == walletID {
			score.ResolutionPoints += base * 2
		}
	}

	// Accumulate uptime points recorded by AddNodeUptime calls.
	score.UptimePoints = cs.uptimePoints[walletID]

	score.TotalPoints = score.IssuesFiledPoints + score.VotePoints +
		score.ResolutionPoints + score.UptimePoints

	return score
}

// CalculateAllScores computes ContributionScore for every wallet that appears
// in the issue tracker (as creator or resolver) or has uptime points recorded.
//
// userRegistry is used to look up roles. If nil, all wallets default to "contributor".
func (cs *ContributionScorer) CalculateAllScores(userRegistry *UserRegistry) map[string]*ContributionScore {
	// Collect all unique wallet IDs that could have points.
	wallets := make(map[string]bool)

	issues := cs.issueTracker.ListIssues(IssueFilters{})
	for _, issue := range issues {
		if issue.CreatorWallet != "" {
			wallets[issue.CreatorWallet] = true
		}
		if issue.ResolverWallet != "" {
			wallets[issue.ResolverWallet] = true
		}
	}

	// Also include wallets that have uptime points even without filed issues.
	for walletID := range cs.uptimePoints {
		wallets[walletID] = true
	}

	scores := make(map[string]*ContributionScore)

	for walletID := range wallets {
		role := "contributor"
		username := walletID // default username = walletID when not registered

		if userRegistry != nil {
			// Attempt to look up the username from the wallet to find the role.
			// GetUsernameForWallet returns an error if not registered — ignore it.
			if un, err := userRegistry.GetUsernameForWallet(walletID); err == nil && un != "" {
				username = un
				role = userRegistry.GetUserRole(un)
			}
		}

		s := cs.CalculateScore(walletID, role)
		s.Username = username
		scores[walletID] = s
	}

	return scores
}

// SnapshotBPS converts contribution scores into basis points (BPS) suitable for
// the ContributorShareManager. BPS is proportional to TotalPoints.
//
// Only contributors (role != "dev") with TotalPoints > 0 receive BPS.
// The sum of all BPS may differ from 10000 by ±1 due to integer rounding.
//
// Returns a map of walletID → BPS. An empty map is returned if no one has points.
func (cs *ContributionScorer) SnapshotBPS(userRegistry *UserRegistry) map[string]int {
	allScores := cs.CalculateAllScores(userRegistry)

	// Sum all eligible points to derive the denominator for proportional BPS.
	totalPoints := 0
	for _, s := range allScores {
		if s.Role != "dev" && s.TotalPoints > 0 {
			totalPoints += s.TotalPoints
		}
	}

	result := make(map[string]int)

	if totalPoints == 0 {
		// No points awarded yet — return empty map rather than dividing by zero.
		return result
	}

	for walletID, s := range allScores {
		if s.Role == "dev" || s.TotalPoints <= 0 {
			continue
		}
		bps := (s.TotalPoints * 10000) / totalPoints
		if bps > 0 {
			result[walletID] = bps
		}
	}

	return result
}

// AddNodeUptime records uptime contribution points for a wallet.
// Each day of verified uptime earns 5 points. The total is capped at monthCap
// per call to prevent gaming (a month of uptime should cap at 150 by default).
//
// Example: AddNodeUptime("wallet1", 20, 150) → adds 100 pts (20*5).
// Example: AddNodeUptime("wallet1", 40, 150) → adds 150 pts (capped at monthCap=150).
func (cs *ContributionScorer) AddNodeUptime(walletID string, days int, monthCap int) {
	points := days * 5
	if points > monthCap {
		points = monthCap
	}
	cs.uptimePoints[walletID] += points
}

// ExportToContributorContract snaps BPS and registers all qualifying contributors
// into the given ContributorShareManager. This bridges the IssueTracker-based
// contribution scoring with the on-chain revenue distribution system.
//
// Only contributors with BPS > 0 are exported. Dev-role wallets are excluded.
// If a contributor is already registered in the CSM, AddContributor will return
// an error — callers should ensure a fresh contract before exporting.
func (cs *ContributionScorer) ExportToContributorContract(csm *ContributorShareManager, userRegistry *UserRegistry) error {
	bpsMap := cs.SnapshotBPS(userRegistry)

	if len(bpsMap) == 0 {
		cs.logger.Info("no contributors with points to export")
		return nil
	}

	// Build a username lookup from all scored contributors.
	allScores := cs.CalculateAllScores(userRegistry)

	for walletID, bps := range bpsMap {
		if bps <= 0 {
			continue
		}

		username := walletID // fallback
		if s, ok := allScores[walletID]; ok && s.Username != "" {
			username = s.Username
		}

		contributor := &Contributor{
			WalletID: walletID,
			Alias:    username,
			ShareBPS: bps,
		}

		if err := csm.AddContributor(contributor); err != nil {
			return fmt.Errorf("failed to add contributor %s: %w", walletID, err)
		}

		cs.logger.Info("exported contributor to contract",
			zap.String("wallet_id", walletID),
			zap.String("alias", username),
			zap.Int("bps", bps))
	}

	return nil
}

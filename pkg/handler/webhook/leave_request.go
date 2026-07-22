package webhook

// Contractor leave self-service submission (SPEC-088). The submission-side twin of the SPEC-087
// approval endpoints and of gen-invoice: Neko Bot -> Hermes gateway (identity token) -> fortress-mcp
// -> this endpoint with a REAL fortress API key. The endpoint takes the requester identity ONLY from
// the authenticated handle the gateway put in requester_handle (never model-supplied text), resolves
// it to an ACTIVE contractor by Discord handle (net-new resolver, DEC-008/009), validates the
// parameters, idempotently creates a Status=New "Unavailability Notices" row (Contractor + Type +
// dates set, DEC-013), and drives the SAME notification core the automation path uses
// (announceNewLeaveRequest, DEC-008) so a lead's SPEC-087 approve/reject runs unchanged.
//
// Unlike gen-invoice (no HTTP auth, trusts its body), this creates a row AS a named contractor, so it
// is gated by fortress's API-key + permission middleware in routes/v1.go (same as the leave decision
// routes, DEC-006/AMEND-001); an unauthenticated create-as-contractor endpoint would be an
// impersonation-write bypass.

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	notionSvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// SPEC-088 DEC-012 execute-time defaults (named consts, sensible defaults; the exact day-bounds are
// an ops call at ship time). A bounded BACKDATE window keeps parity with the off.d.foundation form,
// which accepts a backdated sick day; a max future HORIZON caps obviously-bogus far-future requests.
const (
	maxLeaveBackdateDays = 14
	maxLeaveHorizonDays  = 365
)

// maxLeaveRequestsPerDay is the per-contractor daily submission cap (gen_invoice parity, SPEC-088).
const maxLeaveRequestsPerDay = 3

// leaveIdempotencyTTL is how long a (contractor,start,end,type) key stays claimed after a successful
// submission, collapsing a replayed identical request to a 409. It comfortably covers the gateway
// identity token's 120s TTL (DEC-002), which is the actual replay window.
const leaveIdempotencyTTL = 10 * time.Minute

// defaultLeaveTypes is the accepted "Unavailability Type" option set.
//
// SPEC-088 open question / DEC-013: the authoritative set is the LIVE Notion "Unavailability Type"
// select options, which can only be read against live Notion. This default mirrors the set documented
// on notion.LeaveRequest and is what the request-leave skill lists to the contractor; reconcile it
// with the live options at ship time.
var defaultLeaveTypes = []string{
	"Personal Time",
	"Health / Illness",
	"Family / Emergency",
	"Travel / Vacation",
	"Other",
}

// LeaveRequestPayload is the body the fortress MCP posts for a self-service leave submission.
// requester_handle is the AUTHENTICATED contractor handle (from the gateway token), NOT model text.
type LeaveRequestPayload struct {
	RequesterHandle string `json:"requester_handle"`
	StartDate       string `json:"start_date"` // YYYY-MM-DD
	EndDate         string `json:"end_date"`   // YYYY-MM-DD
	Type            string `json:"type"`
	Reason          string `json:"reason"` // optional; -> Additional Context
}

// validateLeaveRequestParams validates the submission SHAPE and returns the parsed dates plus the
// canonical type. On failure it returns a non-empty code and a user-facing message; the dates are
// only meaningful when code == "". now is injected so the backdate/horizon bounds are testable.
func validateLeaveRequestParams(startStr, endStr, leaveType string, allowedTypes []string, now time.Time) (start, end time.Time, canonicalType, code, msg string) {
	start, err := time.Parse("2006-01-02", strings.TrimSpace(startStr))
	if err != nil {
		return start, end, "", "invalid_start_date", "start_date must be a valid YYYY-MM-DD date"
	}
	end, err = time.Parse("2006-01-02", strings.TrimSpace(endStr))
	if err != nil {
		return start, end, "", "invalid_end_date", "end_date must be a valid YYYY-MM-DD date"
	}
	if end.Before(start) {
		return start, end, "", "end_before_start", "end_date must not be before start_date"
	}

	today := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)
	if start.Before(today.AddDate(0, 0, -maxLeaveBackdateDays)) {
		return start, end, "", "backdated", fmt.Sprintf("start_date is backdated more than %d days", maxLeaveBackdateDays)
	}
	if start.After(today.AddDate(0, 0, maxLeaveHorizonDays)) {
		return start, end, "", "too_far", fmt.Sprintf("start_date is more than %d days in the future", maxLeaveHorizonDays)
	}

	canonicalType, ok := canonicalLeaveType(leaveType, allowedTypes)
	if !ok {
		return start, end, "", "invalid_type", fmt.Sprintf("type must be one of: %s", strings.Join(allowedTypes, ", "))
	}
	return start, end, canonicalType, "", ""
}

// canonicalLeaveType matches an input type against the allowed set (case/space tolerant) and returns
// the canonical option name to write to the Notion select.
func canonicalLeaveType(input string, allowedTypes []string) (string, bool) {
	want := strings.ToLower(strings.TrimSpace(input))
	for _, t := range allowedTypes {
		if strings.ToLower(t) == want {
			return t, true
		}
	}
	return "", false
}

// pickActiveContractor narrows the (deliberately broad) by-Discord Notion query to the single ACTIVE
// contractor whose handle EXACTLY matches (after normalization) the requester (SPEC-088 TASK-003).
// Returns not-found for an unknown handle and for an inactive/departed contractor. Kept pure so the
// resolver's decision is unit-testable without live Notion.
func pickActiveContractor(candidates []*notionSvc.ContractorDetails, handle string) (*notionSvc.ContractorDetails, bool) {
	want := normalizeHandle(handle)
	for _, c := range candidates {
		if c == nil {
			continue
		}
		if normalizeHandle(c.DiscordUsername) != want {
			continue // over-fetched by "contains"; require an exact normalized match
		}
		if !strings.EqualFold(strings.TrimSpace(c.Status), "Active") {
			continue // inactive / departed contractor -> treated as not an active contractor
		}
		return c, true
	}
	return nil, false
}

// leaveIdempotencyKey is the deterministic dedup key over the identity-bearing fields of a request
// (DEC-011). Two concurrent or replayed identical submissions produce the same key.
func leaveIdempotencyKey(contractorPageID string, start, end time.Time, leaveType string) string {
	raw := strings.Join([]string{
		contractorPageID,
		start.Format("2006-01-02"),
		end.Format("2006-01-02"),
		strings.ToLower(strings.TrimSpace(leaveType)),
	}, "|")
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// leaveIdempotencyClaims is the race-safe idempotency guard (DEC-011). It maps an idempotency key to
// the time the claim expires. claimLeaveIdempotency is atomic (LoadOrStore), so of N concurrent
// identical requests EXACTLY ONE claims the key and proceeds to create the row; the rest see the key
// already held and 409. This is the "advisory" arm of DEC-011; the api deployment is single-replica
// (see leave_decision.go's leaveDecisionLocks note), so an in-process guard is sufficient. A
// multi-replica deployment would additionally need a DB unique constraint on the key.
var leaveIdempotencyClaims sync.Map // key -> time.Time (expiry)

// claimLeaveIdempotency atomically claims key for ttl. Returns true if the caller now OWNS the key
// (must proceed and eventually create or release), false if an unexpired claim already exists (409).
func claimLeaveIdempotency(key string, now time.Time, ttl time.Duration) bool {
	expiry := now.Add(ttl)
	for {
		actual, loaded := leaveIdempotencyClaims.LoadOrStore(key, expiry)
		if !loaded {
			return true // we stored it; we own the claim
		}
		if exp, ok := actual.(time.Time); ok && now.After(exp) {
			// Existing claim expired; try to take it over atomically.
			if leaveIdempotencyClaims.CompareAndSwap(key, actual, expiry) {
				return true
			}
			continue // lost the race to another taker; re-evaluate
		}
		return false // a live claim is held by someone else
	}
}

// releaseLeaveIdempotency frees a claim so a retry can proceed. Called when the create fails, so a
// genuine failure does not permanently wedge that (contractor,dates,type) behind a 409.
func releaseLeaveIdempotency(key string) {
	leaveIdempotencyClaims.Delete(key)
}

// dailyLimiter is a small in-process per-key daily counter (contractor -> submissions today),
// resetting at UTC midnight. Mirrors the shape of the invoice rate limiter but stays local so the
// endpoint needs no server-init wiring.
type dailyLimiter struct {
	mu       sync.Mutex
	max      int
	counters map[string]dailyCount
}

type dailyCount struct {
	count   int
	resetAt time.Time
}

func newDailyLimiter(max int) *dailyLimiter {
	return &dailyLimiter{max: max, counters: make(map[string]dailyCount)}
}

// allow increments the key's counter for the day and reports whether it stayed within the cap.
func (d *dailyLimiter) allow(key string, now time.Time) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	c, ok := d.counters[key]
	if !ok || !now.Before(c.resetAt) {
		d.counters[key] = dailyCount{count: 1, resetAt: nextMidnightUTC(now)}
		return true
	}
	if c.count >= d.max {
		return false
	}
	c.count++
	d.counters[key] = c
	return true
}

func nextMidnightUTC(now time.Time) time.Time {
	u := now.UTC()
	return time.Date(u.Year(), u.Month(), u.Day()+1, 0, 0, 0, 0, time.UTC)
}

// leaveRequestLimiter is the package-level daily limiter for self-service leave submissions.
var leaveRequestLimiter = newDailyLimiter(maxLeaveRequestsPerDay)

// leaveRequestCode is the 4-char uppercase-alphanumeric suffix matching the form's title format
// (e.g. the "V9VN" in "OOO-2026-nlk0211-V9VN").
func leaveRequestCode() string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Extremely unlikely; fall back to a time-derived code so a title is always produced.
		t := time.Now().UnixNano()
		for i := range b {
			b[i] = alphabet[int(t>>(uint(i)*8))%len(alphabet)]
		}
		return string(b)
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}

// generateLeaveRequestTitle builds a form-shaped, SPEC-087-parsable title
// "OOO-<year>-<discord>-<CODE>" for a bot-created row. The self-service create sets the Contractor
// relation up front, but unlike the form we do not rely on Notion's own auto-ID (which only fills the
// Discord slot after the relation resolves); we set a valid title directly so the request_id we hand
// back matches what the SPEC-087 approve/reject flow looks up.
func generateLeaveRequestTitle(discord string, start time.Time) string {
	handle := strings.TrimSpace(discord)
	if handle == "" {
		handle = "unknown"
	}
	return fmt.Sprintf("OOO-%d-%s-%s", start.Year(), handle, leaveRequestCode())
}

// HandleLeaveRequest handles POST /webhooks/discord/leave/request: a contractor filing their OWN
// leave through Neko Bot (SPEC-088 TASK-005).
func (h *handler) HandleLeaveRequest(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{"handler": "webhook", "method": "HandleLeaveRequest"})

	var req LeaveRequestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(req.RequesterHandle) == "" || strings.TrimSpace(req.StartDate) == "" ||
		strings.TrimSpace(req.EndDate) == "" || strings.TrimSpace(req.Type) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requester_handle, start_date, end_date and type are required"})
		return
	}

	// Validate the SHAPE first (cheap, no Notion): a malformed request is rejected before we touch the
	// resolver. Reordered ahead of the resolver vs the TASK-005 sketch; either order is sound (edge
	// case 3 notes end<start is caught before any fortress work) and this keeps 400s Notion-free.
	start, end, leaveType, code, msg := validateLeaveRequestParams(req.StartDate, req.EndDate, req.Type, defaultLeaveTypes, time.Now())
	if code != "" {
		l.Infof("HandleLeaveRequest: rejected params: handle=%s code=%s", normalizeHandle(req.RequesterHandle), code)
		c.JSON(http.StatusBadRequest, gin.H{"error": msg, "code": code})
		return
	}

	ctx := context.Background()
	leaveService := notionSvc.NewLeaveService(h.config, h.store, h.repo, h.logger)
	if leaveService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "notion service not configured"})
		return
	}

	// Resolve the AUTHENTICATED handle to an active contractor (page id + team email). 403 with no row
	// for an unknown handle or an inactive/departed contractor.
	candidates, err := leaveService.LookupContractorDetailsByDiscord(ctx, req.RequesterHandle)
	if err != nil {
		l.Errorf(err, "HandleLeaveRequest: contractor lookup failed: handle=%s", normalizeHandle(req.RequesterHandle))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up contractor"})
		return
	}
	contractor, ok := pickActiveContractor(candidates, req.RequesterHandle)
	if !ok {
		l.Infof("HandleLeaveRequest: refused non-active-contractor handle: %s", normalizeHandle(req.RequesterHandle))
		c.JSON(http.StatusForbidden, gin.H{"error": "I don't recognize you as an active contractor"})
		return
	}

	now := time.Now()

	// Idempotency (DEC-011): claim the (contractor,dates,type) key BEFORE creating. Of N concurrent
	// identical requests exactly one claims and proceeds; the rest 409 with no row.
	idemKey := leaveIdempotencyKey(contractor.PageID, start, end, leaveType)
	if !claimLeaveIdempotency(idemKey, now, leaveIdempotencyTTL) {
		l.Infof("HandleLeaveRequest: idempotency collision: contractor=%s", contractor.PageID)
		c.JSON(http.StatusConflict, gin.H{"error": "you already have that request pending"})
		return
	}

	// Daily rate limit, keyed per contractor. Release the claim if over limit so it is not wedged.
	if !leaveRequestLimiter.allow(contractor.PageID, now) {
		releaseLeaveIdempotency(idemKey)
		l.Infof("HandleLeaveRequest: daily limit exceeded: contractor=%s", contractor.PageID)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": fmt.Sprintf("daily leave-request limit reached (%d/day)", maxLeaveRequestsPerDay)})
		return
	}

	// Create the Status=New row (Contractor + Type + dates set, DEC-013). On failure release the claim
	// so a retry can succeed, and report no row (failure-modes table).
	title := generateLeaveRequestTitle(contractor.DiscordUsername, start)
	pageID, err := leaveService.CreateLeaveRequest(ctx, notionSvc.CreateLeaveInput{
		Title:              title,
		ContractorPageID:   contractor.PageID,
		UnavailabilityType: leaveType,
		StartDate:          start,
		EndDate:            end,
		Description:        truncateString(req.Reason, 1900), // Notion rich_text field limit is ~2000 (edge case 8: truncate, don't reject)
	})
	if err != nil {
		releaseLeaveIdempotency(idemKey)
		l.Errorf(err, "HandleLeaveRequest: create failed: contractor=%s", contractor.PageID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to file leave request"})
		return
	}

	// Row exists. Respond immediately with the request id, then announce best-effort (edge case 7: if
	// announce fails the row still exists and a lead sees it in Notion). Matches the gen-invoice /
	// leave-decision async-side-effect pattern.
	c.JSON(http.StatusOK, view.CreateResponse[any](gin.H{"request_id": title, "status": "New"}, nil, nil, nil, ""))

	leave := &notionSvc.LeaveRequest{
		PageID:             pageID,
		LeaveRequestTitle:  title,
		EmployeeID:         contractor.PageID,
		Email:              contractor.TeamEmail, // drives real AM/DL resolution, not the fallback
		UnavailabilityType: leaveType,
		StartDate:          &start,
		EndDate:            &end,
		Status:             "New",
		AdditionalContext:  req.Reason,
	}
	go h.announceNewLeaveRequest(context.Background(), l, leaveService, leave, contractor)
}

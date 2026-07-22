package webhook

// Leave approval backend (SPEC-087), shared by TWO front-ends that both route through the Hermes
// gateway (Neko Bot's interactions can't reach an HTTP endpoint without killing its slash commands):
//   - a Discord Approve/Reject BUTTON click -> gateway notion_leave_* handler (hermes patch 0010)
//   - a DM "approve <id>" -> the fortress MCP approve_leave tool
// Both authenticate the clicker/sender, mint an identity token, and call these endpoints. request_id
// is a title (DM) or a Notion page id (button); findPendingLeave accepts either.
//
// These endpoints add the authorization the old button flow never had: only an Account Manager or
// Delivery Lead on the requester's active deployment (or an admin) may decide a request. The old
// buttons recorded whoever clicked into "Reviewed By" but authorized nobody.
//
// Auth: unlike /webhooks/discord/gen-invoice (which has no HTTP auth and trusts its body), a leave
// decision is a state change keyed on a CLAIMED approver handle, so an unauthenticated endpoint
// would be a full bypass (POST any AM's handle -> approved). The routes are therefore gated by
// fortress's own API-key + permission middleware (see routes/v1.go); this handler trusts
// approver_handle only because that gate ensures the caller is the MCP behind the desk bot.

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	notionSvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// adminApproverEmails may decide any request even with no active deployment (empty AM/DL set),
// mirroring the notification's fallback assignees (SPEC-087 DEC-007).
var adminApproverEmails = []string{"han@d.foundation", "thanhpd@d.foundation"}

// LeaveDecisionRequest is the body the fortress MCP posts for approve/reject.
type LeaveDecisionRequest struct {
	ApproverHandle string `json:"approver_handle"` // the AUTHENTICATED lead handle (from the gateway token, not model text)
	RequestID      string `json:"request_id"`      // the leave request title, e.g. OOO-2026-minhth-YCZT
	Reason         string `json:"reason"`          // optional, reject only
}

// LeaveListRequest is the body for list_pending_leaves.
type LeaveListRequest struct {
	ApproverHandle string `json:"approver_handle"`
}

// normalizeHandle folds case and drops a legacy #discriminator so the caller's handle and the
// Notion "Discord" field compare equal even when the two sources disagree on format (DEC-006).
func normalizeHandle(h string) string {
	h = strings.TrimSpace(strings.ToLower(h))
	if i := strings.IndexByte(h, '#'); i >= 0 {
		h = h[:i]
	}
	return h
}

// resolveApproverHandles returns the normalized Discord handles allowed to decide this request:
// the AM/DL on the requester's active deployments, plus the admin allowlist. The same resolution
// the notification uses (getAMDLMentionsFromDeployments), but returning handles for a membership
// check rather than mentions.
func (h *handler) resolveApproverHandles(ctx context.Context, l logger.Logger, leaveService *notionSvc.LeaveService, leave notionSvc.LeaveRequest) map[string]bool {
	allowed := map[string]bool{}

	contractorPageID, err := leaveService.LookupContractorByEmail(ctx, leave.Email)
	if err == nil && contractorPageID != "" {
		deployments, err := leaveService.GetActiveDeploymentsForContractor(ctx, contractorPageID)
		if err != nil {
			l.Warnf("resolveApproverHandles: deployments lookup failed: contractor_id=%s: %v", contractorPageID, err)
		}
		seen := map[string]bool{}
		for _, dep := range deployments {
			for _, stakeholderID := range leaveService.ExtractStakeholdersFromDeployment(ctx, dep) {
				if stakeholderID == contractorPageID || seen[stakeholderID] {
					continue // the requester is not their own approver; dedup across deployments
				}
				seen[stakeholderID] = true
				username, err := leaveService.GetDiscordUsernameFromContractor(ctx, stakeholderID)
				if err != nil || username == "" {
					continue
				}
				allowed[normalizeHandle(username)] = true
			}
		}
	} else if err != nil {
		l.Warnf("resolveApproverHandles: contractor lookup failed: email=%s: %v", leave.Email, err)
	}

	// Admin allowlist: resolve the fallback-assignee emails to handles via the DB.
	for _, email := range adminApproverEmails {
		if handle := h.discordHandleForEmail(l, email); handle != "" {
			allowed[normalizeHandle(handle)] = true
		}
	}
	return allowed
}

// discordHandleForEmail resolves an employee email to their Discord username via the local mirror.
func (h *handler) discordHandleForEmail(l logger.Logger, email string) string {
	db := h.repo.DB()
	employee, err := h.store.Employee.OneByEmail(db, email)
	if err != nil || employee == nil || employee.DiscordAccountID.String() == "" {
		return ""
	}
	acc, err := h.store.DiscordAccount.One(db, employee.DiscordAccountID.String())
	if err != nil || acc == nil {
		return ""
	}
	return acc.DiscordUsername
}

// findPendingLeave resolves a request id to the pending LeaveRequest. The id is either the
// human title (from the DM tool, e.g. OOO-2026-innno_-HGU4) or the Notion page id (from a button's
// custom_id). Both front-ends converge here. Returns ok=false when no pending request matches.
func (h *handler) findPendingLeave(ctx context.Context, leaveService *notionSvc.LeaveService, requestID string) (notionSvc.LeaveRequest, bool, error) {
	want := strings.TrimSpace(requestID)

	// A page id (button path) resolves directly, no list query needed. Only a still-New request is
	// actionable; an already-decided one is treated as "not pending" so approve/reject 404s cleanly.
	if looksLikeNotionPageID(want) {
		lr, err := leaveService.GetLeaveRequest(ctx, want)
		if err != nil || lr == nil {
			return notionSvc.LeaveRequest{}, false, nil
		}
		// Scope check: GetLeaveRequest fetches ANY page the integration can read, so confirm this
		// page is actually a leave request (its title has the OOO-/LVR- shape) before we act on it.
		// Otherwise a 32-hex request_id could point a decision at an unrelated page.
		if !isLeaveRequestTitle(lr.LeaveRequestTitle) {
			return notionSvc.LeaveRequest{}, false, nil
		}
		if isLeaveAlreadyDecided(lr.Status) {
			return notionSvc.LeaveRequest{}, false, nil
		}
		return *lr, true, nil
	}

	// A title (DM path) needs the pending list to match against.
	pending, err := leaveService.QueryPendingLeaveRequests(ctx)
	if err != nil {
		return notionSvc.LeaveRequest{}, false, err
	}
	for _, lr := range pending {
		if leaveTitleMatches(lr.LeaveRequestTitle, requestID) {
			return lr, true, nil
		}
	}
	return notionSvc.LeaveRequest{}, false, nil
}

// looksLikeNotionPageID reports whether s is a Notion page id (32 hex chars, dashed or not) rather
// than a human leave title.
func looksLikeNotionPageID(s string) bool {
	stripped := strings.ReplaceAll(s, "-", "")
	if len(stripped) != 32 {
		return false
	}
	for _, r := range stripped {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

// leaveTitleMatches compares a candidate leave title against a requested id, tolerant of
// surrounding whitespace and case (the model relays the title as the lead typed it).
func leaveTitleMatches(candidate, requestID string) bool {
	return strings.EqualFold(strings.TrimSpace(candidate), strings.TrimSpace(requestID))
}

// isLeaveRequestTitle reports whether a title has the leave-request shape (OOO-... or LVR-...),
// used to confirm a page resolved by id is actually a leave request and not an arbitrary page.
func isLeaveRequestTitle(title string) bool {
	t := strings.ToUpper(strings.TrimSpace(title))
	return strings.HasPrefix(t, "OOO-") || strings.HasPrefix(t, "LVR-")
}

// isLeaveAlreadyDecided reports whether a leave request has already been acted on, so a repeat or
// concurrent decision does not re-run the calendar side effect (SPEC-087 edge case 2).
func isLeaveAlreadyDecided(status string) bool {
	switch status {
	case "Acknowledged", "Not Applicable", "Withdrawn":
		return true
	default:
		return false
	}
}

// leaveDecisionLocks serializes decisions per leave page id. The idempotency guard is a
// read-status-then-write, which without a lock has a TOCTOU window: two near-simultaneous
// approvals of the same request (a lead double-tapping the button) could both read Status=New
// before either writes, and both create a calendar event. Locking per page id closes that window.
// The api deployment is a single replica, so an in-process lock is sufficient; a multi-replica
// deployment would additionally need a Notion conditional update (write only if Status is New).
var leaveDecisionLocks sync.Map // pageID -> *sync.Mutex

func lockLeaveDecision(pageID string) func() {
	m, _ := leaveDecisionLocks.LoadOrStore(pageID, &sync.Mutex{})
	mu := m.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// applyLeaveDecision is the side-effect core shared by the endpoints (the button handlers keep
// their own inline copy for now). "approve" -> Status Acknowledged + a calendar event; "reject" ->
// Status Not Applicable, no calendar. Idempotent: if the request is already decided, it does NOT
// re-create the calendar event, so two near-simultaneous approvals cannot double-book (edge case 2).
func (h *handler) applyLeaveDecision(ctx context.Context, l logger.Logger, leaveService *notionSvc.LeaveService, pageID, approverPageID, decision string) {
	unlock := lockLeaveDecision(pageID)
	defer unlock()

	// Idempotency guard: read current status first; skip the calendar side effect if already decided.
	alreadyDecided := false
	if cur, err := leaveService.GetLeaveRequest(ctx, pageID); err == nil && cur != nil {
		alreadyDecided = isLeaveAlreadyDecided(cur.Status)
	}

	status := "Acknowledged"
	if decision == "reject" {
		status = "Not Applicable"
	}
	if err := leaveService.UpdateLeaveStatus(ctx, pageID, status, approverPageID); err != nil {
		l.Errorf(err, "applyLeaveDecision: UpdateLeaveStatus failed: page_id=%s decision=%s", pageID, decision)
	}
	if decision == "approve" && !alreadyDecided {
		if err := h.createCalendarEventForLeave(ctx, l, leaveService, pageID); err != nil {
			l.Errorf(err, "applyLeaveDecision: calendar event failed: page_id=%s", pageID)
		}
	}
}

// HandleLeaveList handles POST /webhooks/discord/leave/list: return the pending requests THIS
// caller may decide. ponytail: resolves approvers per pending request (N small Notion queries);
// pending count is a handful, upgrade to a batched rollup only if that stops being true.
func (h *handler) HandleLeaveList(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{"handler": "webhook", "method": "HandleLeaveList"})
	var req LeaveListRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ApproverHandle == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "approver_handle required"})
		return
	}
	ctx := context.Background()
	leaveService := notionSvc.NewLeaveService(h.config, h.store, h.repo, h.logger)
	if leaveService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "notion service not configured"})
		return
	}
	pending, err := leaveService.QueryPendingLeaveRequests(ctx)
	if err != nil {
		l.Errorf(err, "HandleLeaveList: query pending failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query leave requests"})
		return
	}
	caller := normalizeHandle(req.ApproverHandle)
	type item struct {
		RequestID string `json:"request_id"`
		Type      string `json:"type"`
		Start     string `json:"start_date"`
		End       string `json:"end_date"`
	}
	out := []item{}
	for _, lr := range pending {
		if !h.resolveApproverHandles(ctx, l, leaveService, lr)[caller] {
			continue
		}
		it := item{RequestID: lr.LeaveRequestTitle, Type: lr.UnavailabilityType}
		if lr.StartDate != nil {
			it.Start = lr.StartDate.Format("2006-01-02")
		}
		if lr.EndDate != nil {
			it.End = lr.EndDate.Format("2006-01-02")
		}
		out = append(out, it)
	}
	c.JSON(http.StatusOK, gin.H{"requests": out})
}

// HandleLeaveApprove handles POST /webhooks/discord/leave/approve.
func (h *handler) HandleLeaveApprove(c *gin.Context) { h.handleLeaveDecision(c, "approve") }

// HandleLeaveReject handles POST /webhooks/discord/leave/reject.
func (h *handler) HandleLeaveReject(c *gin.Context) { h.handleLeaveDecision(c, "reject") }

func (h *handler) handleLeaveDecision(c *gin.Context, decision string) {
	l := h.logger.Fields(logger.Fields{"handler": "webhook", "method": "handleLeaveDecision", "decision": decision})
	var req LeaveDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ApproverHandle == "" || req.RequestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "approver_handle and request_id required"})
		return
	}
	ctx := context.Background()
	leaveService := notionSvc.NewLeaveService(h.config, h.store, h.repo, h.logger)
	if leaveService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "notion service not configured"})
		return
	}

	leave, ok, err := h.findPendingLeave(ctx, leaveService, req.RequestID)
	if err != nil {
		l.Errorf(err, "handleLeaveDecision: lookup failed: request=%s", req.RequestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up leave request"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("no pending leave request: %s", req.RequestID)})
		return
	}

	// Authorization: the caller must be an AM/DL on the requester's active deployment (or an admin).
	allowed := h.resolveApproverHandles(ctx, l, leaveService, leave)
	caller := normalizeHandle(req.ApproverHandle)
	if !allowed[caller] {
		l.Infof("handleLeaveDecision: refused non-approver: caller=%s request=%s", caller, req.RequestID)
		c.JSON(http.StatusForbidden, gin.H{"error": "not an approver (Account Manager or Delivery Lead) for this request"})
		return
	}

	// The approver's own contractor page (for "Reviewed By"); empty is acceptable (as in the button
	// flow) for an admin who is not a deployment stakeholder.
	approverPageID := ""
	if pid, err := leaveService.LookupContractorByEmail(ctx, h.emailForDiscordHandle(l, req.ApproverHandle)); err == nil {
		approverPageID = pid
	}

	// Return immediately; apply the side effects async, matching the gen-invoice pattern.
	c.JSON(http.StatusOK, view.CreateResponse[any](gin.H{"request_id": req.RequestID, "decision": decision}, nil, nil, nil, ""))
	go h.applyLeaveDecision(context.Background(), l, leaveService, leave.PageID, approverPageID, decision)
}

// emailForDiscordHandle resolves a Discord handle to the employee's team email via the local mirror
// (the inverse of discordHandleForEmail). Used to find the approver's contractor page.
func (h *handler) emailForDiscordHandle(l logger.Logger, handle string) string {
	db := h.repo.DB()
	acc, err := h.store.DiscordAccount.OneByUsername(db, strings.TrimSpace(handle))
	if err != nil || acc == nil {
		return ""
	}
	employee, err := h.store.Employee.GetByDiscordID(db, acc.DiscordID, false)
	if err != nil || employee == nil {
		return ""
	}
	return employee.TeamEmail
}

package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	notionSvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
)

// fixedNow anchors the backdate/horizon bounds so the date-window tests are deterministic.
var fixedNow = time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)

// validateLeaveRequestParams is the 400 gate: it must accept a well-formed request, reject malformed
// dates / an inverted range / a type outside the option set, allow a bounded backdate (form parity,
// DEC-012), and reject anything backdated past the window or beyond the horizon.
func TestValidateLeaveRequestParams(t *testing.T) {
	cases := []struct {
		name                  string
		start, end, leaveType string
		wantCode              string
		wantCanonical         string
	}{
		{"valid", "2026-08-04", "2026-08-06", "Personal Time", "", "Personal Time"},
		{"valid_type_case_insensitive", "2026-08-04", "2026-08-06", "personal time", "", "Personal Time"},
		{"valid_single_day", "2026-07-22", "2026-07-22", "Health / Illness", "", "Health / Illness"},
		{"backdate_within_window", "2026-07-15", "2026-07-16", "Health / Illness", "", "Health / Illness"}, // 7 days back, allowed
		{"bad_start_date", "2026-13-04", "2026-08-06", "Personal Time", "invalid_start_date", ""},
		{"bad_start_not_a_date", "next tuesday", "2026-08-06", "Personal Time", "invalid_start_date", ""},
		{"bad_end_date", "2026-08-04", "nope", "Personal Time", "invalid_end_date", ""},
		{"end_before_start", "2026-08-06", "2026-08-04", "Personal Time", "end_before_start", ""},
		{"backdated_past_window", "2026-07-01", "2026-07-02", "Personal Time", "backdated", ""}, // 21 days back
		{"beyond_horizon", "2027-08-04", "2027-08-06", "Personal Time", "too_far", ""},          // >365 days out
		{"invalid_type", "2026-08-04", "2026-08-06", "Sabbatical", "invalid_type", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, canonical, code, msg := validateLeaveRequestParams(c.start, c.end, c.leaveType, defaultLeaveTypes, fixedNow)
			if code != c.wantCode {
				t.Fatalf("code = %q (msg=%q), want %q", code, msg, c.wantCode)
			}
			if canonical != c.wantCanonical {
				t.Fatalf("canonical type = %q, want %q", canonical, c.wantCanonical)
			}
		})
	}
}

func TestCanonicalLeaveType(t *testing.T) {
	got, ok := canonicalLeaveType("  TRAVEL / VACATION ", defaultLeaveTypes)
	if !ok || got != "Travel / Vacation" {
		t.Fatalf("canonicalLeaveType folded/trimmed match = (%q,%v), want (Travel / Vacation,true)", got, ok)
	}
	if _, ok := canonicalLeaveType("unknown", defaultLeaveTypes); ok {
		t.Fatal("an unknown type must not match")
	}
}

// pickActiveContractor is the resolver's decision core: an exact normalized handle match on an ACTIVE
// contractor resolves; an inactive contractor, a different handle, or an empty set do not (403).
func TestPickActiveContractor(t *testing.T) {
	active := &notionSvc.ContractorDetails{PageID: "page-1", DiscordUsername: "nlk0211", TeamEmail: "quang@d.foundation", Status: "Active"}
	inactive := &notionSvc.ContractorDetails{PageID: "page-2", DiscordUsername: "gone", Status: "Inactive"}
	other := &notionSvc.ContractorDetails{PageID: "page-3", DiscordUsername: "someoneelse", Status: "Active"}

	t.Run("active_match", func(t *testing.T) {
		got, ok := pickActiveContractor([]*notionSvc.ContractorDetails{other, active}, "nlk0211")
		if !ok || got.PageID != "page-1" {
			t.Fatalf("expected page-1 active match, got (%v,%v)", got, ok)
		}
	})
	t.Run("case_and_discriminator_tolerant", func(t *testing.T) {
		got, ok := pickActiveContractor([]*notionSvc.ContractorDetails{active}, "NLK0211#1234")
		if !ok || got.PageID != "page-1" {
			t.Fatalf("normalized handle should still match active contractor, got (%v,%v)", got, ok)
		}
	})
	t.Run("inactive_not_found", func(t *testing.T) {
		if _, ok := pickActiveContractor([]*notionSvc.ContractorDetails{inactive}, "gone"); ok {
			t.Fatal("an inactive contractor must resolve to not-found (403)")
		}
	})
	t.Run("unknown_handle_not_found", func(t *testing.T) {
		if _, ok := pickActiveContractor([]*notionSvc.ContractorDetails{active}, "stranger"); ok {
			t.Fatal("a non-matching handle must resolve to not-found (403)")
		}
	})
	t.Run("empty_not_found", func(t *testing.T) {
		if _, ok := pickActiveContractor(nil, "nlk0211"); ok {
			t.Fatal("no candidates must resolve to not-found (403)")
		}
	})
}

func TestLeaveIdempotencyKey(t *testing.T) {
	s := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	e := time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC)
	k1 := leaveIdempotencyKey("page-1", s, e, "Personal Time")
	k2 := leaveIdempotencyKey("page-1", s, e, "personal time") // type folded -> same key
	if k1 != k2 {
		t.Fatal("type case must not change the idempotency key")
	}
	if k1 == leaveIdempotencyKey("page-2", s, e, "Personal Time") {
		t.Fatal("a different contractor must produce a different key")
	}
	if k1 == leaveIdempotencyKey("page-1", s, e, "Travel / Vacation") {
		t.Fatal("a different type must produce a different key")
	}
	if k1 == leaveIdempotencyKey("page-1", s, e.AddDate(0, 0, 1), "Personal Time") {
		t.Fatal("a different end date must produce a different key")
	}
}

// The idempotency claim is the race-safe backstop (DEC-011): of N concurrent claims of the SAME key,
// EXACTLY ONE wins (creates the row); the rest 409. This is the "two concurrent identical -> one row"
// acceptance criterion at the mechanism level.
func TestClaimLeaveIdempotency_Concurrent(t *testing.T) {
	key := leaveIdempotencyKey("concurrent-contractor", fixedNow, fixedNow, "Personal Time")
	defer releaseLeaveIdempotency(key)

	const n = 64
	var winners int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if claimLeaveIdempotency(key, fixedNow, leaveIdempotencyTTL) {
				atomic.AddInt32(&winners, 1)
			}
		}()
	}
	close(start)
	wg.Wait()

	if winners != 1 {
		t.Fatalf("exactly one concurrent claimer must win, got %d", winners)
	}
	// A replay while the claim is live still collides (409).
	if claimLeaveIdempotency(key, fixedNow, leaveIdempotencyTTL) {
		t.Fatal("a replay against a live claim must collide")
	}
	// After release, the key is free again (a genuine retry can proceed).
	releaseLeaveIdempotency(key)
	if !claimLeaveIdempotency(key, fixedNow, leaveIdempotencyTTL) {
		t.Fatal("after release the key must be claimable again")
	}
}

// An expired claim must be reclaimable (the TTL is the replay window, not a permanent lock).
func TestClaimLeaveIdempotency_Expiry(t *testing.T) {
	key := leaveIdempotencyKey("expiry-contractor", fixedNow, fixedNow, "Other")
	defer releaseLeaveIdempotency(key)
	if !claimLeaveIdempotency(key, fixedNow, time.Minute) {
		t.Fatal("first claim should win")
	}
	if claimLeaveIdempotency(key, fixedNow.Add(30*time.Second), time.Minute) {
		t.Fatal("a claim inside the TTL must collide")
	}
	if !claimLeaveIdempotency(key, fixedNow.Add(2*time.Minute), time.Minute) {
		t.Fatal("a claim after the TTL must succeed")
	}
}

// The daily limiter allows up to the cap and blocks beyond it, resetting after UTC midnight.
func TestDailyLimiter(t *testing.T) {
	d := newDailyLimiter(3)
	for i := 1; i <= 3; i++ {
		if !d.allow("c1", fixedNow) {
			t.Fatalf("request %d should be within the daily cap", i)
		}
	}
	if d.allow("c1", fixedNow) {
		t.Fatal("the 4th request in a day must be blocked")
	}
	// A different contractor has its own budget.
	if !d.allow("c2", fixedNow) {
		t.Fatal("a different contractor must not share the counter")
	}
	// Next day resets.
	if !d.allow("c1", fixedNow.AddDate(0, 0, 1)) {
		t.Fatal("the counter must reset the next day")
	}
}

// The bot-created title must be a valid SPEC-087 leave request id so the approve/reject lookup can
// resolve it.
func TestGenerateLeaveRequestTitle(t *testing.T) {
	title := generateLeaveRequestTitle("nlk0211", time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC))
	if !isValidLeaveRequestID(title) {
		t.Fatalf("generated title %q is not a valid leave request id", title)
	}
	if got := title[:9]; got != "OOO-2026-" {
		t.Fatalf("title prefix = %q, want OOO-2026-", got)
	}
}

// Regression for TASK-004 DEC-010: while the integration user id is unset the skip-guard is INERT, so
// the Notion-automation (form) path behaves exactly as before; once the id is set it skips only rows
// created by that exact user. The const is empty until the TASK-001 spike, so the automation path is
// unchanged today.
func TestShouldSkipIntegrationCreatedRow(t *testing.T) {
	if fortressIntegrationUserID == "" {
		if shouldSkipIntegrationCreatedRow("") || shouldSkipIntegrationCreatedRow("any-user-id") {
			t.Fatal("skip-guard must be inert while fortressIntegrationUserID is empty (form path unchanged)")
		}
	} else {
		if !shouldSkipIntegrationCreatedRow(fortressIntegrationUserID) {
			t.Fatal("a row created by the integration user must be skipped")
		}
		if shouldSkipIntegrationCreatedRow("some-contractor-user") {
			t.Fatal("a row created by a real contractor must NOT be skipped")
		}
	}
}

// The HTTP handler rejects a malformed body, missing required fields, and bad params with a 400 and
// NEVER touches Notion (these all return before the leave service is created).
func TestHandleLeaveRequest_BadRequests(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	h := &handler{logger: l}

	post := func(body []byte) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/webhooks/discord/leave/request", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.HandleLeaveRequest(c)
		return w
	}

	t.Run("invalid_json", func(t *testing.T) {
		w := post([]byte("not json"))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	missing := []map[string]string{
		{"start_date": "2026-08-04", "end_date": "2026-08-06", "type": "Personal Time"}, // no handle
		{"requester_handle": "nlk0211", "end_date": "2026-08-06", "type": "Personal Time"},
		{"requester_handle": "nlk0211", "start_date": "2026-08-04", "type": "Personal Time"},
		{"requester_handle": "nlk0211", "start_date": "2026-08-04", "end_date": "2026-08-06"},
	}
	for i, m := range missing {
		t.Run("missing_field", func(t *testing.T) {
			b, _ := json.Marshal(m)
			w := post(b)
			assert.Equalf(t, http.StatusBadRequest, w.Code, "case %d", i)
		})
	}

	// Clock-independent 400s (the horizon/backdate windows are covered deterministically in the pure
	// validateLeaveRequestParams test, which injects a fixed now).
	bad := []map[string]string{
		{"requester_handle": "nlk0211", "start_date": "2026-08-06", "end_date": "2026-08-04", "type": "Personal Time"}, // end<start
		{"requester_handle": "nlk0211", "start_date": "nope", "end_date": "2026-08-06", "type": "Personal Time"},       // bad date
	}
	for i, m := range bad {
		t.Run("bad_params", func(t *testing.T) {
			b, _ := json.Marshal(m)
			w := post(b)
			assert.Equalf(t, http.StatusBadRequest, w.Code, "case %d", i)
		})
	}
}

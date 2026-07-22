package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

// canonicalLeaveType must EXACT-match (case/space tolerant) against the whole option name: every
// allowed option round-trips to itself, and a substring of an option (e.g. "Personal" for
// "Personal Time") must NOT match, or a contractor could smuggle a non-option value past the gate.
func TestCanonicalLeaveType_Exhaustive(t *testing.T) {
	for _, want := range defaultLeaveTypes {
		if got, ok := canonicalLeaveType(want, defaultLeaveTypes); !ok || got != want {
			t.Fatalf("allowed type %q must map to itself, got (%q,%v)", want, got, ok)
		}
		if got, ok := canonicalLeaveType(strings.ToUpper(want), defaultLeaveTypes); !ok || got != want {
			t.Fatalf("uppercased %q must map to the canonical %q, got (%q,%v)", want, want, got, ok)
		}
	}
	for _, bad := range []string{"", "  ", "Personal", "Time", "Health", "Vacation", "personal  time"} {
		if got, ok := canonicalLeaveType(bad, defaultLeaveTypes); ok {
			t.Fatalf("non-option %q must not match (got %q); only whole-name matches are valid", bad, got)
		}
	}
}

// Boundary coverage for the DEC-012 backdate/horizon windows: the inclusive edges (exactly
// maxLeaveBackdateDays back, exactly maxLeaveHorizonDays out, and today itself) are ACCEPTED, while
// one day past either edge is rejected. Also asserts surrounding whitespace on the dates is trimmed.
func TestValidateLeaveRequestParams_Boundaries(t *testing.T) {
	day := func(y, m, d int) string {
		return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	}
	cases := []struct {
		name, start, end, wantCode string
	}{
		{"today_allowed", day(2026, 7, 22), day(2026, 7, 22), ""},
		{"backdate_exact_edge_allowed", day(2026, 7, 8), day(2026, 7, 22), ""},             // exactly 14 days back
		{"backdate_one_past_edge_rejected", day(2026, 7, 7), day(2026, 7, 8), "backdated"}, // 15 days back
		{"horizon_exact_edge_allowed", day(2027, 7, 22), day(2027, 7, 22), ""},             // exactly +365
		{"horizon_one_past_edge_rejected", day(2027, 7, 23), day(2027, 7, 23), "too_far"},  // +366
		{"whitespace_trimmed", "  2026-08-04 ", " 2026-08-06  ", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, _, code, msg := validateLeaveRequestParams(c.start, c.end, "Personal Time", defaultLeaveTypes, fixedNow)
			if code != c.wantCode {
				t.Fatalf("code = %q (msg=%q), want %q", code, msg, c.wantCode)
			}
		})
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

// Additional resolver edge cases: a nil slot in the over-fetched candidates is skipped (not a panic);
// a status with surrounding whitespace is still "Active"; a non-Active status other than "Inactive"
// (e.g. "Probation") is excluded; the stored Discord handle is normalized on the CANDIDATE side too
// (case + #discriminator); an empty request handle matches nothing; and among several active
// namesakes the first is returned deterministically.
func TestPickActiveContractor_EdgeCases(t *testing.T) {
	t.Run("nil_entry_skipped", func(t *testing.T) {
		active := &notionSvc.ContractorDetails{PageID: "p1", DiscordUsername: "nlk0211", Status: "Active"}
		got, ok := pickActiveContractor([]*notionSvc.ContractorDetails{nil, active}, "nlk0211")
		if !ok || got.PageID != "p1" {
			t.Fatalf("a nil candidate must be skipped, not fatal; got (%v,%v)", got, ok)
		}
	})
	t.Run("status_whitespace_still_active", func(t *testing.T) {
		c := &notionSvc.ContractorDetails{PageID: "p1", DiscordUsername: "nlk0211", Status: " Active "}
		if _, ok := pickActiveContractor([]*notionSvc.ContractorDetails{c}, "nlk0211"); !ok {
			t.Fatal("a status of ' Active ' must count as Active after trim")
		}
	})
	t.Run("probation_excluded", func(t *testing.T) {
		c := &notionSvc.ContractorDetails{PageID: "p1", DiscordUsername: "nlk0211", Status: "Probation"}
		if _, ok := pickActiveContractor([]*notionSvc.ContractorDetails{c}, "nlk0211"); ok {
			t.Fatal("any non-Active status (Probation) must resolve to not-found (403)")
		}
	})
	t.Run("stored_handle_case_and_discriminator_normalized", func(t *testing.T) {
		c := &notionSvc.ContractorDetails{PageID: "p1", DiscordUsername: "NLK0211#1234", Status: "Active"}
		if _, ok := pickActiveContractor([]*notionSvc.ContractorDetails{c}, "nlk0211"); !ok {
			t.Fatal("the stored handle must be normalized (case + #discriminator) before comparison")
		}
	})
	t.Run("empty_request_handle_matches_nothing", func(t *testing.T) {
		c := &notionSvc.ContractorDetails{PageID: "p1", DiscordUsername: "nlk0211", Status: "Active"}
		if _, ok := pickActiveContractor([]*notionSvc.ContractorDetails{c}, "   "); ok {
			t.Fatal("a blank requester handle must not resolve to any contractor")
		}
	})
	t.Run("first_of_several_actives", func(t *testing.T) {
		a := &notionSvc.ContractorDetails{PageID: "first", DiscordUsername: "dup", Status: "Active"}
		b := &notionSvc.ContractorDetails{PageID: "second", DiscordUsername: "dup", Status: "Active"}
		got, ok := pickActiveContractor([]*notionSvc.ContractorDetails{a, b}, "dup")
		if !ok || got.PageID != "first" {
			t.Fatalf("among active namesakes the first candidate must win, got (%v,%v)", got, ok)
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

// The idempotency key must be a stable, deterministic function of its inputs (same inputs -> byte-
// identical key across calls), fold surrounding whitespace on the type, and change when the START
// date changes (the existing test only varied the end date).
func TestLeaveIdempotencyKey_Determinism(t *testing.T) {
	s := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	e := time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC)
	if leaveIdempotencyKey("page-1", s, e, "Personal Time") != leaveIdempotencyKey("page-1", s, e, "Personal Time") {
		t.Fatal("identical inputs must produce an identical key")
	}
	if leaveIdempotencyKey("page-1", s, e, "Personal Time") != leaveIdempotencyKey("page-1", s, e, "  personal time  ") {
		t.Fatal("surrounding whitespace and case on the type must fold to the same key")
	}
	if leaveIdempotencyKey("page-1", s, e, "Personal Time") == leaveIdempotencyKey("page-1", s.AddDate(0, 0, 1), e, "Personal Time") {
		t.Fatal("a different start date must produce a different key")
	}
	// A stable 64-hex-char SHA-256 digest (guards against an accidental format change).
	if k := leaveIdempotencyKey("page-1", s, e, "Personal Time"); len(k) != 64 {
		t.Fatalf("idempotency key must be a 64-char hex sha256 digest, got len %d", len(k))
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

// The daily limiter is hit concurrently (mutex-guarded), so of N simultaneous submissions for one
// contractor EXACTLY cap are allowed and the rest are blocked, with no lost updates. Run with -race
// this also guards the counter map against a data race.
func TestDailyLimiter_Concurrent(t *testing.T) {
	const cap, n = 3, 64
	d := newDailyLimiter(cap)
	var allowed int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if d.allow("busy-contractor", fixedNow) {
				atomic.AddInt32(&allowed, 1)
			}
		}()
	}
	close(start)
	wg.Wait()
	if allowed != cap {
		t.Fatalf("exactly %d concurrent submissions must be allowed, got %d", cap, allowed)
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

// An empty Discord handle must still yield a well-formed (parsable) request id via the "unknown"
// slot, so the create never emits a title the SPEC-087 lookup would reject; the random suffix is a
// 4-char uppercase-alphanumeric code, and successive calls vary it.
func TestGenerateLeaveRequestTitle_Edge(t *testing.T) {
	when := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	title := generateLeaveRequestTitle("   ", when)
	if !isValidLeaveRequestID(title) {
		t.Fatalf("empty-handle title %q must still be a valid leave request id", title)
	}
	parts := strings.Split(title, "-")
	if parts[2] != "unknown" {
		t.Fatalf("blank handle must fall back to the 'unknown' slot, got %q", parts[2])
	}
	code := parts[3]
	if len(code) != 4 {
		t.Fatalf("the request code must be 4 chars, got %q", code)
	}
	for _, r := range code {
		if !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			t.Fatalf("the request code must be uppercase alphanumeric, got %q", code)
		}
	}
	// The suffix is random: over several draws we expect at least two distinct codes (flake-proof:
	// collision of 8 draws over a 36^4 space is astronomically unlikely).
	seen := map[string]bool{}
	for i := 0; i < 8; i++ {
		seen[strings.Split(generateLeaveRequestTitle("nlk0211", when), "-")[3]] = true
	}
	if len(seen) < 2 {
		t.Fatalf("expected the random code to vary across draws, saw only %v", seen)
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

// The handler rejects an out-of-set type and whitespace-only required fields with a 400 BEFORE the
// Notion service is created (shape validation precedes the resolver). Edge case 1 (bad type) and the
// required-field guard exercised through the real HTTP entrypoint, not just the pure validator.
func TestHandleLeaveRequest_InvalidTypeAndWhitespace(t *testing.T) {
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

	t.Run("type_not_in_option_set", func(t *testing.T) {
		b, _ := json.Marshal(map[string]string{
			"requester_handle": "nlk0211", "start_date": "2026-08-04", "end_date": "2026-08-06", "type": "Sabbatical",
		})
		w := post(b)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_type")
	})

	// A field that is present but only whitespace is treated as missing (required-field guard), so it
	// never reaches Notion as an empty create.
	whitespace := []map[string]string{
		{"requester_handle": "   ", "start_date": "2026-08-04", "end_date": "2026-08-06", "type": "Personal Time"},
		{"requester_handle": "nlk0211", "start_date": "  ", "end_date": "2026-08-06", "type": "Personal Time"},
		{"requester_handle": "nlk0211", "start_date": "2026-08-04", "end_date": " ", "type": "Personal Time"},
		{"requester_handle": "nlk0211", "start_date": "2026-08-04", "end_date": "2026-08-06", "type": "  "},
	}
	for i, m := range whitespace {
		t.Run("whitespace_only_field", func(t *testing.T) {
			b, _ := json.Marshal(m)
			w := post(b)
			assert.Equalf(t, http.StatusBadRequest, w.Code, "case %d", i)
		})
	}
}

package webhook

import "testing"

// normalizeHandle is the security-critical comparison for leave approval: the caller's handle
// (from the gateway token) is matched against the AM/DL set (from the Notion "Discord" field). A
// format disagreement between the two sources must NOT silently refuse a real lead, and must NOT
// let a near-miss through. (SPEC-087 DEC-006.)
func TestNormalizeHandle(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"innno_", "innno_"},
		{"Innno_", "innno_"},      // case fold
		{"  innno_  ", "innno_"},  // trim
		{"innno_#1234", "innno_"}, // legacy discriminator stripped
		{"Innno_#0001", "innno_"}, // case + discriminator
		{"HAN.d", "han.d"},        // dotted new-style handle
		{"", ""},                  // empty stays empty (never matches a real handle)
	}
	for _, c := range cases {
		if got := normalizeHandle(c.in); got != c.want {
			t.Errorf("normalizeHandle(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// Two handles that differ only by case/discriminator must compare EQUAL after normalization
// (a real lead is not refused), and two genuinely different handles must NOT (an impostor is
// refused).
func TestNormalizeHandle_MembershipSemantics(t *testing.T) {
	allowed := map[string]bool{normalizeHandle("Innno_#1234"): true}
	if !allowed[normalizeHandle("innno_")] {
		t.Fatal("a real lead whose Notion handle carries a discriminator must still match")
	}
	if allowed[normalizeHandle("innno")] {
		t.Fatal("a different handle (innno vs innno_) must NOT match")
	}
}

// leaveTitleMatches resolves the request id (the title the lead typed) to a pending request,
// tolerant of case and surrounding whitespace but not of a genuinely different id.
func TestLeaveTitleMatches(t *testing.T) {
	cases := []struct {
		candidate, requestID string
		want                 bool
	}{
		{"OOO-2026-innno_-HGU4", "OOO-2026-innno_-HGU4", true},
		{"OOO-2026-innno_-HGU4", "  OOO-2026-innno_-HGU4  ", true}, // trimmed
		{"OOO-2026-innno_-HGU4", "ooo-2026-innno_-hgu4", true},     // case-fold
		{"OOO-2026-innno_-HGU4", "OOO-2026-innno_-XXXX", false},    // different code
		{"OOO-2026-innno_-HGU4", "OOO-2026-minhth-HGU4", false},    // different person
		{"", "", true}, // degenerate; guarded upstream by required-field checks
	}
	for _, c := range cases {
		if got := leaveTitleMatches(c.candidate, c.requestID); got != c.want {
			t.Errorf("leaveTitleMatches(%q,%q) = %v, want %v", c.candidate, c.requestID, got, c.want)
		}
	}
}

// isLeaveAlreadyDecided drives the idempotency guard: only a still-open (New/pending) request
// gets a fresh calendar event; anything already decided must not double-book.
func TestIsLeaveAlreadyDecided(t *testing.T) {
	decided := []string{"Acknowledged", "Not Applicable", "Withdrawn"}
	open := []string{"New", "", "Pending", "in review"}
	for _, s := range decided {
		if !isLeaveAlreadyDecided(s) {
			t.Errorf("status %q should count as already decided", s)
		}
	}
	for _, s := range open {
		if isLeaveAlreadyDecided(s) {
			t.Errorf("status %q should NOT count as already decided", s)
		}
	}
}

// A button carries the Notion page id (32 hex, dashed or not); a DM carries the human title.
// findPendingLeave routes on this: a page id resolves directly, a title goes through the list.
func TestLooksLikeNotionPageID(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"3a564b29-b84c-81a6-bd55-fa785e0c3b1d", true}, // dashed uuid
		{"3a564b29b84c81a6bd55fa785e0c3b1d", true},     // undashed
		{"OOO-2026-innno_-HGU4", false},                // a title
		{"", false},
		{"3a564b29-b84c-81a6-bd55-fa785e0c3b1", false}, // 31 hex, too short
		{"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", false},    // 32 non-hex
	}
	for _, c := range cases {
		if got := looksLikeNotionPageID(c.in); got != c.want {
			t.Errorf("looksLikeNotionPageID(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

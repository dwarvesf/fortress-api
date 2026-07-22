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

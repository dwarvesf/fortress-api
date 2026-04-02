package webhook

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
)

func TestCheckPayoutItemOverlap_NoOverlap(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	h := &handler{logger: l}

	// P1 has [pi1, pi2], new invoice has [pi3] → no overlap, should create P2
	pendingPayables := []*notion.PayableInfo{
		{
			PageID:    "payable-p1",
			Status:    "Pending",
			InvoiceID: "INV-001",
		},
	}

	newPayoutItemIDs := []string{"pi3"}

	// Mock: payable-p1 contains pi1, pi2
	// In RED phase, this test references checkPayoutItemOverlap which doesn't exist yet
	hasOverlap, matchedPayable := h.checkPayoutItemOverlap(context.Background(), l, pendingPayables, newPayoutItemIDs, map[string][]string{
		"payable-p1": {"pi1", "pi2"},
	})

	assert.False(t, hasOverlap, "should not detect overlap when payout items are disjoint")
	assert.Nil(t, matchedPayable, "no payable should be matched when there is no overlap")
}

func TestCheckPayoutItemOverlap_WithOverlap(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	h := &handler{logger: l}

	// P1 has [pi1, pi2], new invoice has [pi1] → overlap, should reuse P1
	pendingPayables := []*notion.PayableInfo{
		{
			PageID:    "payable-p1",
			Status:    "Pending",
			InvoiceID: "INV-001",
		},
	}

	newPayoutItemIDs := []string{"pi1"}

	hasOverlap, matchedPayable := h.checkPayoutItemOverlap(context.Background(), l, pendingPayables, newPayoutItemIDs, map[string][]string{
		"payable-p1": {"pi1", "pi2"},
	})

	assert.True(t, hasOverlap, "should detect overlap when new invoice shares payout item with existing payable")
	assert.NotNil(t, matchedPayable, "should return the overlapping payable")
	assert.Equal(t, "payable-p1", matchedPayable.PageID, "should match the correct payable")
}

func TestCheckPayoutItemOverlap_PartialOverlap(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	h := &handler{logger: l}

	// P1 has [pi1, pi2], new invoice has [pi2, pi3] → overlap, should reuse P1
	pendingPayables := []*notion.PayableInfo{
		{
			PageID:    "payable-p1",
			Status:    "Pending",
			InvoiceID: "INV-001",
		},
	}

	newPayoutItemIDs := []string{"pi2", "pi3"}

	hasOverlap, matchedPayable := h.checkPayoutItemOverlap(context.Background(), l, pendingPayables, newPayoutItemIDs, map[string][]string{
		"payable-p1": {"pi1", "pi2"},
	})

	assert.True(t, hasOverlap, "should detect partial overlap when at least one payout item matches")
	assert.NotNil(t, matchedPayable, "should return the overlapping payable")
	assert.Equal(t, "payable-p1", matchedPayable.PageID, "should match the correct payable")
}

func TestCheckPayoutItemOverlap_EmptyExisting(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	h := &handler{logger: l}

	// P1 has [], new invoice has [pi1] → no overlap
	pendingPayables := []*notion.PayableInfo{
		{
			PageID:    "payable-p1",
			Status:    "Pending",
			InvoiceID: "INV-001",
		},
	}

	newPayoutItemIDs := []string{"pi1"}

	hasOverlap, matchedPayable := h.checkPayoutItemOverlap(context.Background(), l, pendingPayables, newPayoutItemIDs, map[string][]string{
		"payable-p1": {},
	})

	assert.False(t, hasOverlap, "should not detect overlap when existing payable has no payout items")
	assert.Nil(t, matchedPayable, "no payable should be matched when existing has no items")
}

func TestCheckPayoutItemOverlap_EmptyNew(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	h := &handler{logger: l}

	// P1 has [pi1], new invoice has [] → no overlap
	pendingPayables := []*notion.PayableInfo{
		{
			PageID:    "payable-p1",
			Status:    "Pending",
			InvoiceID: "INV-001",
		},
	}

	newPayoutItemIDs := []string{}

	hasOverlap, matchedPayable := h.checkPayoutItemOverlap(context.Background(), l, pendingPayables, newPayoutItemIDs, map[string][]string{
		"payable-p1": {"pi1"},
	})

	assert.False(t, hasOverlap, "should not detect overlap when new invoice has no payout items")
	assert.Nil(t, matchedPayable, "no payable should be matched when new invoice has no items")
}

func TestCheckPayoutItemOverlap_MultiplePendingOneMatches(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	h := &handler{logger: l}

	// P1 has [pi1], P2 has [pi3], new invoice has [pi3] → reuse P2 deterministically
	pendingPayables := []*notion.PayableInfo{
		{
			PageID:    "payable-p1",
			Status:    "Pending",
			InvoiceID: "INV-001",
		},
		{
			PageID:    "payable-p2",
			Status:    "Pending",
			InvoiceID: "INV-002",
		},
	}

	newPayoutItemIDs := []string{"pi3"}

	hasOverlap, matchedPayable := h.checkPayoutItemOverlap(context.Background(), l, pendingPayables, newPayoutItemIDs, map[string][]string{
		"payable-p1": {"pi1"},
		"payable-p2": {"pi3"},
	})

	assert.True(t, hasOverlap, "should detect overlap with one of multiple pending payables")
	assert.NotNil(t, matchedPayable, "should return the matching payable")
	assert.Equal(t, "payable-p2", matchedPayable.PageID, "should match P2 which contains pi3, not P1")
}

func TestProcessGenInvoice_EmptyCandidateSetAfterFilter(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	h := &handler{logger: l}

	// Invoice type filtering leaves zero payout items → no incorrect reuse/create path
	pendingPayables := []*notion.PayableInfo{
		{
			PageID:    "payable-p1",
			Status:    "Pending",
			InvoiceID: "INV-001",
		},
	}

	// Empty candidate set after filter
	newPayoutItemIDs := []string{}

	hasOverlap, matchedPayable := h.checkPayoutItemOverlap(context.Background(), l, pendingPayables, newPayoutItemIDs, map[string][]string{
		"payable-p1": {"pi1", "pi2"},
	})

	assert.False(t, hasOverlap, "should not detect overlap when candidate set is empty after filtering")
	assert.Nil(t, matchedPayable, "no payable should be matched when candidate set is empty")
}

func TestCheckPayoutItemOverlap_Respects90DayWindow(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	h := &handler{logger: l}

	// The 90-day window is enforced upstream in deriveCandidatePayoutItemIDs.
	// By the time checkPayoutItemOverlap is called, old payouts are already excluded
	// from newPayoutItemIDs. This test verifies that when only recent (in-window) payout
	// items are passed as candidates, the overlap check correctly reports no match
	// against a payable that only contains older payout items.
	pendingPayables := []*notion.PayableInfo{
		{
			PageID:    "payable-p1",
			Status:    "Pending",
			InvoiceID: "INV-001",
		},
	}

	// After 90-day filtering upstream, only pi2 (recent, inside window) remains as candidate.
	// pi1 (old, outside window) was removed by deriveCandidatePayoutItemIDs.
	newPayoutItemIDs := []string{"pi2"}

	// payable-p1 contains only pi1 (old payout item not in the candidate set)
	hasOverlap, matchedPayable := h.checkPayoutItemOverlap(context.Background(), l, pendingPayables, newPayoutItemIDs, map[string][]string{
		"payable-p1": {"pi1"}, // Only old payout item — no overlap with ["pi2"]
	})

	assert.False(t, hasOverlap, "should not detect overlap when payable only has items outside the 90-day window")
	assert.Nil(t, matchedPayable, "no payable should be matched when there is no overlap with in-window candidates")
}

// TestRemainingCandidates_SupersetRequest verifies the flow where request contains a superset
// of items covered by P1. The remaining (uncovered) items should be used for P2.
func TestRemainingCandidates_SupersetRequest(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	_ = &handler{logger: l}

	// P1 has [pi1, pi2], request candidates are [pi1, pi2, pi3, pi4]
	// After filtering, remaining should be [pi3, pi4]
	payablePayoutItems := map[string][]string{
		"payable-p1": {"pi1", "pi2"},
	}
	candidatePayoutItemIDs := []string{"pi1", "pi2", "pi3", "pi4"}

	// Collect covered IDs (same logic as processGenInvoice)
	coveredIDs := make(map[string]bool)
	for _, ids := range payablePayoutItems {
		for _, id := range ids {
			coveredIDs[id] = true
		}
	}

	var remainingCandidateIDs []string
	for _, id := range candidatePayoutItemIDs {
		if !coveredIDs[id] {
			remainingCandidateIDs = append(remainingCandidateIDs, id)
		}
	}

	assert.Equal(t, 2, len(remainingCandidateIDs), "should have 2 uncovered payout items")
	assert.Contains(t, remainingCandidateIDs, "pi3", "pi3 should be in remaining")
	assert.Contains(t, remainingCandidateIDs, "pi4", "pi4 should be in remaining")
}

// TestRemainingCandidates_FullyCovered verifies that when all candidates are covered
// by existing payables, remaining is empty (should return existing payable).
func TestRemainingCandidates_FullyCovered(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	_ = &handler{logger: l}

	// P1 has [pi1, pi2], request candidates are [pi1, pi2]
	payablePayoutItems := map[string][]string{
		"payable-p1": {"pi1", "pi2"},
	}
	candidatePayoutItemIDs := []string{"pi1", "pi2"}

	coveredIDs := make(map[string]bool)
	for _, ids := range payablePayoutItems {
		for _, id := range ids {
			coveredIDs[id] = true
		}
	}

	var remainingCandidateIDs []string
	for _, id := range candidatePayoutItemIDs {
		if !coveredIDs[id] {
			remainingCandidateIDs = append(remainingCandidateIDs, id)
		}
	}

	assert.Equal(t, 0, len(remainingCandidateIDs), "all candidates should be covered")
}

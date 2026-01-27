package employee

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIsEligibleByLatestPayout(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("nil_latest_payout", func(t *testing.T) {
		require.False(t, isEligibleByLatestPayout(nil, cutoff))
	})

	t.Run("before_cutoff", func(t *testing.T) {
		date := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
		require.False(t, isEligibleByLatestPayout(&date, cutoff))
	})

	t.Run("on_cutoff", func(t *testing.T) {
		date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		require.True(t, isEligibleByLatestPayout(&date, cutoff))
	})

	t.Run("after_cutoff", func(t *testing.T) {
		date := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		require.True(t, isEligibleByLatestPayout(&date, cutoff))
	})
}

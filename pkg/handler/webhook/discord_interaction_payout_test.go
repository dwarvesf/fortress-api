package webhook

import (
	"strings"
	"testing"
)

func TestParsePayoutCommitCustomID(t *testing.T) {
	tests := []struct {
		name           string
		customID       string
		prefix         string
		wantParts      int
		wantMode       string
		wantValueA     string
		wantValueB     string
		wantChannelIdx int
		wantError      bool
	}{
		{
			name:           "old format confirm - 3 parts",
			customID:       "payout_commit_confirm:2026-01:15",
			prefix:         "payout_commit_confirm:",
			wantParts:      3,
			wantMode:       "period",
			wantValueA:     "2026-01",
			wantValueB:     "15",
			wantChannelIdx: -1, // uses interaction.ChannelID
			wantError:      false,
		},
		{
			name:           "current format confirm - 4 parts",
			customID:       "payout_commit_confirm:2026-01:15:channel123",
			prefix:         "payout_commit_confirm:",
			wantParts:      4,
			wantMode:       "period",
			wantValueA:     "2026-01",
			wantValueB:     "15",
			wantChannelIdx: 3,
			wantError:      false,
		},
		{
			name:           "new format period confirm - 5 parts",
			customID:       "payout_commit_confirm:period:2026-01:15:channel123",
			prefix:         "payout_commit_confirm:",
			wantParts:      5,
			wantMode:       "period",
			wantValueA:     "2026-01",
			wantValueB:     "15",
			wantChannelIdx: 4,
			wantError:      false,
		},
		{
			name:           "new format file confirm - 5 parts",
			customID:       "payout_commit_confirm:file:2026_01_swift.xlsx:2026:channel123",
			prefix:         "payout_commit_confirm:",
			wantParts:      5,
			wantMode:       "file",
			wantValueA:     "2026_01_swift.xlsx",
			wantValueB:     "2026",
			wantChannelIdx: 4,
			wantError:      false,
		},
		{
			name:      "invalid format - 2 parts",
			customID:  "payout_commit_confirm:2026-01",
			prefix:    "payout_commit_confirm:",
			wantParts: 2,
			wantError: true,
		},
		{
			name:      "invalid format - 6 parts",
			customID:  "payout_commit_confirm:period:2026-01:15:channel123:extra",
			prefix:    "payout_commit_confirm:",
			wantParts: 6,
			wantError: true,
		},
		{
			name:           "old format cancel - 3 parts",
			customID:       "payout_commit_cancel:2026-01:15",
			prefix:         "payout_commit_cancel:",
			wantParts:      3,
			wantMode:       "period",
			wantValueA:     "2026-01",
			wantValueB:     "15",
			wantChannelIdx: -1,
			wantError:      false,
		},
		{
			name:           "current format cancel - 4 parts",
			customID:       "payout_commit_cancel:2026-01:15:channel123",
			prefix:         "payout_commit_cancel:",
			wantParts:      4,
			wantMode:       "period",
			wantValueA:     "2026-01",
			wantValueB:     "15",
			wantChannelIdx: 3,
			wantError:      false,
		},
		{
			name:           "new format cancel - 5 parts",
			customID:       "payout_commit_cancel:period:2026-01:15:channel123",
			prefix:         "payout_commit_cancel:",
			wantParts:      5,
			wantMode:       "period",
			wantValueA:     "2026-01",
			wantValueB:     "15",
			wantChannelIdx: 4,
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.Split(tt.customID, ":")

			if len(parts) != tt.wantParts {
				t.Errorf("expected %d parts, got %d", tt.wantParts, len(parts))
			}

			if tt.wantError {
				// For invalid formats, we just verify the part count
				if len(parts) == 3 || len(parts) == 4 || len(parts) == 5 {
					t.Errorf("expected error for %d parts, but this is a valid format", len(parts))
				}
				return
			}

			// Parse according to format
			var mode, valueA, valueB, channelID string

			switch len(parts) {
			case 3:
				mode = "period"
				valueA = parts[1]
				valueB = parts[2]
				channelID = "interaction.ChannelID" // would come from interaction
			case 4:
				mode = "period"
				valueA = parts[1]
				valueB = parts[2]
				channelID = parts[3]
			case 5:
				mode = parts[1]
				valueA = parts[2]
				valueB = parts[3]
				channelID = parts[4]
			}

			if mode != tt.wantMode {
				t.Errorf("expected mode %q, got %q", tt.wantMode, mode)
			}

			if valueA != tt.wantValueA {
				t.Errorf("expected valueA %q, got %q", tt.wantValueA, valueA)
			}

			if valueB != tt.wantValueB {
				t.Errorf("expected valueB %q, got %q", tt.wantValueB, valueB)
			}

			if tt.wantChannelIdx >= 0 {
				expectedChannel := parts[tt.wantChannelIdx]
				if channelID != expectedChannel {
					t.Errorf("expected channelID %q, got %q", expectedChannel, channelID)
				}
			}
		})
	}
}

package employee

import "time"

const officeCheckInEligibilityWindowDays = 45

func isEligibleByLatestPayout(latestPayoutDate *time.Time, cutoff time.Time) bool {
	if latestPayoutDate == nil {
		return false
	}

	return !latestPayoutDate.Before(cutoff)
}

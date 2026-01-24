package employee

import "time"

const officeCheckInEligibilityWindowDays = 45

func isEligibleByFirstPayout(firstPayoutDate *time.Time, cutoff time.Time) bool {
	if firstPayoutDate == nil {
		return false
	}

	return !firstPayoutDate.Before(cutoff)
}

package scheduler

import "time"

const (
	retentionRampDays = 90.0
	retentionCeiling  = 0.95
)

// EffectiveRetention returns base when targetDate is nil. Otherwise it
// interpolates linearly from base to retentionCeiling over the last
// retentionRampDays days before targetDate, and stays at retentionCeiling
// on or after targetDate.
func EffectiveRetention(base float64, targetDate *time.Time, now time.Time) float64 {
	if targetDate == nil {
		return base
	}

	daysRemaining := targetDate.Sub(now).Hours() / 24
	if daysRemaining <= 0 {
		return retentionCeiling
	}
	if daysRemaining >= retentionRampDays {
		return base
	}

	t := (retentionRampDays - daysRemaining) / retentionRampDays
	return base + (retentionCeiling-base)*t
}

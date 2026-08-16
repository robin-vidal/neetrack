package httpapi

// Placeholder settings until user_settings lands: these mirror the schema's
// own column defaults and will become per-user values read from the
// database once settings are wired up.
const (
	defaultBlockSize       = 4
	defaultOpenMax         = 1
	defaultTargetRetention = 0.90

	minSessionSize = 1
	maxSessionSize = 20
)

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

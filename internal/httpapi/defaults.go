package httpapi

const (
	minSessionSize = 1
	maxSessionSize = 20
)

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

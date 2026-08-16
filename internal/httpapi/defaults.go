package httpapi

const (
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

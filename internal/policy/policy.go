package policy

// ComputeReward converts an integer evaluation score (0..10) into internal currency and academic credits.
// Rules (fixed):
// - currency = score * 10 (maps 0..10 -> 0..100)
// - credits = float64(score) / 2.5 (so 10 -> 4 credits)
// - clamps score to [0,10]
func ComputeReward(score int) (int64, float64) {
	if score < 0 {
		score = 0
	}
	if score > 10 {
		score = 10
	}
	currency := int64(score * 10)
	credits := float64(score) / 2.5
	return currency, credits
}

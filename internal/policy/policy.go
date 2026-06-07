package policy

// MaxPoints is the upper bound for an evaluation score.
const MaxPoints = 10

// CreditsForPoints converts an integer score (0..MaxPoints) into academic credits.
// Rule: credits = points / 2.5 (so 10 points -> 4 credits). Score is clamped.
func CreditsForPoints(points int) float64 {
	if points < 0 {
		points = 0
	}
	if points > MaxPoints {
		points = MaxPoints
	}
	return float64(points) / 2.5
}

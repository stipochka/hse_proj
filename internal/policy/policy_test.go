package policy

import "testing"

func TestCreditsForPoints(t *testing.T) {
	cases := []struct {
		points int
		want   float64
	}{
		{0, 0}, {5, 2.0}, {10, 4.0},
		{-3, 0},   // clamped low
		{99, 4.0}, // clamped high
	}
	for _, c := range cases {
		if got := CreditsForPoints(c.points); got != c.want {
			t.Errorf("CreditsForPoints(%d) = %v, want %v", c.points, got, c.want)
		}
	}
}

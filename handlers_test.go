package main

import "testing"

func TestUsedPercent(t *testing.T) {
	tests := []struct {
		name           string
		remaining      int
		boostPermille  int
		wantUsed       int
	}{
		{"5h: boost 2x, remaining 98%", 98, 2000, 4},
		{"weekly: boost 3x, remaining 98%", 98, 3000, 6},
		{"boost 2x, remaining 100% (nothing used)", 100, 2000, 0},
		{"boost 2x, remaining 0% (all used)", 0, 2000, 100},
		{"boost 1x, remaining 95%", 95, 1000, 5},
		{"boost 1x, remaining 100%", 100, 1000, 0},
		{"boost 1x, remaining 0%", 0, 1000, 100},
		{"no boost info: remaining 50%", 50, 0, 50},
		{"no boost info: remaining 100%", 100, 0, 0},
		{"no boost info: remaining 0%", 0, 0, 100},
		{"boost 2x, remaining 50% clamps to 100", 50, 2000, 100},
		{"boost 3x, remaining 0% clamps to 100", 0, 3000, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := usedPercent(tt.remaining, tt.boostPermille)
			if got != tt.wantUsed {
				t.Errorf("usedPercent(remaining=%d, boost=%d) = %d, want %d",
					tt.remaining, tt.boostPermille, got, tt.wantUsed)
			}
		})
	}
}

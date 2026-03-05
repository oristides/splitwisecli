package expense

import (
	"testing"
)

func TestParseSplitPercentages(t *testing.T) {
	tests := []struct {
		name      string
		val       string
		cost      float64
		wantMy    float64
		wantFr    float64
		wantError bool
	}{
		{"50,50", "50,50", 120, 60, 60, false},
		{"40,60", "40,60", 100, 40, 60, false},
		{"33,67", "33,67", 90, 29.70, 60.30, false},
		{"100,0", "100,0", 80, 80, 0, false},
		{"0,100", "0,100", 50, 0, 50, false},
		{"with spaces", " 40 , 60 ", 100, 40, 60, false},
		{"not 100", "30,60", 100, 0, 0, true},
		{"not 100", "50,40", 100, 0, 0, true},
		{"invalid format", "40", 100, 0, 0, true},
		{"invalid format", "40,60,20", 100, 0, 0, true},
		{"invalid number", "abc,60", 100, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			myOwed, friendOwed, err := ParseSplitPercentages(tt.val, tt.cost)
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseSplitPercentages() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ParseSplitPercentages() unexpected error: %v", err)
				return
			}
			if myOwed < tt.wantMy-0.02 || myOwed > tt.wantMy+0.02 {
				t.Errorf("ParseSplitPercentages() myOwed = %.2f, want ~%.2f", myOwed, tt.wantMy)
			}
			if friendOwed < tt.wantFr-0.02 || friendOwed > tt.wantFr+0.02 {
				t.Errorf("ParseSplitPercentages() friendOwed = %.2f, want ~%.2f", friendOwed, tt.wantFr)
			}
		})
	}
}

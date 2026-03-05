package expense

import (
	"testing"
)

func TestFindGroupID(t *testing.T) {
	groups := []GroupInfo{
		{ID: 1, Name: "Trip to Japan"},
		{ID: 2, Name: "Roommates"},
		{ID: 3, Name: "Summer Trip"},
	}

	tests := []struct {
		name      string
		val       string
		groups    []GroupInfo
		wantID    int
		wantError bool
	}{
		{"empty", "", groups, 0, false},
		{"numeric", "2", groups, 2, false},
		{"exact match", "Trip to Japan", groups, 1, false},
		{"exact match case insensitive", "trip to japan", groups, 1, false},
		{"partial match", "Room", groups, 2, false},
		{"partial match case insensitive", "summer", groups, 3, false},
		{"not found", "Nonexistent", groups, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindGroupID(tt.val, tt.groups)
			if tt.wantError {
				if err == nil {
					t.Errorf("FindGroupID() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("FindGroupID() unexpected error: %v", err)
				return
			}
			if got != tt.wantID {
				t.Errorf("FindGroupID() = %d, want %d", got, tt.wantID)
			}
		})
	}
}

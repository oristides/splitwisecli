package expense

import (
	"testing"
)

func TestResolvePaidBy(t *testing.T) {
	myID := 100
	friendID := 200

	tests := []struct {
		name      string
		val       string
		myID      int
		friendID  int
		wantID    int
		wantError bool
	}{
		{"empty -> me", "", myID, friendID, myID, false},
		{"me -> me", "me", myID, friendID, myID, false},
		{"ME -> me", "ME", myID, friendID, myID, false},
		{" friend -> friend", " friend ", myID, friendID, friendID, false},
		{"friend requires --friend", "friend", myID, 0, 0, true},
		{"friend with friendID", "friend", myID, friendID, friendID, false},
		{"numeric user ID", "456", myID, friendID, 456, false},
		{"invalid -> error", "invalid", myID, friendID, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolvePaidBy(tt.val, tt.myID, tt.friendID)
			if tt.wantError {
				if err == nil {
					t.Errorf("ResolvePaidBy() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ResolvePaidBy() unexpected error: %v", err)
				return
			}
			if got != tt.wantID {
				t.Errorf("ResolvePaidBy() = %d, want %d", got, tt.wantID)
			}
		})
	}
}

package expense

import (
	"fmt"
	"strconv"
	"strings"
)

// ResolvePaidBy returns the user ID of who paid.
// "" or "me" -> myID; "friend" -> friendID (when in friend expense); "123" -> 123.
func ResolvePaidBy(val string, myID int, friendID int) (int, error) {
	val = strings.TrimSpace(strings.ToLower(val))
	if val == "" || val == "me" {
		return myID, nil
	}
	if val == "friend" {
		if friendID == 0 {
			return 0, fmt.Errorf("--paid-by friend requires --friend <user_id>")
		}
		return friendID, nil
	}
	id, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("--paid-by must be 'me', 'friend', or a user ID (e.g. 456), got %q", val)
	}
	return id, nil
}

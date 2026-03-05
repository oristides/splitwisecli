package expense

import (
	"fmt"
	"strconv"
	"strings"
)

// GroupInfo is a minimal group for resolution (avoids importing client).
type GroupInfo struct {
	ID   int
	Name string
}

// FindGroupID returns the group ID. "" -> 0; "123" -> 123; "Trip" -> looks up by name in groups.
func FindGroupID(val string, groups []GroupInfo) (int, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return 0, nil
	}
	if id, err := strconv.Atoi(val); err == nil {
		return id, nil
	}
	valLower := strings.ToLower(val)
	for _, g := range groups {
		if strings.ToLower(g.Name) == valLower {
			return g.ID, nil
		}
	}
	for _, g := range groups {
		if strings.Contains(strings.ToLower(g.Name), valLower) {
			return g.ID, nil
		}
	}
	return 0, fmt.Errorf("group %q not found", val)
}

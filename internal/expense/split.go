package expense

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseSplitPercentages parses "a,b" as percentages (must sum to 100) and returns
// the amounts for the given cost. E.g. "40,60" with cost 120 -> 48, 72.
func ParseSplitPercentages(val string, cost float64) (myOwed, friendOwed float64, err error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return 0, 0, fmt.Errorf("split cannot be empty")
	}
	parts := strings.Split(val, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("--split must be 'myPercent,friendPercent' (e.g. 40,60), must sum to 100%%")
	}
	myPct, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	frPct, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("--split must be percentages (e.g. 40,60)")
	}
	sum := myPct + frPct
	if sum < 99.9 || sum > 100.1 {
		return 0, 0, fmt.Errorf("split percentages (%.1f + %.1f) must sum to 100", myPct, frPct)
	}
	myOwed = cost * myPct / 100
	friendOwed = cost * frPct / 100
	return myOwed, friendOwed, nil
}

package rules

const (
	High   = "HIGH"
	Medium = "MEDIUM"
	Low    = "LOW"

	// None is a valid --fail-on value meaning "never fail based on
	// finding severity" (parse/load errors still cause a non-zero exit).
	None = "NONE"
)

var severityRank = map[string]int{
	Low:    1,
	Medium: 2,
	High:   3,
}

// ValidSeverityThreshold reports whether s is a recognized --fail-on value.
func ValidSeverityThreshold(s string) bool {
	if s == None {
		return true
	}
	_, ok := severityRank[s]
	return ok
}

// MeetsThreshold reports whether severity is at or above threshold.
// A threshold of None never matches, since None means "don't fail".
func MeetsThreshold(severity, threshold string) bool {
	if threshold == None {
		return false
	}
	return severityRank[severity] >= severityRank[threshold]
}

package rules

type Finding struct {
	RuleID   string `json:"ruleId"`
	Category string `json:"category"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

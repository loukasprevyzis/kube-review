package rules

type Finding struct {
	Category string `json:"category"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

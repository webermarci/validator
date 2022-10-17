package validator

type RuleType string

const (
	StartsWith         = "startsWith"
	EndsWith           = "endsWith"
	LongerThan         = "longerThan"
	LongerThanOrEqual  = "longerThanOrEqual"
	ShorterThan        = "shorterThan"
	ShorterThanOrEqual = "shorterThanOrEqual"
	Contains           = "contains"
	ContainsACharacter = "containsACharacter"
	ContainsANumber    = "containsANumber"
	Ignore             = "ignore"
	IgnoreDuplicates   = "ignoreDuplicates"
	Regexp             = "regexp"
	Custom             = "custom"
)

type Rule struct {
	reason   string
	ruleType RuleType
	function func(input string) bool
}

type Result struct {
	Approval bool
	RuleType RuleType
	Reason   string
}

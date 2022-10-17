package validator

import (
	"fmt"
	"testing"
	"time"
)

func TestRules(t *testing.T) {
	var tests = []struct {
		name      string
		validator *Validator
		ruleType  RuleType
		reason    string
		approved  []string
		denied    []string
	}{
		{
			name:      "StartsWith",
			validator: NewValidator().StartsWith("123"),
			ruleType:  StartsWith,
			reason:    "starts with 123",
			approved:  []string{"123aaa", "123bbb", "123ccc"},
			denied:    []string{"aaa123", "bbb123", "ccc123"},
		},
		{
			name:      "EndsWith",
			validator: NewValidator().EndsWith("123"),
			ruleType:  EndsWith,
			reason:    "ends with 123",
			approved:  []string{"aaa123", "bbb123", "ccc123"},
			denied:    []string{"123aaa", "123bbb", "123ccc"},
		},
		{
			name:      "LongerThan",
			validator: NewValidator().LongerThan(4),
			ruleType:  LongerThan,
			reason:    "longer than 4",
			approved:  []string{"aaaaa", "aaaaa", "aaaaaa"},
			denied:    []string{"a", "aa", "aaa", "aaaa"},
		},
		{
			name:      "LongerThanOrEqual",
			validator: NewValidator().LongerThanOrEqual(5),
			ruleType:  LongerThanOrEqual,
			reason:    "longer than or equal to 5",
			approved:  []string{"aaaaa", "aaaaaa", "aaaaaaa"},
			denied:    []string{"a", "aa", "aaa", "aaaa"},
		},
		{
			name:      "ShorterThan",
			validator: NewValidator().ShorterThan(4),
			ruleType:  ShorterThan,
			reason:    "shorter than 4",
			approved:  []string{"a", "aa", "aaa"},
			denied:    []string{"aaaa", "aaaaa", "aaaaaa"},
		},
		{
			name:      "ShorterThanOrEqual",
			validator: NewValidator().ShorterThanOrEqual(5),
			ruleType:  ShorterThanOrEqual,
			reason:    "shorter than or equal to 5",
			approved:  []string{"a", "aa", "aaa", "aaaa", "aaaaa"},
			denied:    []string{"aaaaaa", "aaaaaaa", "aaaaaaaa"},
		},
		{
			name:      "Contains",
			validator: NewValidator().Contains("123"),
			ruleType:  Contains,
			reason:    "contains 123",
			approved:  []string{"aaa123aaa", "bbb123bbb", "ccc123ccc"},
			denied:    []string{"aaa", "bbb", "ccc"},
		},
		{
			name:      "ContainsACharacter",
			validator: NewValidator().ContainsACharacter(),
			ruleType:  ContainsACharacter,
			reason:    "contains a character",
			approved:  []string{"aaa", "bbb", "ccc"},
			denied:    []string{"111", "222", "333"},
		},
		{
			name:      "ContainsANumber",
			validator: NewValidator().ContainsANumber(),
			ruleType:  ContainsANumber,
			reason:    "contains a number",
			approved:  []string{"111", "222", "333"},
			denied:    []string{"aaa", "bbb", "ccc"},
		},
		{
			name:      "Ignore",
			validator: NewValidator().Ignore("aaa"),
			ruleType:  Ignore,
			reason:    "ignore aaa",
			approved:  []string{"bbb", "ccc"},
			denied:    []string{"aaa"},
		},
		{
			name:      "Regexp",
			validator: NewValidator().Regexp("t([a-z]+)t"),
			ruleType:  Regexp,
			reason:    "regexp t([a-z]+)t",
			approved:  []string{"test", "talent"},
			denied:    []string{"aaa", "bbb", "ccc"},
		},
		{
			name:      "InvalidRegexp",
			validator: NewValidator().Regexp("[0-9]++"),
			ruleType:  Regexp,
			reason:    "regexp [0-9]++",
			approved:  []string{},
			denied:    []string{"aaa", "bbb", "ccc"},
		},
		{
			name: "Custom",
			validator: NewValidator().Custom("custom reason", func(input string) bool {
				return len(input)%3 == 0
			}),
			ruleType: Custom,
			reason:   "custom reason",
			approved: []string{"aaa", "aaaaaa"},
			denied:   []string{"a", "aa", "aaaa", "aaaaa"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, a := range test.approved {
				result := test.validator.Validate(a)

				if !result.Approval {
					t.Fatal("approval expected")
				}

				if result.RuleType != "" {
					t.Fatal("rule type unexpected", result.RuleType)
				}

				if result.Reason != "" {
					t.Fatal("reason unexpected", result.Reason)
				}
			}

			for _, d := range test.denied {
				result := test.validator.Validate(d)

				if result.Approval {
					t.Fatal("deny expected")
				}

				if result.RuleType != test.ruleType {
					t.Fatal("invalid rule type", result.RuleType, test.ruleType)
				}

				if result.Reason != fmt.Sprintf("\"%s\" is not met by \"%s\"", test.reason, d) {
					t.Fatal("invalid reason", result.Reason, fmt.Sprintf("\"%s\" is not met by \"%s\"", test.reason, d))
				}
			}
		})
	}
}

func TestIgnoreDuplicates(t *testing.T) {
	validator := NewValidator().IgnoreDuplicatesFor(time.Millisecond)

	result := validator.Validate("aaa")
	if !result.Approval {
		t.Fatal("approval expected")
	}

	result = validator.Validate("aaa")
	if result.Approval {
		t.Fatal("deny expected")
	}

	time.Sleep(2 * time.Millisecond)

	result = validator.Validate("aaa")
	if !result.Approval {
		t.Fatal("approval expected")
	}

	result = validator.Validate("aaa")
	if result.Approval {
		t.Fatal("deny expected")
	}

	validator.StopIgnoringDuplicates()

	result = validator.Validate("aaa")
	if !result.Approval {
		t.Fatal("approval expected")
	}
}

func TestMultiple(t *testing.T) {
	validator := NewValidator().
		ContainsACharacter().
		ContainsANumber().
		LongerThanOrEqual(5).
		IgnoreAll([]string{
			"ABC002",
			"ABC003",
		}).
		IgnoreDuplicatesFor(time.Millisecond)

	result := validator.Validate("ABC002")
	if result.Approval {
		t.Fatal("deny expected")
	}

	result = validator.Validate("ABC003")
	if result.Approval {
		t.Fatal("deny expected")
	}

	result = validator.Validate("ABC001")
	if !result.Approval {
		t.Fatal("approval expected")
	}

	result = validator.Validate("ABC001")
	if result.Approval {
		t.Fatal("deny expected")
	}
}

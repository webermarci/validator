package validator

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Validator struct {
	rules          []*Rule
	ignoreDuration time.Duration
	recents        map[string]int64
	mutex          sync.RWMutex
	close          chan struct{}
}

func NewValidator() *Validator {
	return &Validator{
		rules:          []*Rule{},
		ignoreDuration: 0,
		recents:        make(map[string]int64),
		mutex:          sync.RWMutex{},
		close:          make(chan struct{}),
	}
}

func (v *Validator) Validate(input string) *Result {
	for _, r := range v.rules {
		if !r.function(input) {
			return &Result{
				Approval: false,
				RuleType: r.ruleType,
				Reason:   fmt.Sprintf("\"%s\" is not met by \"%s\"", r.reason, input),
			}
		}
	}
	if v.ignoreDuration > 0 {
		v.mutex.RLock()
		_, found := v.recents[input]
		v.mutex.RUnlock()
		if found {
			return &Result{
				Approval: false,
				RuleType: IgnoreDuplicates,
				Reason:   "ignore duplication",
			}
		}
		v.mutex.Lock()
		v.recents[input] = time.Now().Add(v.ignoreDuration).UnixNano()
		v.mutex.Unlock()
	}
	return &Result{
		Approval: true,
	}
}

func (v *Validator) Custom(denyReason string, function func(input string) bool) *Validator {
	v.rules = append(v.rules, &Rule{
		reason:   denyReason,
		ruleType: Custom,
		function: function,
	})
	return v
}

func (v *Validator) StartsWith(text string) *Validator {
	v.rules = append(v.rules, &Rule{
		ruleType: StartsWith,
		reason:   fmt.Sprintf("starts with %s", text),
		function: func(input string) bool {
			return strings.HasPrefix(input, text)
		},
	})
	return v
}

func (v *Validator) EndsWith(text string) *Validator {
	v.rules = append(v.rules, &Rule{
		ruleType: EndsWith,
		reason:   fmt.Sprintf("ends with %s", text),
		function: func(input string) bool {
			return strings.HasSuffix(input, text)
		},
	})
	return v
}

func (v *Validator) LongerThan(length int) *Validator {
	v.rules = append(v.rules, &Rule{
		ruleType: LongerThan,
		reason:   fmt.Sprintf("longer than %d", length),
		function: func(input string) bool {
			return len(input) > length
		},
	})
	return v
}

func (v *Validator) LongerThanOrEqual(length int) *Validator {
	v.rules = append(v.rules, &Rule{
		ruleType: LongerThanOrEqual,
		reason:   fmt.Sprintf("longer than or equal to %d", length),
		function: func(input string) bool {
			return len(input) >= length
		},
	})
	return v
}

func (v *Validator) ShorterThan(length int) *Validator {
	v.rules = append(v.rules, &Rule{
		ruleType: ShorterThan,
		reason:   fmt.Sprintf("shorter than %d", length),
		function: func(input string) bool {
			return len(input) < length
		},
	})
	return v
}

func (v *Validator) ShorterThanOrEqual(length int) *Validator {
	v.rules = append(v.rules, &Rule{
		ruleType: ShorterThanOrEqual,
		reason:   fmt.Sprintf("shorter than or equal to %d", length),
		function: func(input string) bool {
			return len(input) <= length
		},
	})
	return v
}

func (v *Validator) Contains(text string) *Validator {
	v.rules = append(v.rules, &Rule{
		ruleType: Contains,
		reason:   fmt.Sprintf("contains %s", text),
		function: func(input string) bool {
			return strings.Contains(input, text)
		},
	})
	return v
}

func (v *Validator) ContainsACharacter() *Validator {
	v.rules = append(v.rules, &Rule{
		ruleType: ContainsACharacter,
		reason:   "contains a character",
		function: func(input string) bool {
			for _, r := range input {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
					return true
				}
			}
			return false
		},
	})
	return v
}

func (v *Validator) ContainsANumber() *Validator {
	v.rules = append(v.rules, &Rule{
		ruleType: ContainsANumber,
		reason:   "contains a number",
		function: func(input string) bool {
			for _, r := range input {
				if r >= '0' && r <= '9' {
					return true
				}
			}
			return false
		},
	})
	return v
}

func (v *Validator) Ignore(text string) *Validator {
	v.rules = append(v.rules, &Rule{
		ruleType: Ignore,
		reason:   fmt.Sprintf("ignore %s", text),
		function: func(input string) bool {
			return input != text
		},
	})
	return v
}

func (v *Validator) IgnoreAll(texts []string) *Validator {
	for _, text := range texts {
		v.Ignore(text)
	}
	return v
}

func (v *Validator) Regexp(r string) *Validator {
	v.rules = append(v.rules, &Rule{
		ruleType: Regexp,
		reason:   fmt.Sprintf("regexp %s", r),
		function: func(input string) bool {
			matched, err := regexp.MatchString(r, input)
			if err != nil {
				return false
			}
			return matched
		},
	})
	return v
}

func (v *Validator) IgnoreDuplicatesFor(duration time.Duration) *Validator {
	go func() {
		ticker := time.NewTicker(duration / 2)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				v.mutex.Lock()
				for text, expires := range v.recents {
					if time.Now().UnixNano() > expires {
						delete(v.recents, text)
					}
				}
				v.mutex.Unlock()

			case <-v.close:
				return
			}
		}
	}()
	v.ignoreDuration = duration
	return v
}

func (v *Validator) StopIgnoringDuplicates() *Validator {
	v.ignoreDuration = 0
	v.close <- struct{}{}
	v.mutex.Lock()
	v.recents = make(map[string]int64)
	v.mutex.Unlock()
	return v
}

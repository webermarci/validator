# Validator

## Usage

```go
import "github.com/webermarci/validator"

v := validator.NewValidator().StartsWith("123")

result := v.Validate("abc123")

fmt.Println(result.Approval)
// False

fmt.Println(result.RuleType)
// startsWith

fmt.Println(result.Reason)
// "starts with 123" is not met by "abc123"

result = v.Validate("123abc")

fmt.Println(result.Approval)
// True
```
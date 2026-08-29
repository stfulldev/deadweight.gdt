package budget

import "fmt"

// PartialOverride is a domain-level override of config fail_on_partial.
// PartialInherit is the zero value used when neither CLI override is present.
type PartialOverride string

const (
	PartialInherit PartialOverride = ""
	PartialFail    PartialOverride = "fail"
	PartialAllow   PartialOverride = "allow"
)

// Valid reports whether override is part of the MVP partial-policy contract.
func (override PartialOverride) Valid() bool {
	return override == PartialInherit || override == PartialFail || override == PartialAllow
}

// PartialOverrideError reports an unsupported partial policy intent.
type PartialOverrideError struct {
	Override PartialOverride
}

func (err *PartialOverrideError) Error() string {
	return fmt.Sprintf("invalid partial policy override %q", err.Override)
}

// ResolveFailOnPartial applies a validated domain override to config policy.
func ResolveFailOnPartial(configured bool, override PartialOverride) (bool, error) {
	switch override {
	case PartialInherit:
		return configured, nil
	case PartialFail:
		return true, nil
	case PartialAllow:
		return false, nil
	default:
		return false, &PartialOverrideError{Override: override}
	}
}

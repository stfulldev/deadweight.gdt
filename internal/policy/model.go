// Package policy resolves presets, custom profiles, and overrides into one
// effective check policy without performing scene I/O or budget evaluation.
package policy

import "github.com/stfulldev/deadweight.gdt/internal/budget"

// Kind identifies the selected base policy domain. KindNone is the zero value.
type Kind string

const (
	KindNone    Kind = ""
	KindPreset  Kind = "preset"
	KindProfile Kind = "profile"
)

// Valid reports whether kind is part of the MVP policy contract.
func (kind Kind) Valid() bool {
	return kind == KindNone || kind == KindPreset || kind == KindProfile
}

// Selector contains the mutually exclusive CLI base selectors.
// Empty fields mean that configuration selection should be used.
type Selector struct {
	Preset  string
	Profile string
}

// Metadata is the complete metadata inherited by an effective base policy.
type Metadata struct {
	Name        string
	Description string
	Platform    string
	Renderer    string
	TargetFPS   int64
	Quality     string
	Status      string
	Stability   string
}

// Effective is one fully resolved check policy. KindNone and an empty ID mean
// the policy was formed exclusively from project or CLI budget overrides.
type Effective struct {
	Kind     Kind
	ID       string
	Metadata Metadata
	Budgets  budget.Limits
}

// Clone returns a deep copy whose optional budget values do not alias input.
func (effective Effective) Clone() Effective {
	effective.Budgets = effective.Budgets.Clone()
	return effective
}

package config

import "regexp"

// IDPattern is the frozen case-sensitive preset/profile identifier syntax.
const IDPattern = `^[a-z0-9][a-z0-9._-]{0,63}$`

var validIDPattern = regexp.MustCompile(IDPattern)

var rendererIDs = [...]string{
	"forward_plus",
	"mobile",
	"compatibility",
	"unspecified",
}

var qualityIDs = [...]string{
	"low",
	"balanced",
	"high",
	"custom",
}

// ValidID reports whether a selector or profile identifier is valid.
func ValidID(value string) bool {
	return validIDPattern.MatchString(value)
}

// ValidRenderer reports whether value is a frozen renderer identifier.
func ValidRenderer(value string) bool {
	return contains(rendererIDs[:], value)
}

// ValidQuality reports whether value is a frozen quality identifier.
func ValidQuality(value string) bool {
	return contains(qualityIDs[:], value)
}

// RendererIDs returns renderer identifiers in deterministic schema order.
func RendererIDs() []string {
	return append([]string(nil), rendererIDs[:]...)
}

// QualityIDs returns quality identifiers in deterministic schema order.
func QualityIDs() []string {
	return append([]string(nil), qualityIDs[:]...)
}

func contains(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}

	return false
}

package diagnostic

// Code is a stable machine-readable diagnostic identifier.
type Code string

const (
	CodeUnresolvedSceneInstance Code = "SB1001"
	CodeImportedScene           Code = "SB1002"
	CodeInheritedScene          Code = "SB1003"
	CodeUnavailableResource     Code = "SB1004"
	CodeInstancePlaceholder     Code = "SB1005"
	CodeUnsupportedResourcePath Code = "SB1006"
	CodeUnclassifiedCustomType  Code = "SB1007"
	CodeUnsupportedParent       Code = "SB1008"
	CodeInvalidTSCNRoot         Code = "SB2001"
	CodeSceneDependencyCycle    Code = "SB2002"
	CodeInvalidConfiguration    Code = "SB2003"
	CodeArithmeticOverflow      Code = "SB2004"
)

// Definition describes one stable diagnostic code and its required severity.
type Definition struct {
	Code     Code
	Severity Severity
}

var definitions = [...]Definition{
	{Code: CodeUnresolvedSceneInstance, Severity: SeverityWarning},
	{Code: CodeImportedScene, Severity: SeverityWarning},
	{Code: CodeInheritedScene, Severity: SeverityWarning},
	{Code: CodeUnavailableResource, Severity: SeverityWarning},
	{Code: CodeInstancePlaceholder, Severity: SeverityWarning},
	{Code: CodeUnsupportedResourcePath, Severity: SeverityWarning},
	{Code: CodeUnclassifiedCustomType, Severity: SeverityWarning},
	{Code: CodeUnsupportedParent, Severity: SeverityWarning},
	{Code: CodeInvalidTSCNRoot, Severity: SeverityError},
	{Code: CodeSceneDependencyCycle, Severity: SeverityError},
	{Code: CodeInvalidConfiguration, Severity: SeverityError},
	{Code: CodeArithmeticOverflow, Severity: SeverityError},
}

// Catalog returns a defensive copy of definitions in stable code order.
func Catalog() []Definition {
	catalog := make([]Definition, len(definitions))
	copy(catalog, definitions[:])

	return catalog
}

// Valid reports whether code is part of the MVP diagnostic taxonomy.
func (code Code) Valid() bool {
	_, ok := code.Severity()
	return ok
}

// Severity returns the severity required for code.
func (code Code) Severity() (Severity, bool) {
	for _, definition := range definitions {
		if definition.Code == code {
			return definition.Severity, true
		}
	}

	return "", false
}

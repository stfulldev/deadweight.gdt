package policy

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

// ParseBudgetOverrides converts ordered CLI metric=limit values into optional
// limits. Duplicate metrics are applied from left to right, so the last wins.
func ParseBudgetOverrides(source string, values []string) (budget.Limits, error) {
	var limits budget.Limits
	for index, value := range values {
		field := fmt.Sprintf("cli.budgets[%d]", index)
		name, limit, err := parseBudgetOverride(source, field, value)
		if err != nil {
			return budget.Limits{}, err
		}
		setLimit(&limits, name, limit)
	}

	return limits.Clone(), nil
}

func parseBudgetOverride(source, field, value string) (metrics.Name, int64, error) {
	if strings.Count(value, "=") != 1 || strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", 0, policyError(source, field, fmt.Sprintf(
			"invalid CLI budget %q: expected exact metric=limit form",
			value,
		))
	}

	metricText, limitText, _ := strings.Cut(value, "=")
	metric := metrics.Name(metricText)
	if metricText == "" || !metric.Valid() {
		return "", 0, policyError(source, field, fmt.Sprintf(
			"invalid CLI budget %q: unknown metric %q",
			value,
			metricText,
		))
	}
	if limitText == "" || !decimalDigits(limitText) {
		return "", 0, policyError(source, field, fmt.Sprintf(
			"invalid CLI budget %q: limit must be a non-negative base-10 integer",
			value,
		))
	}

	limit, err := strconv.ParseInt(limitText, 10, 64)
	if err != nil {
		return "", 0, policyError(source, field, fmt.Sprintf(
			"invalid CLI budget %q: limit exceeds signed 64-bit range",
			value,
		))
	}

	return metric, limit, nil
}

func decimalDigits(value string) bool {
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return value != ""
}

func policyError(source, field, detail string) *config.Error {
	return &config.Error{
		Reason: config.ReasonValidation,
		Source: source,
		Field:  field,
		Detail: detail,
	}
}

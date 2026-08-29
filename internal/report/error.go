package report

import (
	"fmt"
	"strings"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
)

// Error renders one fatal error for stderr without ANSI or stack traces.
func Error(err error) string {
	if err == nil {
		return ""
	}
	code, coded := diagnostic.CodeOf(err)
	if !coded {
		return fmt.Sprintf("ERROR: %v\n", err)
	}

	message := strings.TrimRight(diagnostic.MessageOf(err), "\n")
	lines := strings.Split(message, "\n")
	heading := ""
	if len(lines) > 0 {
		heading = lines[0]
	}

	var output strings.Builder
	fmt.Fprintf(&output, "ERROR %s: %s\n", code, heading)
	evidence := lines[1:]
	for len(evidence) > 0 && strings.TrimSpace(evidence[0]) == "" {
		evidence = evidence[1:]
	}
	if len(evidence) > 0 {
		output.WriteByte('\n')
		for _, line := range evidence {
			if line == "" {
				output.WriteByte('\n')
				continue
			}
			fmt.Fprintf(&output, "  %s\n", line)
		}
	}

	return output.String()
}

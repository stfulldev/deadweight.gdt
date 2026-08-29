package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stfulldev/deadweight.gdt/internal/report"
)

type presentationFormat string

const (
	presentationText presentationFormat = "text"
	presentationJSON presentationFormat = "json"
)

func parsePresentationFormat(value string) (presentationFormat, error) {
	format := presentationFormat(value)
	switch format {
	case presentationText, presentationJSON:
		return format, nil
	default:
		return "", fmt.Errorf("invalid format %q; want text or json", value)
	}
}

func sceneArguments(format *string) cobra.PositionalArgs {
	return func(command *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(command, args); err != nil {
			return err
		}
		_, err := parsePresentationFormat(*format)
		return err
	}
}

type presentationError struct {
	err     error
	format  presentationFormat
	options report.Options
}

func (failure *presentationError) Error() string {
	return failure.err.Error()
}

func (failure *presentationError) Unwrap() error {
	return failure.err
}

func wrapPresentationError(err error, format presentationFormat, options report.Options) error {
	if err == nil || format != presentationJSON {
		return err
	}
	return &presentationError{err: err, format: format, options: options}
}

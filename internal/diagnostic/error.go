package diagnostic

import "errors"

// CodedError exposes a stable diagnostic code without requiring message parsing.
type CodedError interface {
	error
	DiagnosticCode() Code
}

type messageError interface {
	DiagnosticMessage() string
}

// CodeOf finds a valid diagnostic code through an error's wrapping chain.
func CodeOf(err error) (Code, bool) {
	var coded CodedError
	if !errors.As(err, &coded) {
		return "", false
	}

	code := coded.DiagnosticCode()
	return code, code.Valid()
}

// MessageOf returns code-free diagnostic text when an error provides it.
func MessageOf(err error) string {
	var described messageError
	if errors.As(err, &described) {
		return described.DiagnosticMessage()
	}
	if err == nil {
		return ""
	}

	return err.Error()
}

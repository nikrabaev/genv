package core

import "fmt"

// GenvErrorCode is the stable error code included in JSON error envelopes
// and used for exit-code mapping.
type GenvErrorCode string

const (
	ErrValidation  GenvErrorCode = "VALIDATION"
	ErrParse       GenvErrorCode = "PARSE"
	ErrNotFound    GenvErrorCode = "NOT_FOUND"
	ErrAmbiguous   GenvErrorCode = "AMBIGUOUS"
	ErrBlocked     GenvErrorCode = "BLOCKED"
	ErrAuthMissing GenvErrorCode = "AUTH_MISSING"
	ErrAuthFailed  GenvErrorCode = "AUTH_FAILED"
	ErrVaultIO     GenvErrorCode = "VAULT_IO"
)

var exitCodes = map[GenvErrorCode]int{
	ErrValidation:  1,
	ErrParse:       1,
	ErrNotFound:    1,
	ErrAmbiguous:   1,
	ErrBlocked:     1,
	ErrAuthMissing: 3,
	ErrAuthFailed:  3,
	ErrVaultIO:     4,
}

// GenvError is the single error type used throughout genv. The Code drives
// exit code selection; Details carries structured context for JSON output.
type GenvError struct {
	Code    GenvErrorCode
	Message string
	Details any
}

func (e *GenvError) Error() string {
	return string(e.Code) + ": " + e.Message
}

func (e *GenvError) ExitCode() int {
	if code, ok := exitCodes[e.Code]; ok {
		return code
	}
	return 1
}

// Errorf creates a GenvError with a formatted message.
func Errorf(code GenvErrorCode, format string, args ...any) *GenvError {
	return &GenvError{Code: code, Message: fmt.Sprintf(format, args...)}
}

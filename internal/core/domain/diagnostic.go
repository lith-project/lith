package domain

import "errors"

var (
	ErrInvalidDiagnosticCode = errors.New("domain: invalid diagnostic code")
	ErrInvalidSeverity       = errors.New("domain: invalid diagnostic severity")
)

// DiagnosticCode is the stable parser identity LITH-P-NNNN.
type DiagnosticCode struct{ value string }

// NewDiagnosticCode parses a stable parser code.
func NewDiagnosticCode(value string) (DiagnosticCode, error) {
	const prefix = "LITH-P-"
	if len(value) != len(prefix)+4 || value[:len(prefix)] != prefix {
		return DiagnosticCode{}, ErrInvalidDiagnosticCode
	}
	for _, r := range value[len(prefix):] {
		if r < '0' || r > '9' {
			return DiagnosticCode{}, ErrInvalidDiagnosticCode
		}
	}
	return DiagnosticCode{value: value}, nil
}
func (c DiagnosticCode) String() string { return c.value }

// Severity is a stable diagnostic classification.
type Severity string

const (
	Info    Severity = "info"
	Warning Severity = "warning"
	Error   Severity = "error"
)

func (s Severity) valid() bool { return s == Info || s == Warning || s == Error }

// Diagnostic contains only stable code, severity, and source range.
type Diagnostic struct {
	code     DiagnosticCode
	severity Severity
	rangeVal Range
}

// NewDiagnostic constructs a message-independent diagnostic.
func NewDiagnostic(code DiagnosticCode, severity Severity, rangeVal Range) (Diagnostic, error) {
	if code.value == "" {
		return Diagnostic{}, ErrInvalidDiagnosticCode
	}
	if !severity.valid() {
		return Diagnostic{}, ErrInvalidSeverity
	}
	return Diagnostic{code: code, severity: severity, rangeVal: rangeVal}, nil
}
func (d Diagnostic) Code() DiagnosticCode { return d.code }
func (d Diagnostic) Severity() Severity   { return d.severity }
func (d Diagnostic) Range() Range         { return d.rangeVal }

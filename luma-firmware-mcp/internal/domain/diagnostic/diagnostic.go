package diagnostic

import "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/errors"

type Severity string

const (
 SeverityInfo Severity = "info"
 SeverityWarning Severity = "warning"
 SeverityError Severity = "error"
)

func (s Severity) Valid() bool { return s == SeverityInfo || s == SeverityWarning || s == SeverityError }

type Diagnostic struct { Severity Severity; Code, Message, File string; Line, Column int }

func (d Diagnostic) Validate() error {
 if !d.Severity.Valid() { return errors.Invalid("severity", "unknown diagnostic severity") }
 if d.Message == "" { return errors.Invalid("message", "diagnostic message is required") }
 if d.Line < 0 || d.Column < 0 { return errors.Invalid("location", "line and column cannot be negative") }
 return nil
}

func (d Diagnostic) IsError() bool { return d.Severity == SeverityError }

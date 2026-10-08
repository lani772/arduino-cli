package repair

import "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/errors"

type Status string

const (
 StatusProposed Status = "proposed"
 StatusApplied Status = "applied"
 StatusRejected Status = "rejected"
)

func (s Status) Valid() bool { return s == StatusProposed || s == StatusApplied || s == StatusRejected }

type Operation string

const (
 OperationCreate Operation = "create"
 OperationUpdate Operation = "update"
 OperationDelete Operation = "delete"
)

// FileChange is a declarative, content-based change. ExpectedSHA is the SHA-256
// of the current file content and prevents applying a patch to stale source.
type FileChange struct {
 Path string
 Operation Operation
 Content string
 ExpectedSHA string
}

type Patch struct {
 ID string
 Description string
 Diff string // Legacy, display-only unified diff; never executed by the patch engine.
 Status Status
 Files []FileChange
}

func (p Patch) Validate() error {
 if p.ID == "" { return errors.Invalid("id", "patch id is required") }
 if p.Description == "" { return errors.Invalid("description", "patch description is required") }
 if p.Diff == "" && len(p.Files) == 0 { return errors.Invalid("changes", "patch must include file changes or a display diff") }
 if !p.Status.Valid() { return errors.Invalid("status", "unknown patch status") }
 seen := map[string]bool{}
 for i, f := range p.Files {
  field := "files"
  if f.Path == "" { return errors.Invalid(field, "file path is required") }
  if seen[f.Path] { return errors.Invalid(field, "duplicate file path: "+f.Path) }
  seen[f.Path] = true
  switch f.Operation {
  case OperationCreate:
   if f.Content == "" { return errors.Invalid(field, "create operation requires non-empty content") }
   if f.ExpectedSHA != "" { return errors.Invalid(field, "create operation must not specify expected SHA") }
  case OperationUpdate:
   if f.Content == "" { return errors.Invalid(field, "update operation requires non-empty content") }
   if f.ExpectedSHA == "" { return errors.Invalid(field, "update operation requires expected SHA-256") }
  case OperationDelete:
   if f.Content != "" { return errors.Invalid(field, "delete operation must not include content") }
   if f.ExpectedSHA == "" { return errors.Invalid(field, "delete operation requires expected SHA-256") }
  default:
   return errors.Invalid(field, "unknown file operation at index "+itoa(i))
  }
 }
 return nil
}

func itoa(n int) string {
 if n == 0 { return "0" }
 digits := [20]byte{}
 i := len(digits)
 for n > 0 { i--; digits[i] = byte('0' + n%10); n /= 10 }
 return string(digits[i:])
}

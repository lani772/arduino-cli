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
type FileChange struct {
 Path string `json:"path"`
 Operation Operation `json:"operation"`
 Content string `json:"content,omitempty"`
 ExpectedSHA string `json:"expected_sha,omitempty"`
}
type Patch struct {
 ID string `json:"id"`
 Description string `json:"description"`
 Diff string `json:"diff,omitempty"`
 Status Status `json:"status"`
 Files []FileChange `json:"files,omitempty"`
}
func (p Patch) Validate() error {
 if p.ID == "" { return errors.Invalid("id", "patch id is required") }
 if p.Description == "" { return errors.Invalid("description", "patch description is required") }
 if p.Diff == "" && len(p.Files) == 0 { return errors.Invalid("changes", "patch must include file changes or a display diff") }
 if !p.Status.Valid() { return errors.Invalid("status", "unknown patch status") }
 seen := map[string]bool{}
 for i, f := range p.Files {
  if f.Path == "" { return errors.Invalid("files", "file path is required") }
  if seen[f.Path] { return errors.Invalid("files", "duplicate file path: "+f.Path) }; seen[f.Path] = true
  switch f.Operation {
  case OperationCreate:
   if f.Content == "" { return errors.Invalid("files", "create operation requires non-empty content") }
   if f.ExpectedSHA != "" { return errors.Invalid("files", "create operation must not specify expected SHA") }
  case OperationUpdate:
   if f.Content == "" { return errors.Invalid("files", "update operation requires non-empty content") }
   if f.ExpectedSHA == "" { return errors.Invalid("files", "update operation requires expected SHA-256") }
  case OperationDelete:
   if f.Content != "" { return errors.Invalid("files", "delete operation must not include content") }
   if f.ExpectedSHA == "" { return errors.Invalid("files", "delete operation requires expected SHA-256") }
  default:
   return errors.Invalid("files", "unknown file operation at index "+itoa(i))
  }
 }
 return nil
}
func itoa(n int) string {
 if n == 0 { return "0" }; digits := [20]byte{}; i := len(digits)
 for n > 0 { i--; digits[i] = byte('0' + n%10); n /= 10 }
 return string(digits[i:])
}

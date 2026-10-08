package repair

import "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/errors"

type Status string

const (
 StatusProposed Status = "proposed"
 StatusApplied Status = "applied"
 StatusRejected Status = "rejected"
)

func (s Status) Valid() bool { return s == StatusProposed || s == StatusApplied || s == StatusRejected }

type Patch struct { ID, Description, Diff string; Status Status }

func (p Patch) Validate() error {
 if p.ID == "" { return errors.Invalid("id", "patch id is required") }
 if p.Description == "" { return errors.Invalid("description", "patch description is required") }
 if p.Diff == "" { return errors.Invalid("diff", "patch diff is required") }
 if !p.Status.Valid() { return errors.Invalid("status", "unknown patch status") }
 return nil
}

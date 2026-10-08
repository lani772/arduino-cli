package build

import "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/errors"

type Status string

const (
 StatusPending Status = "pending"
 StatusRunning Status = "running"
 StatusPassed Status = "passed"
 StatusFailed Status = "failed"
)

func (s Status) Valid() bool { return s == StatusPending || s == StatusRunning || s == StatusPassed || s == StatusFailed }

type Result struct { ID string; Status Status; ExitCode int; Output, Artifact string }

func (r Result) Validate() error {
 if r.ID == "" { return errors.Invalid("id", "build result id is required") }
 if !r.Status.Valid() { return errors.Invalid("status", "unknown build status") }
 if r.ExitCode < 0 { return errors.Invalid("exit_code", "exit code cannot be negative") }
 return nil
}

package workspace

import "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/errors"

type Workspace struct { ID, Path string }

func (w Workspace) Validate() error {
 if w.ID == "" { return errors.Invalid("id", "workspace id is required") }
 if w.Path == "" { return errors.Invalid("path", "workspace path is required") }
 return nil
}

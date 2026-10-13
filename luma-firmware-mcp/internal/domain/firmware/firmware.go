package firmware

import "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/errors"

type Target struct { BoardFQBN, Port string }

func (t Target) Validate() error {
 if t.BoardFQBN == "" { return errors.Invalid("target.board_fqbn", "board FQBN is required") }
 return nil
}

type Firmware struct { ID, Version, Artifact string; Target Target }

func (f Firmware) Validate() error {
 if f.ID == "" { return errors.Invalid("id", "firmware id is required") }
 if f.Version == "" { return errors.Invalid("version", "firmware version is required") }
 return f.Target.Validate()
}

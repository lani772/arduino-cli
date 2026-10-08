package device

import "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/errors"

type Device struct { ID, Port, BoardFQBN string; Online bool }

func (d Device) Validate() error {
 if d.ID == "" { return errors.Invalid("id", "device id is required") }
 if d.BoardFQBN == "" { return errors.Invalid("board_fqbn", "device board FQBN is required") }
 return nil
}

package board

import "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/errors"

type Board struct { FQBN, Platform, Architecture, Variant string }

func (b Board) Validate() error {
 if b.FQBN == "" { return errors.Invalid("fqbn", "FQBN is required") }
 if b.Platform == "" { return errors.Invalid("platform", "platform is required") }
 if b.Architecture == "" { return errors.Invalid("architecture", "architecture is required") }
 return nil
}

func (b Board) Valid() bool { return b.Validate() == nil }

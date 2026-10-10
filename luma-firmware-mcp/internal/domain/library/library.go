package library

import "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/errors"

type Dependency struct { Name, Version, Source string }

func (d Dependency) Validate() error {
 if d.Name == "" { return errors.Invalid("name", "dependency name is required") }
 return nil
}

type InstallPolicy struct { AllowCoreInstall, AllowLibraryInstall bool }

package domain_test

import (
 "errors"
 "testing"
 domainerrors "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/errors"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/board"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/build"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/diagnostic"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/device"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/firmware"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/library"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/workspace"
)

func TestProjectValidation(t *testing.T) {
 _, err := project.New("","demo","ws")
 var target domainerrors.ValidationError
 if !errors.As(err,&target) || target.Field!="id" { t.Fatalf("expected id validation error, got %v",err) }
}

func TestCoreContractsRejectInvalidValues(t *testing.T) {
 tests := []struct{name string; err error}{
  {"board",board.Board{Platform:"esp32",Architecture:"xtensa"}.Validate()},
  {"build",build.Result{ID:"b",Status:"unknown"}.Validate()},
  {"diagnostic",diagnostic.Diagnostic{Severity:"bad",Message:"x"}.Validate()},
  {"device",device.Device{ID:"d"}.Validate()},
  {"firmware",firmware.Firmware{ID:"f",Version:"0.1"}.Validate()},
  {"library",library.Dependency{}.Validate()},
  {"patch",repair.Patch{ID:"p",Description:"d",Diff:"x",Status:"bad"}.Validate()},
  {"workspace",workspace.Workspace{ID:"w"}.Validate()},
 }
 for _,tt := range tests { if tt.err==nil { t.Errorf("%s: invalid value unexpectedly validated",tt.name) } }
}

func TestKnownEnums(t *testing.T) {
 if !build.StatusPassed.Valid() || !repair.StatusProposed.Valid() || !diagnostic.SeverityError.Valid() { t.Fatal("known enum values must validate") }
}

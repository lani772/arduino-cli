package application
import (
 "context"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/board"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/build"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/device"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/diagnostic"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/library"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
)
type ArduinoCLI interface {
 Version(context.Context) (string, error)
 ListBoards(context.Context) ([]board.Board, error)
 Compile(context.Context, string, string) (build.Result, error)
 Upload(context.Context, string, string, string) (build.Result, error)
 InstallLibrary(context.Context, library.Dependency) error
}
type SerialDiscovery interface { ListSerialPorts(context.Context) ([]device.SerialPort, error) }
type AIEngineer interface {
 GenerateFirmware(context.Context, string) (string, error)
 Diagnose(context.Context, []diagnostic.Diagnostic, string) (repair.Patch, error)
 Repair(context.Context, string, repair.Patch) (string, error)
}

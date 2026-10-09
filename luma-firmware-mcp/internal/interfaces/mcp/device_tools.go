package mcp
import (
 "context"
 "encoding/json"
 "fmt"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/application"
)
// RegisterDeviceTools adds read-only serial discovery. It never opens a serial
// session, resets a board, installs a core, or uploads firmware.
func RegisterDeviceTools(r *Registry, discovery application.SerialDiscovery) error {
 if discovery==nil{return fmt.Errorf("serial discovery adapter is required")}
 return r.Register(RegisteredTool{Definition:Tool{Name:"arduino_list_serial_devices",Description:"Discover serial ports reported by Arduino CLI, including ports with no matching board candidate. Detection does not verify device identity. This does not open ports, reset devices, install cores, or flash firmware.",InputSchema:objectSchema(map[string]any{})},Handler:func(ctx context.Context,_ json.RawMessage)(any,error){ports,err:=discovery.ListSerialPorts(ctx);if err!=nil{return nil,err};return map[string]any{"discovery_source":"arduino-cli board list --format json","read_only":true,"ports":ports},nil}})
}

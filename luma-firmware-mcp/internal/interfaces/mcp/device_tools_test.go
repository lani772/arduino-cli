package mcp
import ("context";"encoding/json";"strings";"testing";"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/device")
type fakeSerialCLI struct{ports []device.SerialPort}
func(f fakeSerialCLI)ListSerialPorts(context.Context)([]device.SerialPort,error){return f.ports,nil}
func TestDeviceToolIncludesUnmatchedPorts(t *testing.T){
 r:=NewRegistry();cli:=struct{fakeSerialCLI}{fakeSerialCLI{ports:[]device.SerialPort{{Address:"COM8"},{Address:"COM9",MatchingBoards:[]device.BoardMatch{{FQBN:"esp32:esp32:esp32",Name:"ESP32 Dev Module"}}}}}}
 if err:=RegisterDeviceTools(r,cli);err!=nil{t.Fatal(err)}
 result,err:=r.Call(context.Background(),"arduino_list_serial_devices",json.RawMessage("{}"));if err!=nil{t.Fatal(err)};if result.IsError{t.Fatalf("unexpected tool error: %+v",result)}
 content:=result.Content[0]["text"].(string);if !strings.Contains(content,"COM8")||!strings.Contains(content,"COM9")||!strings.Contains(content,"read_only"){t.Fatalf("unexpected result: %s",content)}
}

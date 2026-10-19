package mcp
import (
 "context"
 "encoding/json"
 "fmt"
 "strings"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/application"
)
// RegisterDeviceTools adds read-only discovery and conservative verification. Neither tool opens serial ports, resets boards, or uploads firmware.
func RegisterDeviceTools(r *Registry, discovery application.SerialDiscovery) error {
 if discovery == nil { return fmt.Errorf("serial discovery adapter is required") }
 if err := r.Register(RegisteredTool{Definition: Tool{Name:"arduino_list_serial_devices", Description:"Discover serial ports, including unmatched ports. Detection does not verify device identity.", InputSchema:objectSchema(map[string]any{})}, Handler:func(ctx context.Context, _ json.RawMessage)(any,error){ ports,err:=discovery.ListSerialPorts(ctx); if err!=nil{return nil,err}; return map[string]any{"discovery_source":"arduino-cli board list --format json","read_only":true,"ports":ports},nil }}); err != nil { return err }
 return r.Register(RegisteredTool{Definition:Tool{Name:"arduino_verify_device", Description:"Read-only post-upload evidence check. Separately reports port presence, matching board candidate, and optional supplied upload exit code. Port/board matching is not proof firmware is running; hardware_verified remains false without a runtime handshake.", InputSchema:objectSchema(map[string]any{"port":map[string]any{"type":"string"},"fqbn":map[string]any{"type":"string"},"upload_exit_code":map[string]any{"type":"integer","description":"Optional exit code from the preceding upload command."}},"port","fqbn")}, Handler:func(ctx context.Context, raw json.RawMessage)(any,error){
  var args map[string]json.RawMessage; if err:=json.Unmarshal(raw,&args);err!=nil{return nil,fmt.Errorf("invalid verification arguments: %w",err)}
  var port,fqbn string; if v:=args["port"];len(v)>0{_ = json.Unmarshal(v,&port)};if v:=args["fqbn"];len(v)>0{_ = json.Unmarshal(v,&fqbn)}
  port=strings.TrimSpace(port);fqbn=strings.TrimSpace(fqbn);if port==""||fqbn==""{return nil,fmt.Errorf("port and fqbn are required")}
  var uploadCode *int;if v:=args["upload_exit_code"];len(v)>0{var n int;if err:=json.Unmarshal(v,&n);err!=nil{return nil,fmt.Errorf("upload_exit_code must be an integer")};uploadCode=&n}
  ports,err:=discovery.ListSerialPorts(ctx);if err!=nil{return nil,fmt.Errorf("discover ports for verification: %w",err)}
  portDetected,boardMatched:=false,false
  for _,detected:=range ports{if detected.Address!=port{continue};portDetected=true;for _,candidate:=range detected.MatchingBoards{if candidate.FQBN==fqbn{boardMatched=true}}}
  var uploadSucceeded any;var reportedCode any;if uploadCode!=nil{uploadSucceeded=*uploadCode==0;reportedCode=*uploadCode}
  return map[string]any{"port":port,"expected_fqbn":fqbn,"serial_port_detected":portDetected,"board_candidate_matched":boardMatched,"upload_command_succeeded":uploadSucceeded,"upload_exit_code":reportedCode,"hardware_verified":false,"verification_level":"port_and_board_match","runtime_handshake_performed":false,"evidence_source":"arduino-cli board list --format json","note":"Port presence and a board candidate do not prove firmware is running on the intended device."},nil
 }})
}

package mcp
import ("context";"encoding/json";"testing";"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/device")
type fakeSerialCLI struct{ports []device.SerialPort}
func(f fakeSerialCLI)ListSerialPorts(context.Context)([]device.SerialPort,error){return f.ports,nil}
func verifyValue(t *testing.T,ports []device.SerialPort,args string)map[string]any{t.Helper();r:=NewRegistry();if err:=RegisterDeviceTools(r,fakeSerialCLI{ports});err!=nil{t.Fatal(err)};res,err:=r.Call(context.Background(),"arduino_verify_device",json.RawMessage(args));if err!=nil{t.Fatal(err)};if res.IsError{t.Fatalf("tool error: %+v",res)};var out map[string]any;if err:=json.Unmarshal([]byte(res.Content[0]["text"].(string)),&out);err!=nil{t.Fatal(err)};return out}
func TestVerifyDeviceReportsSeparateEvidence(t *testing.T){
 v:=verifyValue(t,[]device.SerialPort{{Address:"COM5",MatchingBoards:[]device.BoardMatch{{FQBN:"esp32:esp32:esp32"}}}},`{"port":"COM5","fqbn":"esp32:esp32:esp32","upload_exit_code":0}`)
 if v["serial_port_detected"]!=true||v["board_candidate_matched"]!=true||v["upload_command_succeeded"]!=true{t.Fatalf("unexpected statuses: %#v",v)}
 if v["hardware_verified"]!=false||v["runtime_handshake_performed"]!=false{t.Fatalf("must not claim runtime verification: %#v",v)}
}
func TestVerifyDeviceReportsMissingAndUnmatchedEvidence(t *testing.T){
 missing:=verifyValue(t,[]device.SerialPort{{Address:"COM5"}},`{"port":"COM6","fqbn":"esp32:esp32:esp32"}`)
 if missing["serial_port_detected"]!=false||missing["board_candidate_matched"]!=false||missing["upload_command_succeeded"]!=nil{t.Fatalf("unexpected missing-port status: %#v",missing)}
 unmatched:=verifyValue(t,[]device.SerialPort{{Address:"COM5",MatchingBoards:[]device.BoardMatch{{FQBN:"arduino:avr:uno"}}}},`{"port":"COM5","fqbn":"esp32:esp32:esp32","upload_exit_code":1}`)
 if unmatched["serial_port_detected"]!=true||unmatched["board_candidate_matched"]!=false||unmatched["upload_command_succeeded"]!=false||unmatched["hardware_verified"]!=false{t.Fatalf("unexpected unmatched status: %#v",unmatched)}
}

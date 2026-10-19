package arduino
import ("context";"testing";"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/policy")
func TestListSerialPortsPreservesMatchedAndUnmatchedPorts(t *testing.T){
 out:=`{"detected_ports":[{"port":{"address":"COM5","label":"USB Serial","protocol":"serial","protocol_label":"Serial Port"},"matching_boards":[{"fqbn":"esp32:esp32:esp32","name":"ESP32 Dev Module"}]},{"port":{"address":"/dev/ttyUSB0","protocol":"serial"},"matching_boards":[]}]}`
 cli:=NewCLI("arduino-cli",&fakeRunner{out:out},policy.DefaultExecutionPolicy());ports,err:=cli.ListSerialPorts(context.Background());if err!=nil{t.Fatal(err)}
 if len(ports)!=2{t.Fatalf("expected 2 ports, got %+v",ports)}
 if ports[0].Address!="COM5"||len(ports[0].MatchingBoards)!=1||ports[0].MatchingBoards[0].FQBN!="esp32:esp32:esp32"{t.Fatalf("unexpected first port: %+v",ports[0])}
 if ports[1].Address!="/dev/ttyUSB0"||len(ports[1].MatchingBoards)!=0{t.Fatalf("unmatched port must remain visible: %+v",ports[1])}
}

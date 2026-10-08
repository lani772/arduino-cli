package arduino
import("context";"testing";"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/policy")
func TestListBoardsDecodesJSON(t *testing.T){out:="{\"detected_ports\":[{\"port\":{\"address\":\"COM5\"},\"matching_boards\":[{\"fqbn\":\"esp32:esp32:esp32\",\"name\":\"ESP32 Dev Module\"}]}]}";cli:=NewCLI("arduino-cli",&fakeRunner{out:out},policy.DefaultExecutionPolicy());boards,err:=cli.ListBoards(context.Background());if err!=nil{t.Fatal(err)};if len(boards)!=1||boards[0].FQBN!="esp32:esp32:esp32"{t.Fatalf("unexpected boards: %+v",boards)}}

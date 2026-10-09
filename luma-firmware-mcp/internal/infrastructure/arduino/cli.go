package arduino
import (
 "bytes"
 "context"
 "encoding/json"
 "fmt"
 "os/exec"
 "strings"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/board"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/build"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/device"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/library"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/policy"
)
type CommandRunner interface { Run(context.Context,string,[]string,string)(string,string,int,error) }
type OSCommandRunner struct{}
func (OSCommandRunner) Run(ctx context.Context,name string,args []string,dir string)(string,string,int,error){
 cmd:=exec.CommandContext(ctx,name,args...);cmd.Dir=dir
 var out,errOut bytes.Buffer;cmd.Stdout=&out;cmd.Stderr=&errOut
 err:=cmd.Run();if err==nil{return out.String(),errOut.String(),0,nil}
 if ctx.Err()!=nil{return out.String(),errOut.String(),-1,ctx.Err()}
 if ee,ok:=err.(*exec.ExitError);ok{return out.String(),errOut.String(),ee.ExitCode(),nil}
 return out.String(),errOut.String(),-1,err
}
type CLI struct{Executable string;Runner CommandRunner;Policy policy.ExecutionPolicy}
func NewCLI(executable string,runner CommandRunner,p policy.ExecutionPolicy)*CLI{if executable==""{executable="arduino-cli"};if runner==nil{runner=OSCommandRunner{}};return &CLI{Executable:executable,Runner:runner,Policy:p}}
func(c *CLI)run(ctx context.Context,args []string,dir string)(string,string,int,error){if !c.Policy.AllowsExecutable(c.Executable){return "","",-1,fmt.Errorf("execution denied for %q",c.Executable)};return c.Runner.Run(ctx,c.Executable,args,dir)}
func(c *CLI)Version(ctx context.Context)(string,error){out,errOut,code,err:=c.run(ctx,[]string{"version"},"");if err!=nil{return "",err};if code!=0{return "",fmt.Errorf("arduino-cli version failed with exit code %d: %s",code,strings.TrimSpace(errOut))};return strings.TrimSpace(out),nil}
type boardListResponse struct { DetectedPorts []struct {
 Port struct { Address string `json:"address"`; Label string `json:"label"`; Protocol string `json:"protocol"`; ProtocolLabel string `json:"protocol_label"` } `json:"port"`
 MatchingBoards []struct { FQBN string `json:"fqbn"`; Name string `json:"name"` } `json:"matching_boards"`
} `json:"detected_ports"` }
func(c *CLI)boardList(ctx context.Context)(boardListResponse,error){out,errOut,code,err:=c.run(ctx,[]string{"board","list","--format","json"},"");if err!=nil{return boardListResponse{},err};if code!=0{return boardListResponse{},fmt.Errorf("arduino-cli board list failed with exit code %d: %s",code,strings.TrimSpace(errOut))};var raw boardListResponse;if err:=json.Unmarshal([]byte(out),&raw);err!=nil{return boardListResponse{},fmt.Errorf("decode board list: %w",err)};return raw,nil}
func(c *CLI)ListBoards(ctx context.Context)([]board.Board,error){raw,err:=c.boardList(ctx);if err!=nil{return nil,err};boards:=make([]board.Board,0);for _,p:=range raw.DetectedPorts{for _,b:=range p.MatchingBoards{boards=append(boards,board.Board{FQBN:b.FQBN,Platform:b.Name})}};return boards,nil}
func(c *CLI)ListSerialPorts(ctx context.Context)([]device.SerialPort,error){raw,err:=c.boardList(ctx);if err!=nil{return nil,err};ports:=make([]device.SerialPort,0,len(raw.DetectedPorts));for _,p:=range raw.DetectedPorts{item:=device.SerialPort{Address:p.Port.Address,Label:p.Port.Label,Protocol:p.Port.Protocol,ProtocolLabel:p.Port.ProtocolLabel,MatchingBoards:[]device.BoardMatch{}};for _,b:=range p.MatchingBoards{item.MatchingBoards=append(item.MatchingBoards,device.BoardMatch{FQBN:b.FQBN,Name:b.Name})};ports=append(ports,item)};return ports,nil}
func(c *CLI)Compile(ctx context.Context,source,fqbn string)(build.Result,error){return c.build(ctx,"compile",source,fqbn,"")}
func(c *CLI)Upload(ctx context.Context,source,fqbn,port string)(build.Result,error){return c.build(ctx,"upload",source,fqbn,port)}
func(c *CLI)build(ctx context.Context,operation,source,fqbn,port string)(build.Result,error){args:=[]string{operation,"--fqbn",fqbn};if port!=""{args=append(args,"--port",port)};args=append(args,source);out,errOut,code,err:=c.run(ctx,args,"");if err!=nil{return build.Result{Status:build.StatusFailed,Output:out+"\n"+errOut,ExitCode:code},err};status:=build.StatusPassed;if code!=0{status=build.StatusFailed};r:=build.Result{ID:operation,Status:status,ExitCode:code,Output:strings.TrimSpace(out+"\n"+errOut)};if code!=0{return r,fmt.Errorf("arduino-cli %s failed with exit code %d",operation,code)};return r,nil}
func(c *CLI)InstallLibrary(ctx context.Context,d library.Dependency)error{if err:=d.Validate();err!=nil{return err};spec:=d.Name;if d.Version!=""{spec=fmt.Sprintf("%s@%s",d.Name,d.Version)};_,errOut,code,err:=c.run(ctx,[]string{"lib","install",spec},"");if err!=nil{return err};if code!=0{return fmt.Errorf("arduino-cli library install failed with exit code %d: %s",code,strings.TrimSpace(errOut))};return nil}

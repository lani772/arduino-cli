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
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/library"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/policy"
)

type CommandRunner interface {
 Run(context.Context,string,[]string,string)(string,string,int,error)
}
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
func NewCLI(executable string,runner CommandRunner,p policy.ExecutionPolicy)*CLI{
 if executable==""{executable="arduino-cli"};if runner==nil{runner=OSCommandRunner{}}
 return &CLI{Executable:executable,Runner:runner,Policy:p}
}
func(c *CLI)run(ctx context.Context,args []string,dir string)(string,string,int,error){
 if !c.Policy.AllowsExecutable(c.Executable){return "","",-1,fmt.Errorf("execution denied for %q",c.Executable)}
 return c.Runner.Run(ctx,c.Executable,args,dir)
}
func(c *CLI)Version(ctx context.Context)(string,error){
 out,errOut,code,err:=c.run(ctx,[]string{"version"},"");if err!=nil{return "",err}
 if code!=0{return "",fmt.Errorf("arduino-cli version failed with exit code %d: %s",code,strings.TrimSpace(errOut))}
 return strings.TrimSpace(out),nil
}
func(c *CLI)ListBoards(ctx context.Context)([]board.Board,error){
 out,errOut,code,err:=c.run(ctx,[]string{"board","list","--format","json"},"");if err!=nil{return nil,err}
 if code!=0{return nil,fmt.Errorf("arduino-cli board list failed with exit code %d: %s",code,strings.TrimSpace(errOut))}
 var raw map[string]interface{};if err:=json.Unmarshal([]byte(out),&raw);err!=nil{return nil,fmt.Errorf("decode board list: %w",err)}
 detected,_:=raw["detected_ports"].([]interface{});boards:=make([]board.Board,0)
 for _,item:=range detected{p,_:=item.(map[string]interface{});matches,_:=p["matching_boards"].([]interface{});for _,m:=range matches{b,_:=m.(map[string]interface{});fqbn,_:=b["fqbn"].(string);name,_:=b["name"].(string);boards=append(boards,board.Board{FQBN:fqbn,Platform:name})}}
 return boards,nil
}
func(c *CLI)Compile(ctx context.Context,source,fqbn string)(build.Result,error){return c.build(ctx,"compile",source,fqbn,"")}
func(c *CLI)Upload(ctx context.Context,source,fqbn,port string)(build.Result,error){return c.build(ctx,"upload",source,fqbn,port)}
func(c *CLI)build(ctx context.Context,operation,source,fqbn,port string)(build.Result,error){
 args:=[]string{operation,"--fqbn",fqbn};if port!=""{args=append(args,"--port",port)};args=append(args,source)
 out,errOut,code,err:=c.run(ctx,args,"");if err!=nil{return build.Result{Status:build.StatusFailed,Output:out+"\n"+errOut,ExitCode:code},err}
 status:=build.StatusPassed;if code!=0{status=build.StatusFailed}
 r:=build.Result{ID:operation,Status:status,ExitCode:code,Output:strings.TrimSpace(out+"\n"+errOut)}
 if code!=0{return r,fmt.Errorf("arduino-cli %s failed with exit code %d",operation,code)}
 return r,nil
}
func(c *CLI)InstallLibrary(ctx context.Context,d library.Dependency)error{
 if err:=d.Validate();err!=nil{return err};spec:=d.Name;if d.Version!=""{spec=fmt.Sprintf("%s@%s",d.Name,d.Version)}
 _,errOut,code,err:=c.run(ctx,[]string{"lib","install",spec},"");if err!=nil{return err}
 if code!=0{return fmt.Errorf("arduino-cli library install failed with exit code %d: %s",code,strings.TrimSpace(errOut))}
 return nil
}

package arduino

import("context";"testing";"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/build";"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/policy")
type fakeRunner struct{name string;args []string;code int;out,errOut string;err error}
func(f *fakeRunner)Run(_ context.Context,name string,args []string,_ string)(string,string,int,error){f.name=name;f.args=append([]string(nil),args...);return f.out,f.errOut,f.code,f.err}
func TestCompileUsesControlledCommand(t *testing.T){r:=&fakeRunner{code:0,out:"compiled"};cli:=NewCLI("arduino-cli",r,policy.DefaultExecutionPolicy());got,err:=cli.Compile(context.Background(),"./firmware","esp32:esp32:esp32");if err!=nil{t.Fatal(err)};want:=[]string{"compile","--fqbn","esp32:esp32:esp32","./firmware"};if len(r.args)!=len(want){t.Fatalf("args=%v",r.args)};for i:=range want{if r.args[i]!=want[i]{t.Fatalf("args=%v want=%v",r.args,want)}};if got.Status!=build.StatusPassed||got.ExitCode!=0{t.Fatalf("unexpected result: %+v",got)}}
func TestPolicyBlocksUnknownExecutable(t *testing.T){cli:=NewCLI("sh",&fakeRunner{},policy.DefaultExecutionPolicy());if _,err:=cli.Version(context.Background());err==nil{t.Fatal("expected policy denial")}}
func TestNonZeroExitReturnsFailedResult(t *testing.T){cli:=NewCLI("arduino-cli",&fakeRunner{code:2,errOut:"compile error"},policy.DefaultExecutionPolicy());got,err:=cli.Compile(context.Background(),"./firmware","esp32:esp32:esp32");if err==nil{t.Fatal("expected error")};if got.Status!=build.StatusFailed||got.ExitCode!=2{t.Fatalf("unexpected result: %+v",got)}}

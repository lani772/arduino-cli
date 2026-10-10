package arduino

import (
 "context"
 "strings"
 "testing"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/policy"
)

func TestUploadBlockedByDefaultPolicy(t *testing.T) {
 runner:=&fakeRunner{code:0,out:"uploaded"}
 cli:=NewCLI("arduino-cli",runner,policy.DefaultExecutionPolicy())
 _,err:=cli.Upload(context.Background(),"./firmware","esp32:esp32:esp32","COM5")
 if err==nil||!strings.Contains(err.Error(),"disabled by execution policy"){t.Fatalf("expected policy denial, got %v",err)}
 if len(runner.args)!=0{t.Fatalf("blocked upload must not execute CLI, args=%v",runner.args)}
}

func TestUploadRequiresExplicitPort(t *testing.T) {
 p:=policy.DefaultExecutionPolicy();p.AllowDeviceFlash=true
 runner:=&fakeRunner{code:0,out:"uploaded"}
 cli:=NewCLI("arduino-cli",runner,p)
 _,err:=cli.Upload(context.Background(),"./firmware","esp32:esp32:esp32","")
 if err==nil||!strings.Contains(err.Error(),"serial port are required"){t.Fatalf("expected missing port error, got %v",err)}
 if len(runner.args)!=0{t.Fatalf("invalid upload must not execute CLI, args=%v",runner.args)}
}

func TestUploadUsesExplicitPortWhenPolicyAllows(t *testing.T) {
 p:=policy.DefaultExecutionPolicy();p.AllowDeviceFlash=true
 runner:=&fakeRunner{code:0,out:"uploaded"}
 cli:=NewCLI("arduino-cli",runner,p)
 got,err:=cli.Upload(context.Background(),"./firmware","esp32:esp32:esp32","COM5")
 if err!=nil{t.Fatal(err)}
 want:=[]string{"upload","--fqbn","esp32:esp32:esp32","--port","COM5","./firmware"}
 if len(runner.args)!=len(want){t.Fatalf("args=%v want=%v",runner.args,want)}
 for i:=range want{if runner.args[i]!=want[i]{t.Fatalf("args=%v want=%v",runner.args,want)}}
 if got.Status!= "passed" {t.Fatalf("unexpected upload result: %+v",got)}
}

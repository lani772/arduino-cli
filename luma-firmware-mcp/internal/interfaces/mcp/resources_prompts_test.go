package mcp

import (
 "context"
 "strings"
 "testing"

 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/board"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/build"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/library"
)
func TestDefaultPromptsRequireInputsAndReturnMessages(t *testing.T){
 r:=NewPromptRegistry();if err:=RegisterDefaultPrompts(r);err!=nil{t.Fatal(err)}
 if len(r.List())!=3{t.Fatalf("expected 3 prompts, got %d",len(r.List()))}
 if _,err:=r.Get("diagnose_compiler_errors",map[string]string{"source":"x"});err==nil{t.Fatal("expected required argument error")}
 result,err:=r.Get("diagnose_compiler_errors",map[string]string{"source":"void setup(){}","diagnostics":"missing symbol"})
 if err!=nil{t.Fatal(err)}
 if len(result.Messages)!=1||!strings.Contains(result.Messages[0].Content["text"].(string),"missing symbol"){t.Fatalf("unexpected prompt result: %+v",result)}
}
func TestDefaultResourcesReadSafeData(t *testing.T){
 r:=NewResourceRegistry();fake:=&resourceTestCLI{}
 if err:=RegisterDefaultResources(r,fake);err!=nil{t.Fatal(err)}
 if len(r.List())!=3{t.Fatalf("expected 3 resources, got %d",len(r.List()))}
 item,err:=r.Read(context.Background(),"luma://toolchain/arduino-cli");if err!=nil{t.Fatal(err)}
 if !strings.Contains(item.Text,"test-version")||strings.Contains(item.Text,"password"){t.Fatalf("unexpected resource data: %s",item.Text)}
 if _,err:=r.Read(context.Background(),"luma://unknown");err==nil{t.Fatal("expected unknown resource error")}
}
type resourceTestCLI struct{}
func(*resourceTestCLI)Version(context.Context)(string,error){return "test-version",nil}
func(*resourceTestCLI)ListBoards(context.Context)([]board.Board,error){return []board.Board{},nil}
func(*resourceTestCLI)Compile(context.Context,string,string)(build.Result,error){return build.Result{},nil}
func(*resourceTestCLI)Upload(context.Context,string,string,string)(build.Result,error){return build.Result{},nil}
func(*resourceTestCLI)InstallLibrary(context.Context,library.Dependency)error{return nil}

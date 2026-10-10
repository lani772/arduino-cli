package application

import (
 "context"
 "errors"
 "testing"

 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/board"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/build"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/diagnostic"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/library"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
)

type loopCLI struct { results []build.Result; calls int }
func (c *loopCLI) Version(context.Context)(string,error){return "fake",nil}
func (c *loopCLI) ListBoards(context.Context)([]board.Board,error){return nil,nil}
func (c *loopCLI) Compile(context.Context,string,string)(build.Result,error){
 i:=c.calls;c.calls++;if i>=len(c.results){return c.results[len(c.results)-1],nil};return c.results[i],nil
}
func (c *loopCLI) Upload(context.Context,string,string,string)(build.Result,error){return build.Result{},errors.New("unused")}
func (c *loopCLI) InstallLibrary(context.Context,library.Dependency)error{return nil}

type memoryProjectFiles struct { data map[string]string; replaces int; failReplace bool }
func (f *memoryProjectFiles) Snapshot(context.Context,project.Project)(map[string]string,error){return cloneSnapshot(f.data),nil}
func (f *memoryProjectFiles) Replace(_ context.Context,_ project.Project,next map[string]string)error{
 if f.failReplace{return errors.New("write failed")};f.data=cloneSnapshot(next);f.replaces++;return nil
}
type loopEngineer struct { patches []repair.Patch; calls int }
func (e *loopEngineer) GenerateFirmware(context.Context,string)(string,error){return "",errors.New("unused")}
func (e *loopEngineer) Repair(context.Context,string,repair.Patch)(string,error){return "",errors.New("unused")}
func (e *loopEngineer) Diagnose(context.Context,[]diagnostic.Diagnostic,string)(repair.Patch,error){
 p:=e.patches[e.calls];e.calls++;return p,nil
}
func failedBuild()build.Result{return build.Result{ID:"b1",Status:build.StatusFailed,ExitCode:1,Output:"src/main.ino:4:2: error: expected ';'"}}
func passedBuild()build.Result{return build.Result{ID:"b2",Status:build.StatusPassed,ExitCode:0}}

func TestRepairLoopRepairsAndRebuilds(t *testing.T){
 cli:=&loopCLI{results:[]build.Result{failedBuild(),passedBuild()}}
 files:=&memoryProjectFiles{data:map[string]string{"src/main.ino":"old source"}}
 engineer:=&loopEngineer{patches:[]repair.Patch{{ID:"p1",Description:"fix syntax",Status:repair.StatusProposed,Files:[]repair.FileChange{{Path:"src/main.ino",Operation:repair.OperationUpdate,Content:"fixed source",ExpectedSHA:repair.ContentSHA("old source")}}}}}
 svc:=RepairLoopService{Builder:BuildService{CLI:cli},Engineer:engineer,Parser:CompilerDiagnosticParser{},Files:files,MaxAttempts:2}
 result,err:=svc.Run(context.Background(),project.Project{ID:"p",Name:"lamp",WorkspaceID:"ws",Source:"/tmp/p"},"esp32:esp32:esp32")
 if err!=nil{t.Fatal(err)}
 if !result.Recovered||result.Attempts!=1||cli.calls!=2{t.Fatalf("unexpected result=%+v calls=%d",result,cli.calls)}
 if files.data["src/main.ino"]!="fixed source"{t.Fatalf("patch not persisted: %#v",files.data)}
}
func TestRepairLoopRestoresOriginalAfterExhaustion(t *testing.T){
 cli:=&loopCLI{results:[]build.Result{failedBuild(),failedBuild(),failedBuild()}}
 files:=&memoryProjectFiles{data:map[string]string{"src/main.ino":"original"}}
 engineer:=&loopEngineer{patches:[]repair.Patch{
  {ID:"p1",Description:"fix one",Status:repair.StatusProposed,Files:[]repair.FileChange{{Path:"src/main.ino",Operation:repair.OperationUpdate,Content:"try one",ExpectedSHA:repair.ContentSHA("original")}}},
  {ID:"p2",Description:"fix two",Status:repair.StatusProposed,Files:[]repair.FileChange{{Path:"src/main.ino",Operation:repair.OperationUpdate,Content:"try two",ExpectedSHA:repair.ContentSHA("try one")}}},
 }}
 svc:=RepairLoopService{Builder:BuildService{CLI:cli},Engineer:engineer,Parser:CompilerDiagnosticParser{},Files:files,MaxAttempts:2}
 _,err:=svc.Run(context.Background(),project.Project{ID:"p",Name:"lamp",WorkspaceID:"ws",Source:"/tmp/p"},"esp32:esp32:esp32")
 if !errors.Is(err,ErrRepairAttemptsExhausted){t.Fatalf("expected exhaustion, got %v",err)}
 if files.data["src/main.ino"]!="original"{t.Fatalf("expected original restored: %#v",files.data)}
}
func TestRepairLoopRefusesUnparsedFailure(t *testing.T){
 cli:=&loopCLI{results:[]build.Result{{ID:"b",Status:build.StatusFailed,ExitCode:1,Output:"build failed for unknown reason"}}}
 files:=&memoryProjectFiles{data:map[string]string{"main.ino":"source"}}
 svc:=RepairLoopService{Builder:BuildService{CLI:cli},Engineer:&loopEngineer{},Parser:CompilerDiagnosticParser{},Files:files,MaxAttempts:1}
 _,err:=svc.Run(context.Background(),project.Project{ID:"p",Name:"lamp",WorkspaceID:"ws",Source:"/tmp/p"},"fqbn")
 if err==nil{t.Fatal("expected safe refusal")}
 if files.replaces!=0{t.Fatal("unexpected source mutation")}
}

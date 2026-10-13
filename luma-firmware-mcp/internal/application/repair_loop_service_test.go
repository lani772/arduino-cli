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

func TestRepairLoopProposesWithoutMutatingSource(t *testing.T){
 cli:=&loopCLI{results:[]build.Result{failedBuild()}}
 files:=&memoryProjectFiles{data:map[string]string{"src/main.ino":"old source"}}
 engineer:=&loopEngineer{patches:[]repair.Patch{{ID:"p1",Description:"fix syntax",Status:repair.StatusProposed,Files:[]repair.FileChange{{Path:"src/main.ino",Operation:repair.OperationUpdate,Content:"fixed source",ExpectedSHA:repair.ContentSHA("old source")}}}}}
 svc:=RepairLoopService{Builder:BuildService{CLI:cli},Engineer:engineer,Parser:CompilerDiagnosticParser{},Files:files,MaxAttempts:2}
 result,err:=svc.Run(context.Background(),project.Project{ID:"p",Name:"lamp",WorkspaceID:"ws",Source:"/tmp/p"},"esp32:esp32:esp32")
 if !errors.Is(err,ErrPatchApprovalRequired){t.Fatalf("expected approval boundary, got %v",err)}
 if len(result.Patches)!=1||result.Attempts!=0{t.Fatalf("unexpected proposal result: %+v",result)}
 if files.data["src/main.ino"]!="old source"||files.replaces!=0{t.Fatalf("proposal changed active source: %#v",files.data)}
}
func TestRepairLoopStopsAfterSingleProposalAndNeverRetriesWithoutApproval(t *testing.T) {
 cli := &loopCLI{results: []build.Result{failedBuild(), passedBuild()}}
 files := &memoryProjectFiles{data: map[string]string{"src/main.ino": "original"}}
 engineer := &loopEngineer{patches: []repair.Patch{{ID:"p1", Description:"fix one", Status:repair.StatusProposed, Files:[]repair.FileChange{{Path:"src/main.ino", Operation:repair.OperationUpdate, Content:"candidate", ExpectedSHA:repair.ContentSHA("original")}}}}}
 svc := RepairLoopService{Builder:BuildService{CLI:cli}, Engineer:engineer, Parser:CompilerDiagnosticParser{}, Files:files, MaxAttempts:3}
 result, err := svc.Run(context.Background(), project.Project{ID:"p", Name:"lamp", WorkspaceID:"ws", Source:"/tmp/p"}, "esp32:esp32:esp32")
 if !errors.Is(err, ErrPatchApprovalRequired) { t.Fatalf("expected approval boundary, got %v", err) }
 if cli.calls != 1 || engineer.calls != 1 { t.Fatalf("repair continued without approval: builds=%d AI=%d", cli.calls, engineer.calls) }
 if len(result.Patches) != 1 || files.replaces != 0 || files.data["src/main.ino"] != "original" { t.Fatalf("unexpected source mutation/result: %+v %#v", result, files.data) }
}

func TestRepairLoopRejectsMalformedPatchWithoutMutation(t *testing.T) {
 cli := &loopCLI{results: []build.Result{failedBuild()}}
 files := &memoryProjectFiles{data: map[string]string{"main.ino":"original"}}
 engineer := &loopEngineer{patches: []repair.Patch{{ID:"bad", Description:"unsafe path", Status:repair.StatusProposed, Files:[]repair.FileChange{{Path:"../outside.ino", Operation:repair.OperationCreate, Content:"unsafe"}}}}}
 svc := RepairLoopService{Builder:BuildService{CLI:cli}, Engineer:engineer, Parser:CompilerDiagnosticParser{}, Files:files}
 _, err := svc.Run(context.Background(), project.Project{ID:"p", Name:"lamp", WorkspaceID:"ws", Source:"/tmp/p"}, "esp32:esp32:esp32")
 if err == nil { t.Fatal("expected malformed patch rejection") }
 if files.replaces != 0 || files.data["main.ino"] != "original" { t.Fatal("malformed patch mutated source") }
}

type failingLoopEngineer struct { err error }
func (e *failingLoopEngineer) GenerateFirmware(context.Context, string) (string, error) { return "", e.err }
func (e *failingLoopEngineer) Repair(context.Context, string, repair.Patch) (string, error) { return "", e.err }
func (e *failingLoopEngineer) Diagnose(context.Context, []diagnostic.Diagnostic, string) (repair.Patch, error) { return repair.Patch{}, e.err }

func TestRepairLoopPropagatesEngineerFailureWithoutMutation(t *testing.T) {
 cli := &loopCLI{results: []build.Result{failedBuild()}}
 files := &memoryProjectFiles{data: map[string]string{"main.ino":"original"}}
 svc := RepairLoopService{Builder:BuildService{CLI:cli}, Engineer:&failingLoopEngineer{err:errors.New("AI unavailable")}, Parser:CompilerDiagnosticParser{}, Files:files}
 _, err := svc.Run(context.Background(), project.Project{ID:"p", Name:"lamp", WorkspaceID:"ws", Source:"/tmp/p"}, "esp32:esp32:esp32")
 if err == nil { t.Fatal("expected engineer failure") }
 if files.replaces != 0 || files.data["main.ino"] != "original" { t.Fatal("engineer failure mutated source") }
}

func TestRepairLoopHonorsCancellationBeforeBuild(t *testing.T) {
 cli := &loopCLI{results: []build.Result{failedBuild()}}
 files := &memoryProjectFiles{data: map[string]string{"main.ino":"original"}}
 ctx, cancel := context.WithCancel(context.Background())
 cancel()
 svc := RepairLoopService{Builder:BuildService{CLI:cli}, Engineer:&loopEngineer{}, Parser:CompilerDiagnosticParser{}, Files:files}
 _, err := svc.Run(ctx, project.Project{ID:"p", Name:"lamp", WorkspaceID:"ws", Source:"/tmp/p"}, "esp32:esp32:esp32")
 if !errors.Is(err, context.Canceled) { t.Fatalf("expected cancellation, got %v", err) }
 if cli.calls != 0 || files.replaces != 0 { t.Fatalf("cancelled loop did work: builds=%d replaces=%d", cli.calls, files.replaces) }
}
func TestRepairLoopRefusesUnparsedFailure(t *testing.T){
 cli:=&loopCLI{results:[]build.Result{{ID:"b",Status:build.StatusFailed,ExitCode:1,Output:"build failed for unknown reason"}}}
 files:=&memoryProjectFiles{data:map[string]string{"main.ino":"source"}}
 svc:=RepairLoopService{Builder:BuildService{CLI:cli},Engineer:&loopEngineer{},Parser:CompilerDiagnosticParser{},Files:files,MaxAttempts:1}
 _,err:=svc.Run(context.Background(),project.Project{ID:"p",Name:"lamp",WorkspaceID:"ws",Source:"/tmp/p"},"fqbn")
 if err==nil{t.Fatal("expected safe refusal")}
 if files.replaces!=0{t.Fatal("unexpected source mutation")}
}

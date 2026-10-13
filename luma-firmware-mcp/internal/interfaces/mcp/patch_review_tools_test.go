package mcp

import (
 "context"
 "encoding/json"
 "errors"
	"path/filepath"
 "testing"

 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/application"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/board"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/build"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/library"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/infrastructure/reviewstore"
)

type reviewProjectStore struct { p project.Project }
func (s reviewProjectStore) Create(context.Context,project.Project)error{return errors.New("unused")}
func (s reviewProjectStore) Get(_ context.Context,id string)(project.Project,error){if id!=s.p.ID{return project.Project{},errors.New("not found")};return s.p,nil}

type reviewFiles struct { data map[string]string; writes int; fail bool }
func(f *reviewFiles)Snapshot(context.Context,project.Project)(map[string]string,error){out:=map[string]string{};for k,v:=range f.data{out[k]=v};return out,nil}
func(f *reviewFiles)Replace(_ context.Context,_ project.Project,next map[string]string)error{if f.fail{return errors.New("write failed")};f.data=map[string]string{};for k,v:=range next{f.data[k]=v};f.writes++;return nil}

type reviewCLI struct { builds int; result build.Result }
func(c *reviewCLI)Version(context.Context)(string,error){return "test",nil}
func(c *reviewCLI)ListBoards(context.Context)([]board.Board,error){return nil,nil}
func(c *reviewCLI)Compile(context.Context,string,string)(build.Result,error){c.builds++;return c.result,nil}
func(c *reviewCLI)Upload(context.Context,string,string,string)(build.Result,error){return build.Result{},errors.New("unused")}
func(c *reviewCLI)InstallLibrary(context.Context,library.Dependency)error{return nil}
var _ application.ProjectStore = reviewProjectStore{}
var _ application.ProjectFiles = (*reviewFiles)(nil)
var _ application.ArduinoCLI = (*reviewCLI)(nil)

func reviewTestSetup(t *testing.T)(*Registry,*reviewFiles,*reviewCLI){
 t.Helper()
 r:=NewRegistry()
 files:=&reviewFiles{data:map[string]string{"main.ino":"before"}}
 cli:=&reviewCLI{result:build.Result{ID:"build-1",Status:build.StatusPassed,ExitCode:0}}
 p:=project.Project{ID:"project-1",Name:"lamp",WorkspaceID:"ws",Source:"/workspace/lamp"}
 if err:=RegisterPatchReviewTools(r,reviewProjectStore{p:p},application.PatchApprovalService{Files:files},application.BuildService{CLI:cli});err!=nil{t.Fatal(err)}
 return r,files,cli
}
func callReviewTool(t *testing.T,r *Registry,name string,args any)map[string]any{
 t.Helper();raw,err:=json.Marshal(args);if err!=nil{t.Fatal(err)}
 result,err:=r.Call(context.Background(),name,raw);if err!=nil{t.Fatal(err)}
 if result.IsError{t.Fatalf("tool %s returned error: %v",name,result.Content)}
 var value map[string]any
 if err:=json.Unmarshal([]byte(result.Content[0]["text"].(string)),&value);err!=nil{t.Fatal(err)}
 return value
}
func reviewPatch()repair.Patch{return repair.Patch{ID:"patch-1",Description:"update lamp firmware",Status:repair.StatusProposed,Files:[]repair.FileChange{{Path:"main.ino",Operation:repair.OperationUpdate,Content:"after",ExpectedSHA:repair.ContentSHA("before")}}}}
func proposeReviewPatch(t *testing.T,r *Registry){
 t.Helper();v:=callReviewTool(t,r,"luma_repair_patch_propose",map[string]any{"project_id":"project-1","fqbn":"esp32:esp32:esp32","patch":reviewPatch()})
 if v["status"]!="awaiting_explicit_decision"{t.Fatalf("unexpected proposal status: %#v",v)}
}
func TestPatchReviewRestoresPendingProposalAfterRestart(t *testing.T) {
 t.Helper()
 path:=filepath.Join(t.TempDir(),"review.json")
 journal,err:=reviewstore.New(path);if err!=nil{t.Fatal(err)}
 r:=NewRegistry()
 files:=&reviewFiles{data:map[string]string{"main.ino":"before"}}
 cli:=&reviewCLI{result:build.Result{ID:"build-1",Status:build.StatusPassed,ExitCode:0}}
 p:=project.Project{ID:"project-1",Name:"lamp",WorkspaceID:"ws",Source:"/workspace/lamp"}
 if err:=RegisterPatchReviewToolsWithJournal(r,reviewProjectStore{p:p},application.PatchApprovalService{Files:files},application.BuildService{CLI:cli},journal);err!=nil{t.Fatal(err)}
 proposeReviewPatch(t,r)

 reloadedJournal,err:=reviewstore.New(path);if err!=nil{t.Fatal(err)}
 restarted:=NewRegistry()
 if err:=RegisterPatchReviewToolsWithJournal(restarted,reviewProjectStore{p:p},application.PatchApprovalService{Files:files},application.BuildService{CLI:cli},reloadedJournal);err!=nil{t.Fatal(err)}
 v:=callReviewTool(t,restarted,"luma_repair_patch_decide",map[string]any{"patch_id":"patch-1","approved":false})
 if v["status"]!="rejected"{t.Fatalf("persisted proposal was not available after restart: %#v",v)}
 if files.writes!=0||cli.builds!=0||files.data["main.ino"]!="before"{t.Fatal("restored rejection caused side effects")}
}

func TestPatchReviewRequiresExplicitDecision(t *testing.T) {
 r,files,cli:=reviewTestSetup(t)
 proposeReviewPatch(t,r)
 result,err:=r.Call(context.Background(),"luma_repair_patch_decide",mustJSON(map[string]any{"patch_id":"patch-1"}))
 if err!=nil{t.Fatal(err)}
 if !result.IsError{t.Fatal("missing approved field must not be treated as an explicit rejection")}
 if files.writes!=0||cli.builds!=0||files.data["main.ino"]!="before"{t.Fatal("missing decision caused side effects")}
 v:=callReviewTool(t,r,"luma_repair_patch_decide",map[string]any{"patch_id":"patch-1","approved":false})
 if v["status"]!="rejected"{t.Fatalf("proposal should remain pending for explicit decision: %#v",v)}
}

func TestPatchReviewProposalAndRejectionDoNotWriteOrBuild(t *testing.T){
 r,files,cli:=reviewTestSetup(t);proposeReviewPatch(t,r)
 if files.writes!=0||cli.builds!=0||files.data["main.ino"]!="before"{t.Fatal("proposal unexpectedly wrote files or built firmware")}
 v:=callReviewTool(t,r,"luma_repair_patch_decide",map[string]any{"patch_id":"patch-1","approved":false})
 if v["status"]!="rejected"{t.Fatalf("unexpected decision: %#v",v)}
 if files.writes!=0||cli.builds!=0||files.data["main.ino"]!="before"{t.Fatal("rejection wrote files or built firmware")}
}
func TestPatchReviewApprovalWritesThenBuilds(t *testing.T){
 r,files,cli:=reviewTestSetup(t);proposeReviewPatch(t,r)
 v:=callReviewTool(t,r,"luma_repair_patch_decide",map[string]any{"patch_id":"patch-1","approved":true})
 if v["status"]!="approved_and_applied"{t.Fatalf("unexpected decision: %#v",v)}
 if files.writes!=1||files.data["main.ino"]!="after"{t.Fatalf("approved patch not applied: %#v",files.data)}
 if cli.builds!=1{t.Fatalf("expected one post-approval build, got %d",cli.builds)}
}
func TestPatchReviewRejectsStaleProposalWithoutBuild(t *testing.T){
 r,files,cli:=reviewTestSetup(t);proposeReviewPatch(t,r)
 files.data["main.ino"]="changed"
 result,err:=r.Call(context.Background(),"luma_repair_patch_decide",mustJSON(map[string]any{"patch_id":"patch-1","approved":true}))
 if err!=nil{t.Fatal(err)}
 if !result.IsError{t.Fatal("expected stale proposal to be rejected")}
 if cli.builds!=0||files.writes!=0||files.data["main.ino"]!="changed"{t.Fatal("stale proposal wrote files or triggered a build")}
}
func mustJSON(v any)json.RawMessage{b,_:=json.Marshal(v);return b}

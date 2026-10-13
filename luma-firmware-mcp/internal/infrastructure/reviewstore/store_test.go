package reviewstore

import (
 "path/filepath"
 "testing"
 "time"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
)

func TestProposalAndDecisionSurviveRestart(t *testing.T) {
 path:=filepath.Join(t.TempDir(),"review.json")
 s,err:=New(path);if err!=nil{t.Fatal(err)}
 p:=Proposal{Project:project.Project{ID:"p1",Name:"lamp",WorkspaceID:"default",Source:"/workspace/lamp"},FQBN:"esp32:esp32:esp32",Patch:repair.Patch{ID:"patch-1"},CreatedAt:time.Now().UTC()}
 if err:=s.SaveProposal("patch-1",p);err!=nil{t.Fatal(err)}
 reloaded,err:=New(path);if err!=nil{t.Fatal(err)}
 if got:=reloaded.Snapshot();got.Proposals["patch-1"].Project.ID!="p1"||len(got.Audit)!=1{t.Fatalf("proposal not restored: %+v",got)}
 if _,err:=reloaded.BeginDecision("patch-1","rejected");err!=nil{t.Fatal(err)}
 if err:=reloaded.CompleteDecision("patch-1","rejected","");err!=nil{t.Fatal(err)}
 again,err:=New(path);if err!=nil{t.Fatal(err)}
 state:=again.Snapshot()
 if state.Decisions["patch-1"]!="rejected"||len(state.Proposals)!=0||len(state.Audit)!=3{t.Fatalf("decision/audit not restored: %+v",state)}
	if got:=state.Audit[len(state.Audit)-1].ProjectID;got!="p1"{t.Fatalf("final decision audit lost project identity: %q",got)}
 if _,err:=again.BeginDecision("patch-1","approved");err==nil{t.Fatal("duplicate decision must be rejected")}
 if err:=again.SaveProposal("patch-1",p);err==nil{t.Fatal("decided patch id must not be reused")}
}

func TestApprovalFailureCanRestoreProposal(t *testing.T) {
 s,err:=New(filepath.Join(t.TempDir(),"review.json"));if err!=nil{t.Fatal(err)}
 p:=Proposal{Project:project.Project{ID:"p1",Name:"lamp",WorkspaceID:"default",Source:"/workspace/lamp"},FQBN:"esp32:esp32:esp32",Patch:repair.Patch{ID:"p2"}}
 if err:=s.SaveProposal("p2",p);err!=nil{t.Fatal(err)}
 if _,err:=s.BeginDecision("p2","approved");err!=nil{t.Fatal(err)}
 if err:=s.RestoreProposal("p2",p,"stale source");err!=nil{t.Fatal(err)}
 state:=s.Snapshot()
 if _,ok:=state.Proposals["p2"];!ok{t.Fatal("proposal not restored")}
 if _,ok:=state.Decisions["p2"];ok{t.Fatal("processing decision not cleared")}
}

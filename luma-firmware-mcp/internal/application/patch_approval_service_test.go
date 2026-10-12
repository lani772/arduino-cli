package application

import (
 "context"
 "errors"
 "testing"

 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
)

func TestPatchApprovalRequiresExplicitConsent(t *testing.T) {
 files:=&memoryProjectFiles{data:map[string]string{"main.ino":"before"}}
 patch:=repair.Patch{ID:"p",Description:"safe update",Status:repair.StatusProposed,Files:[]repair.FileChange{{Path:"main.ino",Operation:repair.OperationUpdate,Content:"after",ExpectedSHA:repair.ContentSHA("before")}}}
 err:=(PatchApprovalService{Files:files}).ApplyApproved(context.Background(),project.Project{ID:"p",Name:"lamp",WorkspaceID:"ws"} ,patch,false)
 if err==nil { t.Fatal("expected approval to be required") }
 if files.data["main.ino"]!="before"||files.replaces!=0 { t.Fatal("unapproved patch mutated project") }
}

func TestPatchApprovalAppliesOnlyAgainstCurrentHash(t *testing.T) {
 files:=&memoryProjectFiles{data:map[string]string{"main.ino":"changed since proposal"}}
 patch:=repair.Patch{ID:"p",Description:"stale update",Status:repair.StatusProposed,Files:[]repair.FileChange{{Path:"main.ino",Operation:repair.OperationUpdate,Content:"after",ExpectedSHA:repair.ContentSHA("before")}}}
 err:=(PatchApprovalService{Files:files}).ApplyApproved(context.Background(),project.Project{ID:"p",Name:"lamp",WorkspaceID:"ws"},patch,true)
 if err==nil { t.Fatal("expected stale hash rejection") }
 if files.data["main.ino"]!="changed since proposal"||files.replaces!=0 { t.Fatal("stale patch mutated project") }
}

func TestPatchApprovalAppliesAfterConsent(t *testing.T) {
 files:=&memoryProjectFiles{data:map[string]string{"main.ino":"before"}}
 patch:=repair.Patch{ID:"p",Description:"approved update",Status:repair.StatusProposed,Files:[]repair.FileChange{{Path:"main.ino",Operation:repair.OperationUpdate,Content:"after",ExpectedSHA:repair.ContentSHA("before")}}}
 err:=(PatchApprovalService{Files:files}).ApplyApproved(context.Background(),project.Project{ID:"p",Name:"lamp",WorkspaceID:"ws"},patch,true)
 if err!=nil { t.Fatal(err) }
 if files.data["main.ino"]!="after"||files.replaces!=1 { t.Fatalf("approved patch not persisted: %#v",files.data) }
}

func TestPatchApprovalHonorsCancellation(t *testing.T) {
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 files:=&memoryProjectFiles{data:map[string]string{"main.ino":"before"}}
 patch:=repair.Patch{ID:"p",Description:"cancelled",Status:repair.StatusProposed,Files:[]repair.FileChange{{Path:"main.ino",Operation:repair.OperationUpdate,Content:"after",ExpectedSHA:repair.ContentSHA("before")}}}
 err:=(PatchApprovalService{Files:files}).ApplyApproved(ctx,project.Project{ID:"p",Name:"lamp",WorkspaceID:"ws"},patch,true)
 if !errors.Is(err,context.Canceled) { t.Fatalf("expected cancellation, got %v",err) }
 if files.replaces!=0 { t.Fatal("cancelled patch was persisted") }
}

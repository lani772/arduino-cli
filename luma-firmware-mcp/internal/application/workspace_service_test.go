package application

import (
 "context"
 "testing"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/workspace"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/infrastructure/memory"
)

func TestWorkspaceServiceCreatesProjectOnlyForExistingWorkspace(t *testing.T) {
 ws:=memory.NewWorkspaces(); ps:=memory.NewProjects(); ctx:=context.Background()
 _=ws.Create(ctx,workspace.Workspace{ID:"ws-1",Path:"/tmp/ws"})
 svc:=WorkspaceService{Workspaces:ws,Projects:ps}
 err:=svc.CreateProject(ctx,project.Project{ID:"p-1",Name:"lamp",WorkspaceID:"ws-1"})
 if err!=nil{t.Fatal(err)}
 if _,ok:=ps.Items["p-1"];!ok{t.Fatal("project was not stored")}
}

func TestWorkspaceServiceRejectsMissingWorkspace(t *testing.T) {
 svc:=WorkspaceService{Workspaces:memory.NewWorkspaces(),Projects:memory.NewProjects()}
 err:=svc.CreateProject(context.Background(),project.Project{ID:"p-1",Name:"lamp",WorkspaceID:"missing"})
 if err==nil{t.Fatal("expected missing workspace error")}
}

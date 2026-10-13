package memory

import (
 "context"
 "fmt"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/workspace"
)

type Workspaces struct { Items map[string]workspace.Workspace }
func NewWorkspaces() *Workspaces { return &Workspaces{Items:map[string]workspace.Workspace{}} }
func (s *Workspaces) Create(_ context.Context,w workspace.Workspace) error {
 if err:=w.Validate(); err!=nil{return err}; if _,ok:=s.Items[w.ID];ok{return fmt.Errorf("workspace already exists: %s",w.ID)}
 s.Items[w.ID]=w; return nil
}
func (s *Workspaces) Get(_ context.Context,id string)(workspace.Workspace,error){w,ok:=s.Items[id];if !ok{return workspace.Workspace{},fmt.Errorf("workspace not found: %s",id)};return w,nil}

type Projects struct { Items map[string]project.Project }
func NewProjects() *Projects { return &Projects{Items:map[string]project.Project{}} }
func (s *Projects) Create(_ context.Context,p project.Project) error {
 if err:=p.Validate();err!=nil{return err};if _,ok:=s.Items[p.ID];ok{return fmt.Errorf("project already exists: %s",p.ID)}
 s.Items[p.ID]=p;return nil
}
func (s *Projects) Get(_ context.Context,id string)(project.Project,error){p,ok:=s.Items[id];if !ok{return project.Project{},fmt.Errorf("project not found: %s",id)};return p,nil}

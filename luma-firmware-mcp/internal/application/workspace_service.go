package application

import (
 "context"
 "fmt"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/workspace"
)

type WorkspaceStore interface {
 Create(context.Context, workspace.Workspace) error
 Get(context.Context, string) (workspace.Workspace, error)
}

type ProjectStore interface {
 Create(context.Context, project.Project) error
 Get(context.Context, string) (project.Project, error)
}

type WorkspaceService struct { Workspaces WorkspaceStore; Projects ProjectStore }

func (s WorkspaceService) CreateProject(ctx context.Context, p project.Project) error {
 if err := p.Validate(); err != nil { return err }
 if s.Workspaces == nil || s.Projects == nil { return fmt.Errorf("workspace service dependencies are required") }
 if _, err := s.Workspaces.Get(ctx, p.WorkspaceID); err != nil { return err }
 return s.Projects.Create(ctx, p)
}

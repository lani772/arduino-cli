package project

import "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/errors"

type Project struct { ID, Name, WorkspaceID, Source string }

func New(id, name, workspaceID string) (Project, error) {
 p := Project{ID:id, Name:name, WorkspaceID:workspaceID}
 if err := p.Validate(); err != nil { return Project{}, err }
 return p, nil
}

func (p Project) Validate() error {
 if p.ID == "" { return errors.Invalid("id", "project id is required") }
 if p.Name == "" { return errors.Invalid("name", "project name is required") }
 if p.WorkspaceID == "" { return errors.Invalid("workspace_id", "workspace id is required") }
 return nil
}

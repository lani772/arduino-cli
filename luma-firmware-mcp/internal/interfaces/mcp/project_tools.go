package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/application"
	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
)

// RegisterProjectTools adds a controlled way to persist projects under the configured workspace.
func RegisterProjectTools(r *Registry, projects application.ProjectStore) error {
	if r == nil || projects == nil { return fmt.Errorf("registry and project store are required") }
	return r.Register(RegisteredTool{
		Definition: Tool{
			Name: "luma_project_register",
			Description: "Register an existing firmware project directory inside the server-configured workspace. The source path must be inside that workspace; absolute paths outside it are rejected. This does not compile, modify, or upload firmware.",
			InputSchema: objectSchema(map[string]any{
				"project_id": map[string]any{"type":"string"},
				"name": map[string]any{"type":"string"},
				"source_path": map[string]any{"type":"string","description":"Path relative to the configured workspace, or an absolute path within it."},
			}, "project_id", "name", "source_path"),
		},
		Handler: func(ctx context.Context, raw json.RawMessage) (any, error) {
			var args struct {
				ID string `json:"project_id"`
				Name string `json:"name"`
				Source string `json:"source_path"`
			}
			if err := json.Unmarshal(raw, &args); err != nil { return nil, fmt.Errorf("invalid project registration arguments: %w", err) }
			args.ID, args.Name, args.Source = strings.TrimSpace(args.ID), strings.TrimSpace(args.Name), strings.TrimSpace(args.Source)
			if args.ID == "" || args.Name == "" || args.Source == "" { return nil, fmt.Errorf("project_id, name, and source_path are required") }
			p := project.Project{ID:args.ID, Name:args.Name, WorkspaceID:"default", Source:args.Source}
			if err := projects.Create(ctx, p); err != nil { return nil, fmt.Errorf("register project: %w", err) }
			stored, err := projects.Get(ctx, p.ID)
			if err != nil { return nil, fmt.Errorf("reload registered project: %w", err) }
			return map[string]any{"status":"registered","project_id":stored.ID,"name":stored.Name,"source_path":stored.Source,"workspace_id":stored.WorkspaceID,"files_modified":false,"build_triggered":false}, nil
		},
	})
}

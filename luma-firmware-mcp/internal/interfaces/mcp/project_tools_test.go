package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/infrastructure/filedb"
)

func TestProjectRegistrationToolPersistsWorkspaceRelativeProject(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "projects", "lamp")
	if err := os.MkdirAll(source, 0700); err != nil { t.Fatal(err) }
	store, err := filedb.NewProjectStore(root, "")
	if err != nil { t.Fatal(err) }
	registry := NewRegistry()
	if err := RegisterProjectTools(registry, store); err != nil { t.Fatal(err) }
	raw := json.RawMessage(`{"project_id":"lamp-1","name":"Lamp","source_path":"projects/lamp"}`)
	result, err := registry.Call(context.Background(), "luma_project_register", raw)
	if err != nil { t.Fatal(err) }
	if result.IsError { t.Fatalf("registration failed: %#v", result.Content) }
	var response map[string]any
	if err := json.Unmarshal([]byte(result.Content[0]["text"].(string)), &response); err != nil { t.Fatal(err) }
	if response["status"] != "registered" || response["source_path"] != source { t.Fatalf("unexpected response: %#v", response) }
	reloaded, err := filedb.NewProjectStore(root, "")
	if err != nil { t.Fatal(err) }
	if _, err := reloaded.Get(context.Background(), "lamp-1"); err != nil { t.Fatalf("project was not persisted: %v", err) }
}

func TestProjectRegistrationToolRejectsOutsideWorkspace(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	store, err := filedb.NewProjectStore(root, "")
	if err != nil { t.Fatal(err) }
	registry := NewRegistry()
	if err := RegisterProjectTools(registry, store); err != nil { t.Fatal(err) }
	raw, _ := json.Marshal(map[string]string{"project_id":"escape","name":"Escape","source_path":outside})
	result, err := registry.Call(context.Background(), "luma_project_register", raw)
	if err != nil { t.Fatal(err) }
	if !result.IsError { t.Fatal("expected outside-workspace registration to fail") }
	if _, err := store.Get(context.Background(), "escape"); err == nil { t.Fatal("outside project was persisted") }
}

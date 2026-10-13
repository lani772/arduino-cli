// Package filedb stores project metadata durably in a workspace-confined JSON file.
package filedb

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
)

// ProjectStore persists projects and ensures each source directory stays under Root.
type ProjectStore struct {
	Root string
	MetadataPath string
	mu sync.Mutex
	items map[string]project.Project
}

func NewProjectStore(root, metadataPath string) (*ProjectStore, error) {
	if strings.TrimSpace(root) == "" { return nil, fmt.Errorf("workspace root is required") }
	absRoot, err := filepath.Abs(root); if err != nil { return nil, err }
	rootInfo, err := os.Lstat(absRoot); if err != nil { return nil, fmt.Errorf("inspect workspace root: %w", err) }
	if !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 { return nil, fmt.Errorf("workspace root must be a real directory") }
	realRoot, err := filepath.EvalSymlinks(absRoot); if err != nil { return nil, err }
	if strings.TrimSpace(metadataPath) == "" { metadataPath = filepath.Join(realRoot, ".luma", "projects.json") }
	if !filepath.IsAbs(metadataPath) { metadataPath = filepath.Join(realRoot, metadataPath) }
	metadataPath, err = filepath.Abs(metadataPath); if err != nil { return nil, err }
	parentRel, err := filepath.Rel(realRoot, filepath.Dir(metadataPath))
	if err != nil || parentRel == ".." || strings.HasPrefix(parentRel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("project metadata file must be inside the configured workspace")
	}
	store := &ProjectStore{Root:realRoot, MetadataPath:metadataPath, items:map[string]project.Project{}}
	if err := store.load(); err != nil { return nil, err }
	return store, nil
}

func (s *ProjectStore) load() error {
	s.mu.Lock(); defer s.mu.Unlock()
	data, err := os.ReadFile(s.MetadataPath)
	if os.IsNotExist(err) { return nil }
	if err != nil { return fmt.Errorf("read project metadata: %w", err) }
	var records []project.Project
	if err := json.Unmarshal(data, &records); err != nil { return fmt.Errorf("decode project metadata: %w", err) }
	for _, p := range records {
		if err := p.Validate(); err != nil { return fmt.Errorf("invalid stored project: %w", err) }
		if _, exists := s.items[p.ID]; exists { return fmt.Errorf("duplicate stored project id %q", p.ID) }
		if _, err := s.resolveSource(p.Source, false); err != nil { return fmt.Errorf("stored project %q has unsafe source: %w", p.ID, err) }
		s.items[p.ID] = p
	}
	return nil
}

func (s *ProjectStore) Create(ctx context.Context, p project.Project) error {
	if err := ctx.Err(); err != nil { return err }
	if err := p.Validate(); err != nil { return err }
	source, err := s.resolveSource(p.Source, true); if err != nil { return err }
	p.Source = source
	s.mu.Lock(); defer s.mu.Unlock()
	if err := ctx.Err(); err != nil { return err }
	if _, exists := s.items[p.ID]; exists { return fmt.Errorf("project already exists: %s", p.ID) }
	s.items[p.ID] = p
	if err := s.persistLocked(); err != nil { delete(s.items, p.ID); return err }
	return nil
}

func (s *ProjectStore) Get(ctx context.Context, id string) (project.Project, error) {
	if err := ctx.Err(); err != nil { return project.Project{}, err }
	s.mu.Lock(); defer s.mu.Unlock()
	p, ok := s.items[id]; if !ok { return project.Project{}, fmt.Errorf("project not found: %s", id) }
	return p, nil
}

func (s *ProjectStore) resolveSource(source string, mustExist bool) (string, error) {
	if strings.TrimSpace(source) == "" { return "", fmt.Errorf("project source is required") }
	if !filepath.IsAbs(source) { source = filepath.Join(s.Root, source) }
	abs, err := filepath.Abs(source); if err != nil { return "", err }
	rel, err := filepath.Rel(s.Root, abs)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("project source must be a child of configured workspace")
	}
	current := s.Root
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, statErr := os.Lstat(current)
		if os.IsNotExist(statErr) { if mustExist { return "", fmt.Errorf("project source does not exist") }; break }
		if statErr != nil { return "", statErr }
		if info.Mode()&os.ModeSymlink != 0 { return "", fmt.Errorf("symlink in project source path is not allowed") }
	}
	if mustExist {
		info, err := os.Stat(abs); if err != nil { return "", err }
		if !info.IsDir() { return "", fmt.Errorf("project source must be a directory") }
	}
	return abs, nil
}

func (s *ProjectStore) persistLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.MetadataPath), 0700); err != nil { return err }
	records := make([]project.Project, 0, len(s.items))
	for _, p := range s.items { records = append(records, p) }
	data, err := json.MarshalIndent(records, "", "  "); if err != nil { return err }
	tmp, err := os.CreateTemp(filepath.Dir(s.MetadataPath), ".projects-*.tmp"); if err != nil { return err }
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil { tmp.Close(); return err }
	if _, err := tmp.Write(data); err != nil { tmp.Close(); return err }
	if err := tmp.Sync(); err != nil { tmp.Close(); return err }
	if err := tmp.Close(); err != nil { return err }
	if err := os.Rename(tmpName, s.MetadataPath); err != nil { return fmt.Errorf("replace project metadata: %w", err) }
	return nil
}

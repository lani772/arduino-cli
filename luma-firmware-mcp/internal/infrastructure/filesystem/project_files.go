// Package filesystem implements workspace-confined project source persistence.
package filesystem

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
)

// ProjectFiles confines every project path to an explicitly configured workspace.
// Project.Source may be absolute (inside Root) or relative to Root.
type ProjectFiles struct { Root string }

// New resolves and validates the workspace root. The root must already exist.
func New(root string) (*ProjectFiles, error) {
	if strings.TrimSpace(root) == "" { return nil, fmt.Errorf("workspace root is required") }
	abs, err := filepath.Abs(root); if err != nil { return nil, fmt.Errorf("resolve workspace root: %w", err) }
	info, err := os.Lstat(abs); if err != nil { return nil, fmt.Errorf("inspect workspace root: %w", err) }
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() { return nil, fmt.Errorf("workspace root must be a real directory, not a symlink") }
	real, err := filepath.EvalSymlinks(abs); if err != nil { return nil, fmt.Errorf("resolve workspace root symlinks: %w", err) }
	return &ProjectFiles{Root: real}, nil
}

func (f *ProjectFiles) projectDir(p project.Project) (string, error) {
	if f == nil || f.Root == "" { return "", fmt.Errorf("filesystem workspace root is not configured") }
	if err := p.Validate(); err != nil { return "", err }
	source := strings.TrimSpace(p.Source)
	if source == "" { return "", fmt.Errorf("project source directory is required") }
	if !filepath.IsAbs(source) { source = filepath.Join(f.Root, source) }
	abs, err := filepath.Abs(source); if err != nil { return "", err }
	rel, err := filepath.Rel(f.Root, abs); if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("project source must be a child of the configured workspace")
	}
	// Check every existing component without following symlinks.
	current := f.Root
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, statErr := os.Lstat(current)
		if os.IsNotExist(statErr) { continue }
		if statErr != nil { return "", statErr }
		if info.Mode()&os.ModeSymlink != 0 { return "", fmt.Errorf("symlink in project path is not allowed: %s", current) }
	}
	return abs, nil
}

func (f *ProjectFiles) Snapshot(ctx context.Context, p project.Project) (map[string]string, error) {
	root, err := f.projectDir(p); if err != nil { return nil, err }
	info, err := os.Lstat(root); if err != nil { return nil, err }
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 { return nil, fmt.Errorf("project source must be a real directory") }
	files := map[string]string{}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if err := ctx.Err(); err != nil { return err }
		if walkErr != nil { return walkErr }
		if path == root { return nil }
		if entry.Type()&os.ModeSymlink != 0 { return fmt.Errorf("symlink in project tree is not allowed: %s", path) }
		if entry.IsDir() { return nil }
		info, err := entry.Info(); if err != nil { return err }
		if !info.Mode().IsRegular() { return fmt.Errorf("non-regular file in project tree is not allowed: %s", path) }
		rel, err := filepath.Rel(root, path); if err != nil { return err }
		data, err := os.ReadFile(path); if err != nil { return err }
		files[filepath.ToSlash(rel)] = string(data)
		return nil
	})
	if err != nil { return nil, fmt.Errorf("snapshot project: %w", err) }
	return files, nil
}

// Replace stages a complete source snapshot beside the project and swaps directories.
// If the second rename fails, it attempts to restore the original directory.
func (f *ProjectFiles) Replace(ctx context.Context, p project.Project, next map[string]string) error {
	root, err := f.projectDir(p); if err != nil { return err }
	info, err := os.Lstat(root); if err != nil { return err }
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 { return fmt.Errorf("project source must be a real directory") }
	if err := ctx.Err(); err != nil { return err }
	parent := filepath.Dir(root)
	stage, err := os.MkdirTemp(parent, ".luma-stage-"); if err != nil { return fmt.Errorf("create staging directory: %w", err) }
	defer os.RemoveAll(stage)
	for name, content := range next {
		if err := ctx.Err(); err != nil { return err }
		rel, err := safeRelativePath(name); if err != nil { return err }
		dst := filepath.Join(stage, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil { return err }
		file, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600); if err != nil { return err }
		_, writeErr := file.WriteString(content)
		syncErr := file.Sync()
		closeErr := file.Close()
		if writeErr != nil { return writeErr }; if syncErr != nil { return syncErr }; if closeErr != nil { return closeErr }
	}
	if err := ctx.Err(); err != nil { return err }
	backup, err := os.MkdirTemp(parent, ".luma-backup-"); if err != nil { return err }
	if err := os.Remove(backup); err != nil { return err }
	if err := os.Rename(root, backup); err != nil { return fmt.Errorf("move current project to backup: %w", err) }
	if err := os.Rename(stage, root); err != nil {
		restoreErr := os.Rename(backup, root)
		if restoreErr != nil { return fmt.Errorf("install staged project failed: %v; rollback failed: %w", err, restoreErr) }
		return fmt.Errorf("install staged project (original restored): %w", err)
	}
	if err := os.RemoveAll(backup); err != nil { return fmt.Errorf("new project installed but backup cleanup failed: %w", err) }
	return nil
}

func safeRelativePath(name string) (string, error) {
	if name == "" || strings.Contains(name, "\\") || filepath.IsAbs(name) || strings.ContainsRune(name, 0) {
		return "", fmt.Errorf("unsafe project-relative path %q", name)
	}
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe project-relative path %q", name)
	}
	return clean, nil
}

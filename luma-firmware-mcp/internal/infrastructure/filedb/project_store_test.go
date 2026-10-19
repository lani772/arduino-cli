package filedb

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
)

func TestProjectStorePersistsAndReloads(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "lamp")
	if err := os.Mkdir(source, 0700); err != nil { t.Fatal(err) }
	store, err := NewProjectStore(root, "")
	if err != nil { t.Fatal(err) }
	p := project.Project{ID:"lamp-1", Name:"Living room lamp", WorkspaceID:"default", Source:"lamp"}
	if err := store.Create(context.Background(), p); err != nil { t.Fatal(err) }
	reloaded, err := NewProjectStore(root, "")
	if err != nil { t.Fatal(err) }
	got, err := reloaded.Get(context.Background(), "lamp-1")
	if err != nil { t.Fatal(err) }
	if got.Source != source || got.Name != p.Name { t.Fatalf("unexpected persisted project: %+v", got) }
	if info, err := os.Stat(filepath.Join(root, ".luma", "projects.json")); err != nil || info.Mode().Perm()&0077 != 0 {
		t.Fatalf("metadata should be private (0600): info=%v err=%v", info, err)
	}
}

func TestProjectStoreRejectsTraversalAndSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	store, err := NewProjectStore(root, "")
	if err != nil { t.Fatal(err) }
	for _, source := range []string{"../outside", outside} {
		err := store.Create(context.Background(), project.Project{ID:"bad-" + filepath.Base(source), Name:"bad", WorkspaceID:"default", Source:source})
		if err == nil { t.Errorf("expected outside source rejection for %q", source) }
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(outside, link); err != nil { t.Skipf("symlink unavailable: %v", err) }
	if err := store.Create(context.Background(), project.Project{ID:"linked", Name:"linked", WorkspaceID:"default", Source:link}); err == nil {
		t.Fatal("expected symlink source rejection")
	}
}

func TestProjectStoreRejectsDuplicateAndHonorsCancellation(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "p"), 0700); err != nil { t.Fatal(err) }
	store, err := NewProjectStore(root, "")
	if err != nil { t.Fatal(err) }
	p := project.Project{ID:"p", Name:"project", WorkspaceID:"default", Source:"p"}
	if err := store.Create(context.Background(), p); err != nil { t.Fatal(err) }
	if err := store.Create(context.Background(), p); err == nil { t.Fatal("expected duplicate project rejection") }
	ctx, cancel := context.WithCancel(context.Background()); cancel()
	if err := store.Create(ctx, project.Project{ID:"q", Name:"cancelled", WorkspaceID:"default", Source:"p"}); err == nil { t.Fatal("expected cancellation") }
}

func TestProjectStoreRejectsSymlinkMetadataFile(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "projects.json")
	if err := os.WriteFile(outside, []byte("[]"), 0600); err != nil { t.Fatal(err) }
	metaDir := filepath.Join(root, ".luma")
	if err := os.Mkdir(metaDir, 0700); err != nil { t.Fatal(err) }
	if err := os.Symlink(outside, filepath.Join(metaDir, "projects.json")); err != nil { t.Skipf("symlink unavailable: %v", err) }
	if _, err := NewProjectStore(root, ""); err == nil { t.Fatal("expected symlink metadata rejection") }
}

package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
)

func fixture(t *testing.T) (*ProjectFiles, project.Project, string) {
	t.Helper()
	root := t.TempDir()
	projectDir := filepath.Join(root, "projects", "lamp")
	if err := os.MkdirAll(projectDir, 0700); err != nil { t.Fatal(err) }
	files, err := New(root); if err != nil { t.Fatal(err) }
	p := project.Project{ID:"p1", Name:"lamp", WorkspaceID:"ws1", Source:filepath.Join("projects", "lamp")}
	return files, p, projectDir
}

func TestSnapshotReadsRegularFilesWithSlashPaths(t *testing.T) {
	files, p, dir := fixture(t)
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0700); err != nil { t.Fatal(err) }
	if err := os.WriteFile(filepath.Join(dir, "src", "main.ino"), []byte("original"), 0600); err != nil { t.Fatal(err) }
	got, err := files.Snapshot(context.Background(), p); if err != nil { t.Fatal(err) }
	if got["src/main.ino"] != "original" { t.Fatalf("unexpected snapshot: %#v", got) }
}

func TestReplaceSwapsCompleteSnapshot(t *testing.T) {
	files, p, dir := fixture(t)
	if err := os.WriteFile(filepath.Join(dir, "old.ino"), []byte("old"), 0600); err != nil { t.Fatal(err) }
	if err := files.Replace(context.Background(), p, map[string]string{"src/main.ino":"new source", "config.txt":"ok"}); err != nil { t.Fatal(err) }
	got, err := files.Snapshot(context.Background(), p); if err != nil { t.Fatal(err) }
	if len(got) != 2 || got["src/main.ino"] != "new source" || got["config.txt"] != "ok" { t.Fatalf("replacement mismatch: %#v", got) }
	if _, err := os.Stat(filepath.Join(dir, "old.ino")); !os.IsNotExist(err) { t.Fatalf("deleted file remains, stat error=%v", err) }
}

func TestProjectPathTraversalRejected(t *testing.T) {
	files, p, _ := fixture(t)
	p.Source = filepath.Join("..", "..", "outside")
	if _, err := files.Snapshot(context.Background(), p); err == nil { t.Fatal("expected workspace traversal rejection") }
}

func TestPatchPathTraversalRejected(t *testing.T) {
	for _, name := range []string{"../outside", "/tmp/outside", "src/../../outside", "src\\outside", "."} {
		if _, err := safeRelativePath(name); err == nil { t.Errorf("expected unsafe path rejection for %q", name) }
	}
}

func TestSnapshotRejectsSymlinkInProjectTree(t *testing.T) {
	files, p, dir := fixture(t)
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0600); err != nil { t.Fatal(err) }
	if err := os.Symlink(outside, filepath.Join(dir, "linked.txt")); err != nil {
		if runtime.GOOS == "windows" { t.Skipf("symlink unavailable: %v", err) }
		t.Fatal(err)
	}
	if _, err := files.Snapshot(context.Background(), p); err == nil { t.Fatal("expected symlink rejection") }
}

func TestSnapshotHonorsCancellation(t *testing.T) {
	files, p, _ := fixture(t)
	ctx, cancel := context.WithCancel(context.Background()); cancel()
	if _, err := files.Snapshot(ctx, p); err == nil { t.Fatal("expected cancellation") }
}

func TestReplaceHonorsCancellationWithoutChangingProject(t *testing.T) {
	files, p, dir := fixture(t)
	if err := os.WriteFile(filepath.Join(dir, "main.ino"), []byte("original"), 0600); err != nil { t.Fatal(err) }
	ctx, cancel := context.WithCancel(context.Background()); cancel()
	if err := files.Replace(ctx, p, map[string]string{"main.ino":"changed"}); err == nil { t.Fatal("expected cancellation") }
	data, err := os.ReadFile(filepath.Join(dir, "main.ino")); if err != nil { t.Fatal(err) }
	if string(data) != "original" { t.Fatalf("cancelled replacement changed project: %q", data) }
}

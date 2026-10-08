package arduino

import (
 "context"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/board"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/build"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/library"
)

type FakeCLI struct {
 VersionValue string
 Boards []board.Board
 CompileResult build.Result
 UploadResult build.Result
 Installed []library.Dependency
 Compiles int
 Uploads int
}

func (f *FakeCLI) Version(context.Context) (string,error) { return f.VersionValue,nil }
func (f *FakeCLI) ListBoards(context.Context) ([]board.Board,error) { return append([]board.Board(nil),f.Boards...),nil }
func (f *FakeCLI) Compile(context.Context,string,string) (build.Result,error) { f.Compiles++; return f.CompileResult,nil }
func (f *FakeCLI) Upload(context.Context,string,string,string) (build.Result,error) { f.Uploads++; return f.UploadResult,nil }
func (f *FakeCLI) InstallLibrary(_ context.Context,d library.Dependency) error { f.Installed=append(f.Installed,d); return nil }

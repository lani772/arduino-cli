package application

import (
 "context"
 "testing"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/build"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/infrastructure/arduino"
)

func TestBuildServiceCompilesProject(t *testing.T){
 cli:=arduino.NewFakeCLI()
 cli.CompileResult=build.Result{ID:"compile",Status:build.StatusPassed,ExitCode:0}
 p:=project.Project{ID:"p-1",Name:"lamp",WorkspaceID:"ws-1",Source:"/tmp/ws/lamp"}
 r,err:= (BuildService{CLI:cli}).Compile(context.Background(),p,"esp32:esp32:esp32")
 if err!=nil{t.Fatal(err)}
 if r.Status!=build.StatusPassed{t.Fatalf("unexpected status %q",r.Status)}
 if cli.Compiles!=1{t.Fatalf("expected one compile, got %d",cli.Compiles)}
}

func TestBuildServiceRejectsMissingFQBN(t *testing.T){
 p:=project.Project{ID:"p",Name:"lamp",WorkspaceID:"ws",Source:"/tmp/lamp"}
 if _,err:=(BuildService{CLI:arduino.NewFakeCLI()}).Compile(context.Background(),p,"");err==nil{t.Fatal("expected FQBN validation error")}
}

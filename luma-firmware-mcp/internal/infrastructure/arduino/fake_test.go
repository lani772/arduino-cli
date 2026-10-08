package arduino

import (
 "context"
 "testing"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/build"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/library"
)

func TestFakeCLITracksOperations(t *testing.T) {
 f:=&FakeCLI{VersionValue:"test",CompileResult:build.Result{ID:"b",Status:build.StatusPassed},UploadResult:build.Result{ID:"u",Status:build.StatusPassed}}
 if v,_:=f.Version(context.Background());v!="test"{t.Fatal("unexpected version")}
 if _,err:=f.Compile(context.Background(),"src","esp32:esp32:esp32");err!=nil{t.Fatal(err)}
 if _,err:=f.Upload(context.Background(),"src","esp32:esp32:esp32","COM1");err!=nil{t.Fatal(err)}
 if err:=f.InstallLibrary(context.Background(),library.Dependency{Name:"MQTT"});err!=nil{t.Fatal(err)}
 if f.Compiles!=1||f.Uploads!=1||len(f.Installed)!=1{t.Fatal("fake adapter did not track operations")}
}

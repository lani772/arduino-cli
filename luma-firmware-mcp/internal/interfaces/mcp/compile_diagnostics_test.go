package mcp

import (
 "context"
 "encoding/json"
 "strings"
 "testing"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/build"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/infrastructure/arduino"
)

func TestCompileToolReturnsStructuredDiagnosticsOnCompilerFailure(t *testing.T) {
 cli := &arduino.FakeCLI{}
 cli.CompileResult = build.Result{ID:"compile", Status:build.StatusFailed, ExitCode:1, Output:"src/main.cpp:12:4: error: invalid conversion\nsrc/main.cpp:13:2: warning: unused variable"}
 r := NewRegistry()
 if err := RegisterArduinoTools(r, cli); err != nil { t.Fatal(err) }
 raw, _ := json.Marshal(map[string]string{"source_path":"/tmp/sketch","fqbn":"esp32:esp32:esp32"})
 result, err := r.Call(context.Background(), "arduino_compile", raw)
 if err != nil { t.Fatal(err) }
 if result.IsError { t.Fatal("compiler failure should be represented with structured build result") }
 body := result.Content[0]["text"].(string)
 for _, want := range []string{"diagnostics","diagnostic_summary","invalid conversion","unused variable","\"errors\":1","\"warnings\":1"} { if !strings.Contains(body,want) { t.Fatalf("response missing %q: %s",want,body) } }
}

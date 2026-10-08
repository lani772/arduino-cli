package application

import (
 "context"
 "testing"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/infrastructure/ai"
)
func TestAIServiceGenerate(t *testing.T){
 p:=&ai.FakeProvider{Response:AIResponse{Content:"void setup(){}"}}
 s:=AIService{Provider:p};got,err:=s.GenerateFirmware(context.Background(),"ESP32 lamp")
 if err!=nil||got!="void setup(){}"{t.Fatalf("got=%q err=%v",got,err)}
 if len(p.Requests)!=1||p.Requests[0].Operation!="generate_firmware"{t.Fatalf("unexpected request: %+v",p.Requests)}
}
func TestAIServiceDiagnoseAndRepair(t *testing.T){
 p:=&ai.FakeProvider{Response:AIResponse{Content:"replace badFunction() with digitalWrite()"}}
 s:=AIService{Provider:p}
 patch,err:=s.Diagnose(context.Background(),nil,"source");if err!=nil{t.Fatal(err)}
 if patch.Status!=repair.StatusProposed{t.Fatalf("unexpected patch: %+v",patch)}
 fixed,err:=s.Repair(context.Background(),"source",patch);if err!=nil{t.Fatal(err)}
 if fixed==""{t.Fatal("expected repaired source")};if len(p.Requests)!=2{t.Fatalf("requests=%d",len(p.Requests))}
}

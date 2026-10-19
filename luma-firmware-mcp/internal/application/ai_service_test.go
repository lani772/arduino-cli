package application

import (
	"context"
	"testing"

	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
)

type testAIProvider struct {
	response AIResponse
	requests []AIRequest
	err      error
}

func (p *testAIProvider) Complete(_ context.Context, request AIRequest) (AIResponse, error) {
	p.requests = append(p.requests, request)
	if p.err != nil {
		return AIResponse{}, p.err
	}
	return p.response, nil
}

func TestAIServiceGenerate(t *testing.T) {
	provider := &testAIProvider{response: AIResponse{Content: "void setup(){}"}}
	service := AIService{Provider: provider}

	got, err := service.GenerateFirmware(context.Background(), "ESP32 lamp")
	if err != nil || got != "void setup(){}" {
		t.Fatalf("got=%q err=%v", got, err)
	}
	if len(provider.requests) != 1 || provider.requests[0].Operation != "generate_firmware" {
		t.Fatalf("unexpected request: %+v", provider.requests)
	}
}

func TestAIServiceDiagnoseAndRepair(t *testing.T) {
	content := `{"id":"p1","description":"fix source","status":"proposed","files":[{"path":"main.ino","operation":"update","content":"fixed","expected_sha":"abc"}]}`
	provider := &testAIProvider{response: AIResponse{Content: content}}
	service := AIService{Provider: provider}

	patch, err := service.Diagnose(context.Background(), nil, "source")
	if err != nil {
		t.Fatal(err)
	}
	if patch.Status != repair.StatusProposed || len(patch.Files) != 1 {
		t.Fatalf("unexpected patch: %+v", patch)
	}

	fixed, err := service.Repair(context.Background(), "source", patch)
	if err != nil {
		t.Fatal(err)
	}
	if fixed == "" {
		t.Fatal("expected repaired source")
	}
	if len(provider.requests) != 2 {
		t.Fatalf("requests=%d", len(provider.requests))
	}
}

func TestAIServiceRejectsUnstructuredRepair(t *testing.T) {
	provider := &testAIProvider{response: AIResponse{Content: "replace badFunction()"}}
	if _, err := (AIService{Provider: provider}).Diagnose(context.Background(), nil, "source"); err == nil {
		t.Fatal("expected structured JSON validation error")
	}
}

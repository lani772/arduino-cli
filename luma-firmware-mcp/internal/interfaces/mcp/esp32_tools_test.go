package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestESP32ValidationToolReportsUnsafeGPIO(t *testing.T) {
	r := NewRegistry()
	if err := RegisterESP32Tools(r); err != nil {
		t.Fatal(err)
	}
	args, _ := json.Marshal(map[string]any{
		"board_fqbn": "esp32:esp32:esp32",
		"status_gpio": 13,
		"lamp_gpios": []int{14, 34},
	})
	result, err := r.Call(context.Background(), "luma_validate_esp32_target", args)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("validation should return an assessment: %+v", result)
	}
	if !strings.Contains(result.Content[0]["text"].(string), "input-only") {
		t.Fatalf("expected input-only GPIO error: %+v", result)
	}
}

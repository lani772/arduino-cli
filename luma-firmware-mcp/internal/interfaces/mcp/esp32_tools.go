package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/firmware"
)

// RegisterESP32Tools exposes conservative target and output-pin validation.
// It does not install cores, access serial ports, or flash devices.
func RegisterESP32Tools(r *Registry) error {
	return r.Register(RegisteredTool{
		Definition: Tool{
			Name: "luma_validate_esp32_target",
			Description: "Assess an ESP32 board FQBN and lamp/status GPIO assignments. Known classic ESP32 Dev Module restrictions are checked; other variants are marked unverified. This never installs a core or flashes a device.",
			InputSchema: objectSchema(map[string]any{
				"board_fqbn": map[string]any{"type": "string", "description": "Arduino fully qualified board name, e.g. esp32:esp32:esp32."},
				"status_gpio": map[string]any{"type": "integer", "description": "GPIO used for firmware status output."},
				"lamp_gpios": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}, "description": "Explicit GPIO assignments for lamp outputs."},
			}, "board_fqbn", "status_gpio", "lamp_gpios"),
		},
		Handler: func(_ context.Context, raw json.RawMessage) (any, error) {
			var input struct {
				BoardFQBN string `json:"board_fqbn"`
				StatusGPIO int `json:"status_gpio"`
				LampGPIOs []int `json:"lamp_gpios"`
			}
			if err := json.Unmarshal(raw, &input); err != nil {
				return nil, fmt.Errorf("invalid ESP32 target arguments: %w", err)
			}
			return firmware.AssessESP32Target(input.BoardFQBN, input.StatusGPIO, input.LampGPIOs), nil
		},
	})
}

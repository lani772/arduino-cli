# LUMA Firmware MCP

AI-assisted Arduino/ESP32 firmware engineering service isolated from the Arduino CLI root module.

Phase 1 establishes domain contracts, application ports, execution policy, HTTP health, and validation tests.

```bash
cd luma-firmware-mcp
gofmt -w .
go test ./...
go vet ./...
```

# LUMA Firmware MCP

AI-assisted Arduino/ESP32 firmware engineering service isolated from the Arduino CLI root module.

## Run

HTTP health endpoint (default):

```bash
go run ./cmd/server
# GET http://localhost:8080/healthz
```

MCP over stdio (newline-delimited JSON-RPC):

```bash
go run ./cmd/server -transport stdio
```

In stdio mode, stdout is reserved for protocol messages and logs go to stderr.

## MCP capabilities

Tools:
- `arduino_cli_version` — installed Arduino CLI version.
- `arduino_list_boards` — board discovery.
- `arduino_compile` — compile only; does not upload firmware.
- `luma_validate_firmware_spec` — validate LUMA lamp GPIO/specification input.
- `luma_validate_esp32_target` — assess an ESP32 FQBN and output GPIO assignments; unknown variants are explicitly marked unverified.

Resources:
- `luma://toolchain/arduino-cli` — CLI version.
- `luma://boards/detected` — current board discovery.
- `luma://firmware/specification-guidance` — safe specification constraints. This is guidance, not a live project record.

Prompts:
- `generate_esp32_lamp_firmware`
- `diagnose_compiler_errors`
- `propose_safe_repair`

The ESP32 target assessment includes a profile for `esp32:esp32:esp32` (classic ESP32 Dev Module). It checks reserved flash GPIOs, input-only GPIOs, duplicate assignments, and warns about boot-strapping/UART pins. Other ESP32 FQBNs receive basic checks and an explicit board-specific verification warning; this is not a substitute for the exact board schematic. Tool installation and flashing remain disabled. Compile executes the configured Arduino CLI against the supplied source path, so only trusted paths should be passed. Persistent project storage is not wired into the MCP transport.

## Validate

```bash
gofmt -w .
go test ./...
go vet ./...
```

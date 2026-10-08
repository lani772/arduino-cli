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
- `arduino_cli_version` — installed CLI version.
- `arduino_list_boards` — board discovery.
- `arduino_compile` — compile only; does not upload firmware.
- `luma_validate_firmware_spec` — validate LUMA lamp GPIO/specification input.

Resources:
- `luma://toolchain/arduino-cli` — CLI version.
- `luma://boards/detected` — current board discovery.
- `luma://firmware/specification-guidance` — safe specification constraints. This is guidance, not a live project record.

Prompts:
- `generate_esp32_lamp_firmware`
- `diagnose_compiler_errors`
- `propose_safe_repair`

The server does not currently expose live project records because persistent project storage is not wired into the MCP transport. Tool installation and flashing remain disabled by default. Compile executes the configured Arduino CLI against the supplied source path, so only trusted paths should be passed.

## Validate

```bash
gofmt -w .
go test ./...
go vet ./...
```

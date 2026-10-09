# LUMA Firmware MCP

AI-assisted Arduino/ESP32 firmware engineering service isolated from the Arduino CLI root module.

## Run

```bash
go run ./cmd/server
# GET http://localhost:8080/healthz
go run ./cmd/server -transport stdio
```

In stdio mode, stdout is reserved for protocol messages and logs go to stderr.

## MCP capabilities

Tools:
- `arduino_cli_version` — installed Arduino CLI version.
- `arduino_list_boards` — detected board candidates.
- `arduino_list_serial_devices` — read-only serial port discovery, including unmatched ports.
- `arduino_verify_device` — read-only verification report with separate port, board-candidate, and optional supplied upload-exit-code statuses; never claims runtime hardware verification.
- `arduino_compile` — compile an existing sketch.
- `arduino_upload_firmware` — explicit-port upload tool; requires `confirm_upload=true` and a policy that permits flashing (disabled by default).
- `luma_validate_firmware_spec` — validate LUMA lamp GPIO/specification input.
- `luma_validate_esp32_target` — assess ESP32 FQBN and output GPIO assignments; unknown variants are unverified.

Resources:
- `luma://toolchain/arduino-cli` — CLI version.
- `luma://boards/detected` — current board discovery.
- `luma://firmware/specification-guidance` — safe specification constraints.

Prompts:
- `generate_esp32_lamp_firmware`
- `diagnose_compiler_errors`
- `propose_safe_repair`

Serial discovery uses `arduino-cli board list --format json` and retains reported port address, label, protocol, and board candidates. Ports with no matching board are still reported; detection does not verify device identity. The tool does not open serial ports or reset devices. Core installation remains disabled. Upload requests are rejected by default because device flashing is disabled in the execution policy; enabling flashing requires an explicit policy change and each tool request must confirm the upload. Compile executes the configured Arduino CLI against the supplied source path, so only trusted paths should be passed. Persistent project storage is not wired into the MCP transport.

## Runtime handshake boundary

The hardware-independent `internal/domain/handshake` package validates protocol-v1 JSON records and provides `ReadIdentity` over an injected, context-aware line source. It skips a valid readiness record, enforces a caller-supplied deadline, and rejects malformed responses or challenge mismatches. Tests use fake sources and sample messages only. No real serial source, challenge generation, challenge transmission, cryptographic device authentication, or hardware verification is implemented. A source adapter must honor context cancellation; hardware access remains a separate disabled milestone.

## Validate

```bash
gofmt -w .
go test ./...
go vet ./...
```

Verification uses read-only serial discovery and does not open ports or reset the device. A successful upload exit code and a matching board candidate are evidence only; `hardware_verified` remains false until a live runtime handshake is implemented. The protocol contract is documented in [`docs/runtime-handshake.md`](docs/runtime-handshake.md).

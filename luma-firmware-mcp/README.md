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


## Supported-board status

See [supported boards and GPIO constraints](docs/supported-boards.md) for the exact known board profile, pin restrictions, warnings, and validation gates. Currently, only the exact classic ESP32 Dev Module FQBN has a registered pin profile. Other ESP32-family FQBNs are unverified and require board-specific review. Documentation and simulated tests do not constitute a successful compile or hardware validation.


## Explicit patch review API

The MCP package includes a review-first patch workflow through `RegisterPatchReviewTools`: `luma_repair_patch_propose` validates a candidate against a fresh source snapshot without writing, and `luma_repair_patch_decide` rejects it without side effects or applies it after explicit approval and then compiles. This registration API is covered by tests in `internal/interfaces/mcp/patch_review_tools_test.go`.

These tools are registered by the stdio server only when `LUMA_WORKSPACE_ROOT` points to an existing, trusted workspace directory. Project metadata is persisted to `LUMA_PROJECTS_FILE` when set (the path must be inside the workspace); otherwise it defaults to `<workspace>/.luma/projects.json`. The server registers `luma_project_register`, `luma_repair_patch_propose`, and `luma_repair_patch_decide`. Project registration accepts only existing directories under the configured workspace. If `LUMA_WORKSPACE_ROOT` is unset, patch-review and project-registration tools remain disabled.


## Secure patch-review runtime configuration (Task 44)

Create a dedicated workspace directory and start the MCP stdio transport with the workspace explicitly configured:

```bash
export LUMA_WORKSPACE_ROOT="/absolute/path/to/luma-workspace"
# Optional; must remain inside LUMA_WORKSPACE_ROOT:
export LUMA_PROJECTS_FILE="$LUMA_WORKSPACE_ROOT/.luma/projects.json"
go run ./cmd/server -transport stdio
```

Register an existing project using `luma_project_register` with a stable `project_id`, a display `name`, and a `source_path` relative to the workspace (or an absolute path inside it). Then use `luma_repair_patch_propose` and review the returned changes before calling `luma_repair_patch_decide` with an explicit decision. The project metadata store is JSON-backed and uses staged file replacement. Do not expose a workspace shared with untrusted local processes; path validation does not eliminate local filesystem race attacks. Pending patch proposals remain in memory and are not restored after restart (tracked as Task 45).


## Durable patch review and audit journal (Task 45)

When stdio mode is configured with `LUMA_WORKSPACE_ROOT`, pending repair proposals, in-progress/final decisions, and audit events are persisted to `<workspace>/.luma/patch-review.json` by default. Set `LUMA_REVIEW_JOURNAL` to choose a different path inside the same workspace. The journal uses a private file mode and staged JSON replacement. Proposals are restored when the server starts; a patch ID cannot be proposed again after a recorded decision. Rejection never writes source files. Approval records a processing state before source changes and records the final outcome afterward.

Audit events record proposal, decision start, final outcome, and restoration after failed application. The actor is currently a fixed service label (`mcp-client` or `system`), not an authenticated user identity. This is a local single-process journal, not a multi-instance transactional database. A crash during approval may leave a decision marked `processing`; reconcile source/build state before taking manual action. Tests have been added but not run in this environment.

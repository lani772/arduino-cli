# LUMA Firmware MCP — Supported boards and GPIO constraints

Status: **one documented GPIO profile; no physical board profile has been hardware-validated by this project yet.**

This page describes what the current validator knows. It is not a statement that firmware has compiled successfully or that a board has been tested on a bench.

## Support matrix

| Board target / FQBN | Validator status | GPIO assumptions | Release status |
| --- | --- | --- | --- |
| Classic ESP32 Dev Module — `esp32:esp32:esp32` | Exact profile recognized by `AssessESP32Target` | Conservative classic ESP32 GPIO constraints listed below | Not hardware-validated; compile and supervised board test remain required |
| Other ESP32-family FQBNs (including S2, S3, C3, C6 and vendor-specific targets) | Unknown profile; generic range/duplicate checks only | Do not reuse the classic pin map without checking the exact chip, module, dev-board schematic, and Arduino core variant | Unverified; do not treat generic validation as board approval |
| Non-ESP32 FQBN | Rejected by the ESP32 target assessment | Not applicable | Unsupported by this validator |

An FQBN is the fully qualified board name passed to Arduino CLI. A package-family prefix alone does not identify an exact pinout.

## Classic ESP32 Dev Module profile

The current profile is deliberately conservative:

- GPIOs **6–11** are rejected as reserved for flash on common classic ESP32 modules.
- GPIOs **34–39** are rejected for lamp and status outputs because they are input-only.
- GPIOs **0, 2, 5, 12 and 15** produce boot-strapping warnings. External circuitry can affect boot behavior; verify the actual board and attached driver circuit before using them.
- GPIOs **1 and 3** produce UART0 warnings because they are commonly used for serial logging and upload.
- A lamp GPIO cannot duplicate another lamp GPIO or the configured status GPIO.
- The current generic range check is 0–39. Passing that range check alone does not prove a pin is bonded out, electrically suitable, or safe on a particular module.

LUMA's current default status indicator is **GPIO13**. Lamp defaults have historically started at GPIO14, but allocation must still check for collisions, the exact board, and any other reserved functions. Do not assume GPIO14 onward is universally safe on every ESP32 variant.

### Electrical safety

ESP32 GPIO pins are logic-level signals, not lamp power outputs. Drive lamps or relays through an appropriately rated driver module/transistor/relay interface and suitable power supply; do not power a mains lamp directly from a GPIO. Keep mains wiring isolated and have it handled by a qualified person. During first firmware tests, disconnect lamp/relay loads or use a safe low-voltage test fixture.

## Validation behavior and limitations

- `luma_validate_esp32_target` assesses the requested FQBN and status/lamp GPIO assignments.
- `luma_validate_firmware_spec` applies known-profile constraints when the target exactly matches the registered classic ESP32 FQBN.
- Unknown ESP32 variants are not automatically assigned classic ESP32 restrictions. Their generic validation result must be described as **unverified**, not “supported.”
- The reference runtime handshake sketch is illustrative source until compiled against the selected board package. Simulator tests exercise the host-side protocol, not physical firmware behavior.
- A detected serial port, a board-list candidate, and a successful upload command are separate evidence levels. None alone proves that the intended firmware is running or authenticates a device.
- No physical serial adapter or live handshake is enabled by this documentation change.

## Required checklist before adding a board profile

1. Record the exact manufacturer, board/module name, chip variant, FQBN, Arduino core/package version, and relevant schematic.
2. Confirm usable output pins, input-only pins, flash/PSRAM reservations, strapping pins, UART pins, onboard LED pins, and any board-specific restrictions from authoritative board documentation.
3. Implement an exact-FQBN profile rather than inferring pin behavior from a broad `esp32:` prefix.
4. Add table-driven tests for valid outputs, invalid/reserved pins, duplicate assignments, status-pin collisions, and unknown FQBN behavior.
5. Compile the reference sketch using the pinned board package and record the exact command and result.
6. Run simulator/protocol tests and race detection successfully.
7. Perform a supervised test on a spare board with lamp loads disconnected or replaced by a safe low-voltage fixture. Record reset/open behavior, handshake response, timeout/disconnect handling, and cleanup.
8. Keep the profile marked unverified until those results are recorded. Do not claim authenticated identity without device-bound cryptographic verification.

## Current validation commands

From the `luma-firmware-mcp` directory:

```sh
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
```

These commands are required checks, not evidence that checks have already passed. At the time this document was added, execution of these commands and physical board testing remained unverified.

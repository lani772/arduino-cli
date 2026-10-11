# LUMA ESP32 runtime protocol reference sketch

This directory contains a reference implementation of the newline-delimited
JSON protocol documented in `luma-firmware-mcp/docs/runtime-handshake.md`.

## Target and dependency

- Target: ESP32 Arduino core.
- Serial speed: 115200 baud.
- Dependency: ArduinoJson 6.x or 7.x.
- Define `LUMA_FIRMWARE_VERSION` and `LUMA_PROJECT_ID` at build time to set
  the reported values. The defaults identify this reference project only.

## Behavior

At startup, the sketch emits one `luma.runtime` protocol-v1 `ready` record.
It accepts a single-line JSON `identify` request with exactly four fields:
`protocol`, `protocol_version`, `command`, and `challenge`. A valid request
produces an `identity` record echoing that challenge and reporting firmware
version, project ID, ESP32 eFuse MAC-derived device label, and uptime.

Input is bounded to 4096 bytes per line. Oversized lines are discarded through
the next newline. Malformed, unsupported, or invalid requests produce no
response. Output records end with LF.

## Safety and security boundary

The sketch does not configure, toggle, or otherwise control GPIOs or lamps.
It does not flash firmware, provision secrets, or authenticate the device.
The device label is self-reported and is not proof of ownership or authenticity.
A matching challenge only provides freshness against stale buffered responses;
it does not prevent an active serial peer from impersonating the device.
Do not describe this as authenticated hardware verification.

This is a source-level reference implementation. Compile it with the selected
ESP32 board package and ArduinoJson version, then test it using a serial
simulator before supervised hardware use. No compile or hardware test result
is implied by the presence of these files.

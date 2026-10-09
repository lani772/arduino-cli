# LUMA ESP32 Runtime Handshake — Protocol Design

Status: **parser and simulated-message unit tests implemented; live serial handshake not implemented**

## Purpose

Define the minimum evidence needed before a future verifier can report that LUMA firmware is running. Arduino CLI upload success, a serial port, and an Arduino CLI board candidate are not runtime proof. This document does not authorize the current MCP service to open serial ports or reset boards.

## Proposed transport

- Transport: newline-delimited JSON over USB serial, using a future explicitly enabled verifier adapter.
- Baud rate: 115200 by default; configurable only when both firmware and verifier agree.
- Firmware emits one JSON object per line and terminates each record with `\n`.
- The verifier must use a bounded timeout and bounded line size; it must ignore unrelated boot logs and malformed lines.
- No relay/lamp outputs may change as a side effect of a verification request.

## Protocol version 1

After initialization, firmware may emit a readiness record:

```json
{
  "protocol": "luma.runtime",
  "protocol_version": 1,
  "event": "ready",
  "firmware_version": "1.0.0",
  "project_id": "<configured-project-id>",
  "device_id": "<non-secret-device-id>",
  "uptime_ms": 1234
}
```

The verifier should then send a fresh challenge:

```json
{
  "protocol": "luma.runtime",
  "protocol_version": 1,
  "command": "identify",
  "challenge": "<random-single-use-token>"
}
```

Firmware responds with the same challenge and its current identity:

```json
{
  "protocol": "luma.runtime",
  "protocol_version": 1,
  "event": "identity",
  "challenge": "<same-token>",
  "firmware_version": "1.0.0",
  "project_id": "<configured-project-id>",
  "device_id": "<non-secret-device-id>",
  "uptime_ms": 2345
}
```

These records are illustrative examples, not existing firmware output. The challenge prevents stale buffered identity responses from being mistaken for a fresh response, but does **not** authenticate the board against an active attacker. A stronger authenticity claim requires a device-bound key and a reviewed challenge-response signature or MAC design.

## Parser behavior implemented

The hardware-independent `internal/domain/handshake` package validates a single identity JSON record against a caller-supplied fresh challenge, checks protocol/version/event, rejects malformed or multi-record input, caps record size at 4096 bytes, bounds required identity fields, and rejects absent or negative uptime. A separate readiness validator confirms a valid `ready` record without treating it as an identity response. Tests use simulated JSON only.

The parser does not generate cryptographic challenges, read serial ports, implement a response deadline, authenticate device identity, or prove that a sample originated from a physical device. A caller must not treat parsed sample data as hardware evidence.

## Verification levels

| Level | Required evidence | Allowed report |
| --- | --- | --- |
| `port_detected` | Arduino CLI lists the requested port | Port observed; no identity claim |
| `board_candidate_matched` | Port record includes the requested FQBN | Candidate match only |
| `upload_command_succeeded` | Actual upload process exited zero | Upload command succeeded; runtime not proven |
| `runtime_handshake_received` | Fresh challenge echoed in a valid protocol-v1 identity response received over the selected live transport | Firmware endpoint responded with the reported identity |
| `device_authenticated` | Runtime response validates against provisioned device-bound cryptographic credentials | Authenticated device identity, subject to key lifecycle/security review |

Do not collapse these levels into one boolean. In particular, `hardware_verified` must remain false for the first three levels. A future implementation may define `runtime_responded=true` after a valid fresh challenge response; it must not set `device_authenticated=true` without cryptographic verification.

## Required verifier checks

1. Require explicit operator consent before opening a serial port; do not infer consent from an upload confirmation.
2. Open only the exact user-selected port and close it on every exit path.
3. Do not toggle DTR/RTS or deliberately reset the board by default; document platform-specific serial-open behavior before implementation.
4. Generate a cryptographically random, single-use challenge; compare the returned challenge in constant time where applicable.
5. Validate protocol name/version, event type, required fields, field lengths, expected project/device identifiers when supplied, and response deadline.
6. Reject stale, malformed, oversized, duplicate, or mismatched responses. Never log secrets or the full challenge if it is used as a security credential.
7. Report the exact evidence level and errors. A timeout means `runtime_responded=false`, not that the firmware is definitely absent.
8. Verification must be read-only: no lamp toggles, relay operations, firmware writes, schedule changes, or network configuration changes.

## Failure cases

- Port missing: stop; report `serial_port_detected=false`.
- Port present but no matching board candidate: report candidate mismatch; allow explicit operator review, do not guess a port.
- Upload failed or exit code unavailable: report upload status separately; do not infer runtime status.
- No response before timeout: report `runtime_responded=false` and a timeout reason.
- Challenge mismatch or invalid response: reject the response and do not authenticate the device.
- Protocol version unsupported: report incompatible protocol, not success.
- Identity differs from expected configuration: report mismatch and do not mark the expected device verified.

## Implementation sequence

1. Add a protocol parser and unit tests independent of hardware. **Implemented.**
2. Add a serial transport behind a separate interface with explicit enablement and consent.
3. Test using a fake serial transport and simulated firmware records.
4. Add integration tests with a loopback simulator; test timeout, malformed JSON, stale challenge, mismatch, cancellation, and cleanup.
5. Only then run a supervised test on a spare ESP32 with lamp loads disconnected or otherwise made safe.
6. Keep cryptographic device authentication as a separate milestone; do not equate an echoed challenge with trusted hardware identity.

## Current implementation boundary

The current `arduino_verify_device` tool only re-runs `arduino-cli board list --format json`. It does not implement this protocol, open serial ports, send challenges, receive firmware messages, or authenticate a device. Its `hardware_verified` result must remain false.

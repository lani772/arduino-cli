# LUMA ESP32 Runtime Handshake — Protocol Design

Status: **parser, fake-source reader, and consent-gated transport interface implemented; live serial handshake not implemented**

## Purpose

Define the minimum evidence needed before a future verifier can report that LUMA firmware is running. Arduino CLI upload success, a serial port, and an Arduino CLI board candidate are not runtime proof. This document does not authorize the current MCP service to open serial ports or reset boards.

## Proposed transport

- Transport: newline-delimited JSON over USB serial, using a future explicitly enabled verifier adapter.
- Baud rate: 115200 by default; configurable only when both firmware and verifier agree.
- Firmware emits one JSON object per line and terminates each record with `\n`.
- The verifier must use a bounded timeout and bounded line size; it must ignore unrelated boot logs and malformed lines only if that behavior is explicitly designed and tested.
- No relay/lamp outputs may change as a side effect of a verification request.

## Protocol version 1

Firmware may emit a readiness record:

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

The future verifier sends a fresh challenge:

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

## Parser and injected-source reader

The hardware-independent `internal/domain/handshake` package validates a single identity JSON record against a caller-supplied challenge, checks protocol/version/event, rejects malformed or multi-record input, caps record size at 4096 bytes, bounds required identity fields, and rejects absent or negative uptime. The `ReadIdentity` helper consumes a context-aware injected `LineSource`, skips a valid readiness record, applies a bounded timeout, and closes the source on every read attempt that begins after argument validation. Fake-source tests cover successful flow, challenge mismatch, caller cancellation, deadline expiration, invalid arguments, source failure, and cleanup.

A `LineSource` implementation must make `Close` safe to call when a read is blocked and must honor context cancellation; otherwise its underlying goroutine may not stop promptly. The package does not generate cryptographic challenges, send the challenge to firmware, read serial ports, or authenticate device identity. Parsed data from a sample or fake source is not evidence that a physical device responded.

## Verification levels

| Level | Required evidence | Allowed report |
| --- | --- | --- |
| `port_detected` | Arduino CLI lists the requested port | Port observed; no identity claim |
| `board_candidate_matched` | Port record includes the requested FQBN | Candidate match only |
| `upload_command_succeeded` | Actual upload process exited zero | Upload command succeeded; runtime not proven |
| `runtime_handshake_received` | Fresh challenge echoed in a valid protocol-v1 identity response received over the selected live transport | Firmware endpoint responded with the reported identity |
| `device_authenticated` | Runtime response validates against provisioned device-bound cryptographic credentials | Authenticated device identity, subject to key lifecycle/security review |

Do not collapse these levels into one boolean. In particular, `hardware_verified` must remain false for the first three levels. An echoed challenge alone does not establish trusted hardware identity.

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
2. Add a context-bounded reader over an injected line-source interface, tested with fake sources. **Implemented.**
3. Test source closure on success, malformed responses, cancellation, timeout, and read errors. **Implemented for success, mismatch, cancellation, timeout, and source error.**
4. Add an actual serial adapter only after timeout, cancellation, malformed input, mismatch, and cleanup tests pass and operator-consent behavior is specified.
5. Test using a loopback simulator before supervised hardware trials.
6. Only then run a supervised test on a spare ESP32 with lamp loads disconnected or otherwise made safe.
7. Keep cryptographic device authentication as a separate milestone.

## Current implementation boundary

The current `arduino_verify_device` tool only re-runs `arduino-cli board list --format json`. It does not implement this protocol, open serial ports, send challenges, receive firmware messages, or authenticate a device. Its `hardware_verified` result must remain false.


## Final protocol contract (Task 28)

This section freezes the behavior expected of the protocol-v1 parser and reader before a real serial transport is introduced. If implementation and this contract disagree, resolve the disagreement in code and tests before enabling serial access.

### Record and state rules

- Encoding is UTF-8 JSON, one object per LF-terminated line. A CR immediately before LF is permitted. Serial framing must enforce the 4096-byte maximum before buffering an unbounded line.
- A valid `ready` record is informational only. The reader may skip it and continue waiting for an identity response.
- The first non-readiness record is treated as the identity candidate. If it is malformed, unknown, oversized, for another protocol/version, or has the wrong event/challenge, fail closed; do not silently scan past invalid records looking for a success.
- Exactly one JSON value is permitted per record. Unknown fields are rejected in protocol v1 so unexpected semantics do not pass unnoticed.
- Required identity strings must be non-blank and at most 128 bytes. Uptime must be present and non-negative. Project/device values are claims from the responding endpoint until separately compared with expected values or authenticated.
- A challenge is caller-supplied to the current parser/reader. It must be non-blank, have no surrounding whitespace, and be at most 256 bytes. The live verifier must generate at least 128 bits of cryptographically secure randomness, use each challenge once, and never use a fixed production token. The reader does not itself generate or transmit a challenge.
- A valid echoed challenge provides freshness against stale buffered responses, not cryptographic authenticity. Treat serial input as untrusted, including when the port is physically local.

### Timeout and cancellation contract

- The caller supplies a positive timeout. One overall deadline covers readiness records and the identity response; repeated readiness records must not restart or extend the deadline.
- Caller cancellation takes precedence as a cancellation outcome. Expiration of the reader's own deadline returns a deadline-exceeded error. Neither outcome implies that firmware is absent or defective.
- Once arguments have passed validation and reading begins, the source must be closed on success, parse failure, read failure, cancellation, and timeout. Invalid arguments do not transfer source ownership and must not close the source.
- A `LineSource` implementation must unblock a pending read when its context is cancelled or when `Close` is called. A transport that cannot guarantee this contract must not be wired into this reader, because it can leak a goroutine or retain a port after timeout.
- The current reader closes the source but does not propagate a `Close` error. The future serial adapter should report close failures through diagnostics without replacing the primary handshake failure. Whether close errors become a typed result is an implementation decision for the transport task.
- For live serial, start with a configurable **3-second overall response deadline** after the challenge is sent; validate this against real boot timing on supported boards before treating it as a release default. Port-open time and firmware boot delay must not silently extend the response deadline.

### Threat model

| Threat or failure | Required mitigation | Residual limitation |
| --- | --- | --- |
| Stale bytes or a prior identity record | Fresh single-use random challenge; exact challenge match | Echo-only challenge is not authentication |
| Malicious/noisy serial output | Size limits, strict JSON/schema checks, bounded deadline, fail closed on invalid identity candidate | Parser cannot prove who produced bytes |
| Wrong device/firmware on selected port | Compare returned project/device identifiers with expected configuration and report mismatch | Self-reported identifiers can be spoofed |
| Unsupported or altered protocol fields | Exact protocol/version/event checks; reject unknown fields | Future protocol extensions require an explicit version/schema update |
| Read never completes | Context deadline, cancellation-aware source, close-on-exit tests | A broken adapter can still violate cancellation contract |
| Accidental physical side effects | Read-only protocol; no lamp commands; explicit consent before port open | Some serial drivers may toggle DTR/RTS on open; platform behavior needs testing |
| Replay or active serial impersonation | Random challenge and constant-time equality check | Use device-bound MAC/signature and reviewed key provisioning before claiming authenticated identity |
| Sensitive data leaks through diagnostics | Do not log challenge values, credentials, or raw untrusted records by default | Logging implementation requires its own review |

Out of scope for protocol v1: firmware flashing, GPIO/relay control, secret provisioning, key rotation, and proof of secure boot. The handshake must not authorize any of those actions.

### Test matrix and acceptance criteria

The existing fake-source tests are the hardware-independent baseline. Keep them deterministic and run them before enabling a live adapter.

| Case | Expected result | Existing coverage / required follow-up |
| --- | --- | --- |
| Valid identity and matching challenge | Return parsed identity and close source once | Covered by `TestReadIdentitySkipsReadyThenAcceptsIdentityAndCloses` |
| One or repeated valid readiness records before identity | Continue within the same deadline | Covered by readiness reader tests |
| Wrong challenge | Reject; no identity returned; close source | Covered by `TestReadIdentityRejectsChallengeMismatchAndCloses` |
| Malformed, oversized, unknown-field, multi-value, wrong-version or wrong-event identity | Reject; never report success | Parser and malformed-reader coverage exists; add reader-level cases where needed |
| No data before deadline | Return deadline exceeded and close source | Covered by `TestReadIdentityHonorsTimeoutAndClosesSource` |
| Caller cancellation | Return cancellation and close source | Covered by `TestReadIdentityHonorsCallerCancellationAndCloses` |
| Read error / EOF | Return wrapped source error and close source | Covered by `TestReadIdentityReturnsSourceFailureAndCloses` |
| Invalid nil source/context, non-positive timeout, or invalid challenge | Reject before taking ownership; do not close source | Covered in argument validation tests |
| Endless stream of readiness records | Overall deadline still expires; no deadline reset | **Add/retain a deterministic test before live serial integration** |
| Source ignores cancellation or cannot unblock on close | Adapter contract test must detect/avoid using such an adapter | Required for each concrete transport |
| Close reports an error | Preserve primary result and surface close failure diagnostically | Required when transport adapter defines close-error handling |
| Challenge generation | CSPRNG output, correct minimum entropy, uniqueness across calls | Required in live verifier; parser tests do not cover generation |
| Port selection and consent | Open only selected port after explicit consent; no implicit reset | Required in live adapter integration tests |
| Physical device mismatch | Report mismatch and leave `hardware_verified=false` | Required in simulator and supervised hardware tests |

### Gate before live serial implementation

Do not add a real serial adapter until the following are true:

1. The test suite is executed successfully, including race detection; source-level inspection alone is not sufficient.
2. A deterministic repeated-readiness/deadline test proves that the single overall deadline cannot be extended by incoming records.
3. The adapter contract proves cancellation unblocks reads and port cleanup happens on every exit path.
4. Port opening requires explicit operator consent and an exact selected port; no automatic port guessing or reset behavior.
5. Results distinguish a received runtime response from authenticated device identity. `hardware_verified` remains false unless the higher-level verification contract explicitly defines and validates stronger evidence.
6. The firmware-side protocol is implemented and tested against the same schema. The JSON examples in this document are not proof that firmware currently emits these records.

**Verification note:** this specification update does not mean tests have run. At the time of this commit, test execution, race detection, and physical serial testing remain unverified.


## Transport interface and lifecycle (Task 29)

The hardware-independent handshake.Transport interface accepts a context and the exact selected port, returning a LineSource. OpenSource is the shared guard: it rejects nil context/transport, missing explicit operator consent, blank or whitespace-padded port identifiers, and an already-cancelled context before calling the transport. It never enumerates or guesses ports. Transport-open errors are wrapped for classification, and a nil source is rejected.

Once returned, the caller owns the source. Passing it to ReadIdentity transfers the source into the reader's close-on-exit lifecycle. The transport implementation must honor context cancellation while opening, and the returned source must unblock reads on cancellation or close. No concrete serial adapter, port enumeration, DTR/RTS control, or hardware access is added in this task.

Fake-transport tests cover consent/port validation without opening, exact-port forwarding, open failure, nil source, pre-cancelled context, and source cleanup when the reader returns a read error. These tests have been added but have NOT been executed in this environment.


## Challenge generation (Task 30)

The handshake package now exposes GenerateChallenge, which reads 16 bytes from Go's operating-system cryptographic random source and encodes them as 32 lowercase hexadecimal characters (128 bits of entropy). Entropy read failures are returned as errors; the implementation does not fall back to timestamps, counters, or fixed values. GenerateChallengeFrom exists only to test deterministic encoding and entropy-source failures; production callers should use GenerateChallenge. Each verification attempt must call the generator separately and must not reuse or log a challenge.

The reader's timeout remains a single context deadline spanning readiness records and identity parsing. Tests for challenge generation have been added, including uniqueness checks and short-entropy failure. Tests have not been executed here. Challenge generation does not transmit the challenge: sending it and integrating the live serial lifecycle remain later tasks.

# LUMA Firmware MCP — Architecture

The service is isolated from the Arduino CLI root module. Arduino CLI is a toolchain adapter; LUMA owns firmware-project orchestration, AI engineering, diagnostics, repair policy, and future MCP transport.

## Dependency direction

- Domain: pure business contracts and validation.
- Application: use-case ports.
- Infrastructure: Arduino CLI, AI providers, filesystem, serial/device adapters.
- Interfaces: HTTP/MCP transport.
- Policy: explicit execution, installation, and flashing controls.

AI output is untrusted. It must never execute arbitrary shell commands. Tool installation and device flashing remain disabled by default.


## Build and compiler diagnostics (Task 39)

The `arduino_compile` MCP tool returns the build result and parses recognizable Arduino/GCC diagnostic lines into structured severity, file, line, column, code, and message fields. The response includes counts for errors, warnings, and informational notes. Unrecognized compiler output remains available in raw build output, so parsing is a convenience layer rather than a lossy replacement for compiler logs.

Compilation does not upload firmware. A failed compiler exit is reported alongside the build result and diagnostics; it is not converted into success, and no automatic AI repair or arbitrary shell command is executed by this change. Firmware-spec validation remains a separate explicit tool; this change does not yet orchestrate generation, validation, and compilation as one atomic workflow.

Regression coverage has been added for a failed compile containing both an error and a warning. Tests have not been executed in this environment.

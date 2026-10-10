# LUMA Firmware MCP — Architecture

The service is isolated from the Arduino CLI root module. Arduino CLI is a toolchain adapter; LUMA owns firmware-project orchestration, AI engineering, diagnostics, repair policy, and future MCP transport.

## Dependency direction

- Domain: pure business contracts and validation.
- Application: use-case ports.
- Infrastructure: Arduino CLI, AI providers, filesystem, serial/device adapters.
- Interfaces: HTTP/MCP transport.
- Policy: explicit execution, installation, and flashing controls.

AI output is untrusted. It must never execute arbitrary shell commands. Tool installation and device flashing remain disabled by default.

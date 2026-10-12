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


## Repair proposal and approval boundary (Task 40)

Compiler-guided repair is proposal-only. The repair loop validates a candidate patch against an in-memory snapshot and returns it with `ErrPatchApprovalRequired`; it does not write the candidate into the active project or silently continue rebuilding modified source. Callers must present the proposed diff and obtain an explicit user decision.

The application-layer `PatchApprovalService.ApplyApproved` is the controlled persistence entry point. The caller must pass explicit approval, the service re-snapshots the project, and each update/delete is guarded by the expected SHA-256 of the current file contents. Unsafe or non-normalized paths, stale hashes, invalid operations, cancelled contexts, and write failures are rejected. The project-files adapter must implement `Replace` atomically; filesystem implementations should stage changes and atomically swap or rollback on failure.

This is an application API, not yet an MCP approval UI/tool. A future transport integration must keep the proposed patch separate and collect explicit approval from the human before invoking it. Firmware upload remains a separate operation and is not triggered by patch approval.


## Repair-loop reliability boundary

The repair loop performs one baseline compilation and may return one validated repair proposal when structured compiler errors are available. It does not retry builds, restore files, or persist proposed changes. Malformed patches, stale expected hashes, AI failures, and cancelled contexts return errors without writing project source. The caller must obtain explicit approval and invoke the patch-approval service; a later compile is a separate operation after that approved persistence succeeds.

The filesystem adapter's `Replace` operation must provide atomic replacement. If persistence fails, the approval service returns the error; guarantees about the on-disk state depend on that adapter contract.


## MCP patch review workflow (Task 42)

The MCP package provides `RegisterPatchReviewTools`, which registers `luma_repair_patch_propose` and `luma_repair_patch_decide`. Proposals are validated against the current project snapshot and held in an in-memory review registry. Proposal submission and rejection do not write files or compile. Explicit approval rechecks the expected content hashes through `PatchApprovalService`, persists the approved patch, and only then invokes a separate build. Failed persistence returns an error and leaves the proposal available for retry.

The review registry is process-local and is lost on restart. The current command-line server does not yet register these tools because it does not construct a persistent project store or a production `ProjectFiles` adapter. Production wiring must first provide an explicitly configured, workspace-confined filesystem adapter with atomic replacement semantics and durable project storage; arbitrary client-supplied filesystem paths must not be trusted.

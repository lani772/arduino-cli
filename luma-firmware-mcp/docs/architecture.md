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

The filesystem adapter's `Replace` operation stages the complete source tree and swaps directories, attempting rollback if installation fails. If persistence fails, the approval service returns the error; guarantees about on-disk state depend on the adapter and host filesystem.


## MCP patch review workflow (Task 42)

The MCP package provides `RegisterPatchReviewTools`, which registers `luma_repair_patch_propose` and `luma_repair_patch_decide`. Proposals are validated against the current project snapshot and held in an in-memory review registry. Proposal submission and rejection do not write files or compile. Explicit approval rechecks the expected content hashes through `PatchApprovalService`, persists the approved patch, and only then invokes a separate build. Failed persistence returns an error and leaves the proposal available for retry.

The review registry is process-local and is lost on restart. Task 44 wires these tools into the stdio server only when `LUMA_WORKSPACE_ROOT` is explicitly configured. A durable JSON-backed project store and the workspace-confined filesystem adapter are constructed at startup; arbitrary client-supplied filesystem paths are not used as the workspace root. Pending proposal persistence remains Task 45.


## Workspace-confined filesystem adapter (Task 43)

The `internal/infrastructure/filesystem` package implements the application `ProjectFiles` port. Construct it with `filesystem.New(workspaceRoot)`; the configured root must already exist and must be a real directory. A project's `Source` must identify a child directory of that root (absolute paths inside the root and paths relative to it are supported). Paths that escape the root are rejected.

Snapshots return slash-separated relative file names and contents. Symlinks, non-regular files, and symlinked path components are rejected instead of followed. Snapshot and replacement check context cancellation. Replacement stages the complete resulting snapshot in a sibling directory, moves the current project to a backup, then renames the staged directory into place. If installing the stage fails, it attempts to restore the original directory. Replacement preserves the complete map supplied by the application; files absent from that map are removed as part of the directory swap.

Configure the root only from trusted server configuration (for example, a deployment environment variable); never derive it from an MCP tool argument. The adapter does not make a hostile, concurrently writable workspace safe against local filesystem races: the service account should exclusively control the workspace tree. Directory rename atomicity and crash durability depend on the host filesystem. The adapter is not yet wired into the default server entry point; runtime registration and durable project storage remain Task 44. Tests were added for snapshots, replacement, traversal, symlinks, and cancellation, but have not been executed in this environment.


## Persistent project metadata and runtime wiring (Task 44)

The stdio server enables patch review only when `LUMA_WORKSPACE_ROOT` points to an explicitly configured, existing directory. The JSON-backed `internal/infrastructure/filedb.ProjectStore` persists project metadata using a temporary file, sync, and rename. By default, metadata is stored at `<workspace>/.luma/projects.json`; `LUMA_PROJECTS_FILE` may override the path only within the workspace. Metadata paths and project source paths reject symlinks, and project source directories must be children of the configured root.

The server registers `luma_project_register`, `luma_repair_patch_propose`, and `luma_repair_patch_decide` when the workspace configuration is valid. If no workspace root is configured, these tools stay disabled and the other MCP tools remain available. Registration records only existing project directories; it does not modify or compile them. The build runs only after explicit patch approval and successful patch persistence. This configuration does not yet make pending repair proposals durable, and local filesystem races remain out of scope; keep the workspace writable only by the service account. Tests have been added but have not been executed in this environment.

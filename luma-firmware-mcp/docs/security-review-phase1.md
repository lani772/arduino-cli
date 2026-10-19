# LUMA Firmware MCP — Phase 1 Security Review

Status: **initial source review; not a security certification**

Review scope: process execution policy, library installation, upload controls, repair patch application, filesystem boundaries, and MCP transport exposure. This review is based on source inspection only. Go tests, race tests, `go vet`, dynamic testing, and hardware tests have not been run as part of this review.

## Fixed in this review

### Library installation did not enforce its dedicated policy flag

The CLI adapter checked whether the `arduino-cli` executable was allowed, but `InstallLibrary` did not check `ExecutionPolicy.AllowToolInstallation`. As a result, callers could install libraries while the default policy declared installation disabled.

The adapter now rejects library installation unless `AllowToolInstallation` is explicitly enabled. Tests were added for default denial (and no command execution) and explicit opt-in.

## Findings requiring follow-up

### High — AI repair is automatically persisted by the repair loop

`RepairLoopService.Run` asks the AI engineer for a structured patch, validates the patch and expected content hashes, then calls `ProjectFiles.Replace` without a separate human approval step. Path and hash validation reduce some risks, but they do not establish that a proposed firmware change is safe. A patch can change device behavior while remaining structurally valid.

**Required before enabling this workflow for real projects:** introduce a review/approval boundary before persisting generated changes; keep proposed changes separate from the active project; test rejection, cancellation, and rollback paths.

### High — project source paths are not visibly confined to a workspace boundary

The project domain validates identity fields but does not validate or bind `Project.Source` to a trusted workspace root. The Arduino CLI adapter passes source paths to the external CLI and does not set a controlled working directory. The filesystem adapter was not fully validated during this review, so this is a source-level concern rather than a confirmed exploit.

**Required before exposing untrusted project input:** canonicalize paths at the filesystem boundary, reject paths outside the selected workspace, account for symlinks, and test traversal and path replacement cases.

### Medium — device flashing relies on configuration plus tool-level confirmation

Device flashing is disabled by default, and the MCP upload tool requires explicit confirmation and a port. However, the adapter's policy opt-in is a broad boolean; it does not itself constrain permitted ports or prove the identity of the attached board. An upload exit code is not proof that the intended firmware is running.

**Required before enabling flashing:** preserve default denial, bind confirmation to the exact project/FQBN/port operation, validate the selected port against fresh discovery, and keep runtime identity verification separate from upload success.

### Medium — process policy needs to govern each side-effecting capability

The default policy allows process execution only through the configured `arduino-cli` executable, while tool installation and device flashing are separate capabilities. The library-installation gap identified above is fixed in this branch. Any future process-backed operation must have its own explicit policy check when it performs a side effect.

### Scope note — transport boundary

The inspected HTTP server currently exposes only `/healthz`; this review did not find an HTTP MCP endpoint in that file. The stdio MCP server is intended for a local parent process and has no end-user authentication layer. Do not expose it as a network service without adding and testing an appropriate authentication/authorization boundary, request limits, and operational controls.

## Verification status

- Source review: completed for the files listed in this review's scope, with the filesystem boundary noted as incomplete.
- New tests: added for the library-installation policy gate; **not executed here**.
- Formatting, `go test ./...`, `go test -race ./...`, and `go vet ./...`: **not verified**.
- Physical ESP32 or serial-device testing: not performed.
- Release decision: **not ready for release** until validation and high-priority follow-ups are complete.

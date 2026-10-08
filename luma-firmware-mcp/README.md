# LUMA Firmware MCP

AI-assisted Arduino/ESP32 firmware engineering service isolated from the Arduino CLI root module.

## Run

HTTP health endpoint (default):

```bash
go run ./cmd/server
# GET http://localhost:8080/healthz
```

MCP over stdio (JSON-RPC, newline-delimited messages):

```bash
go run ./cmd/server -transport stdio
```

In stdio mode, stdout is reserved for MCP protocol messages and logs go to stderr. The transport currently supports initialization, ping, and empty tools/resources/prompts lists. MCP tools are added in the next phase.

## Validate

```bash
gofmt -w .
go test ./...
go vet ./...
```

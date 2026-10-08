package main

import (
 "log"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/interfaces/httpserver"
)

const version = "0.1.0"

func main() {
 logger := log.Default()
 server := httpserver.New(":8080", version, logger)
 logger.Printf("version=%s status=ready", version)
 if err := server.Start(); err != nil { logger.Fatal(err) }
}

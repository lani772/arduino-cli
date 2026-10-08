package main

import (
 "context"
 "flag"
 "log"
 "os"

 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/interfaces/httpserver"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/interfaces/mcp"
)

const version = "0.1.0"

func main() {
 transport := flag.String("transport", "http", "transport: http or stdio")
 address := flag.String("address", ":8080", "HTTP listen address")
 flag.Parse()
 logger := log.New(os.Stderr, "", log.LstdFlags)
 switch *transport {
 case "stdio":
  logger.Printf("version=%s transport=stdio status=ready", version)
  if err := mcp.NewStdio(os.Stdin, os.Stdout).Start(context.Background()); err != nil {
   logger.Fatal(err)
  }
 case "http":
  server := httpserver.New(*address, version, logger)
  logger.Printf("version=%s transport=http status=ready address=%s", version, *address)
  if err := server.Start(); err != nil {
   logger.Fatal(err)
  }
 default:
  logger.Fatalf("unsupported transport %q; use http or stdio", *transport)
 }
}

package main

import (
 "context"
 "flag"
 "log"
 "os"

 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/infrastructure/arduino"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/interfaces/httpserver"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/interfaces/mcp"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/policy"
)
const version = "0.1.0"
func main() {
 transport:=flag.String("transport","http","transport: http or stdio");address:=flag.String("address",":8080","HTTP listen address");flag.Parse()
 logger:=log.New(os.Stderr,"",log.LstdFlags)
 switch *transport {
 case "stdio":
  server:=mcp.NewStdio(os.Stdin,os.Stdout)
  cli:=arduino.NewCLI("arduino-cli",nil,policy.DefaultExecutionPolicy())
  if err:=mcp.RegisterArduinoTools(server.Registry,cli);err!=nil{logger.Fatal(err)}
  if err:=mcp.RegisterFirmwareValidationTool(server.Registry);err!=nil{logger.Fatal(err)}
  if err:=mcp.RegisterESP32Tools(server.Registry);err!=nil{logger.Fatal(err)}
  if err:=mcp.RegisterDefaultResources(server.Resources,cli);err!=nil{logger.Fatal(err)}
  if err:=mcp.RegisterDefaultPrompts(server.Prompts);err!=nil{logger.Fatal(err)}
  logger.Printf("version=%s transport=stdio tools=%d resources=%d prompts=%d status=ready",version,len(server.Registry.List()),len(server.Resources.List()),len(server.Prompts.List()))
  if err:=server.Start(context.Background());err!=nil{logger.Fatal(err)}
 case "http":
  server:=httpserver.New(*address,version,logger);logger.Printf("version=%s transport=http status=ready address=%s",version,*address)
  if err:=server.Start();err!=nil{logger.Fatal(err)}
 default:logger.Fatalf("unsupported transport %q; use http or stdio",*transport)
 }
}

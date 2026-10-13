package main

import (
	"context"
	"flag"
	"log"
	"path/filepath"
	"os"
	"strings"

	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/application"
	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/infrastructure/arduino"
	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/infrastructure/filedb"
	workspacefs "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/infrastructure/filesystem"
 	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/infrastructure/reviewstore"
	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/interfaces/httpserver"
	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/interfaces/mcp"
	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/policy"
)

const version = "0.1.0"

func main() {
	transport := flag.String("transport", "http", "transport: http or stdio")
	address := flag.String("address", ":8080", "HTTP listen address")
	flag.Parse()
	logger := log.New(os.Stderr, "", log.LstdFlags)

	switch *transport {
	case "stdio":
		server := mcp.NewStdio(os.Stdin, os.Stdout)
		cli := arduino.NewCLI("arduino-cli", nil, policy.DefaultExecutionPolicy())
		if err := mcp.RegisterArduinoTools(server.Registry, cli); err != nil { logger.Fatal(err) }
		if err := mcp.RegisterUploadTool(server.Registry, cli); err != nil { logger.Fatal(err) }
		if err := mcp.RegisterFirmwareValidationTool(server.Registry); err != nil { logger.Fatal(err) }
		if err := mcp.RegisterESP32Tools(server.Registry); err != nil { logger.Fatal(err) }
		if err := mcp.RegisterDeviceTools(server.Registry, cli); err != nil { logger.Fatal(err) }
		if err := mcp.RegisterDefaultResources(server.Resources, cli); err != nil { logger.Fatal(err) }
		if err := mcp.RegisterDefaultPrompts(server.Prompts); err != nil { logger.Fatal(err) }

		workspaceRoot := strings.TrimSpace(os.Getenv("LUMA_WORKSPACE_ROOT"))
		if workspaceRoot == "" {
			logger.Printf("patch review tools disabled: set LUMA_WORKSPACE_ROOT to an explicitly trusted workspace directory")
		} else {
			files, err := workspacefs.New(workspaceRoot)
			if err != nil { logger.Fatalf("configure workspace filesystem: %v", err) }
			projects, err := filedb.NewProjectStore(files.Root, os.Getenv("LUMA_PROJECTS_FILE"))
			if err != nil { logger.Fatalf("configure persistent project store: %v", err) }
			if err := mcp.RegisterProjectTools(server.Registry, projects); err != nil { logger.Fatal(err) }
			journalPath := strings.TrimSpace(os.Getenv("LUMA_REVIEW_JOURNAL"))
            if journalPath == "" { journalPath = filepath.Join(files.Root, ".luma", "patch-review.json") }
            journalAbs, err := filepath.Abs(journalPath)
            if err != nil { logger.Fatalf("resolve patch review journal path: %v", err) }
            journalRel, err := filepath.Rel(files.Root, journalAbs)
            if err != nil || journalRel == "." || journalRel == ".." || strings.HasPrefix(journalRel, ".."+string(filepath.Separator)) { logger.Fatal("LUMA_REVIEW_JOURNAL must be inside LUMA_WORKSPACE_ROOT") }
            journal, err := reviewstore.New(journalAbs)
            if err != nil { logger.Fatalf("configure patch review journal: %v", err) }
            if err := mcp.RegisterPatchReviewToolsWithJournal(server.Registry, projects, application.PatchApprovalService{Files:files}, application.BuildService{CLI:cli}, journal); err != nil { logger.Fatal(err) }
			logger.Printf("workspace-confined project and patch-review tools enabled root=%s", files.Root)
		}
		logger.Printf("version=%s transport=stdio tools=%d resources=%d prompts=%d status=ready", version, len(server.Registry.List()), len(server.Resources.List()), len(server.Prompts.List()))
		if err := server.Start(context.Background()); err != nil { logger.Fatal(err) }
	case "http":
		server := httpserver.New(*address, version, logger)
		logger.Printf("version=%s transport=http status=ready address=%s", version, *address)
		if err := server.Start(); err != nil { logger.Fatal(err) }
	default:
		logger.Fatalf("unsupported transport %q; use http or stdio", *transport)
	}
}

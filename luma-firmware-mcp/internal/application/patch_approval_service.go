package application

import (
 "context"
 "fmt"

 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
)

// PatchApprovalService is the only application-layer entry point for persisting proposed repair changes.
type PatchApprovalService struct { Files ProjectFiles }

// ApplyApproved persists a patch only when the caller explicitly confirms approval.
// It re-snapshots the project and repair.Apply checks expected content hashes before atomic replacement.
func (s PatchApprovalService) ApplyApproved(ctx context.Context, p project.Project, patch repair.Patch, approved bool) error {
 if err:=p.Validate(); err!=nil { return err }
 if s.Files==nil { return fmt.Errorf("project files adapter is required") }
 if !approved { return fmt.Errorf("patch application requires explicit approval") }
 if err:=ctx.Err(); err!=nil { return err }
 if err:=patch.Validate(); err!=nil { return fmt.Errorf("invalid patch: %w",err) }
 if patch.Status!=repair.StatusProposed { return fmt.Errorf("only proposed patches can be approved") }
 snapshot,err:=s.Files.Snapshot(ctx,p); if err!=nil { return fmt.Errorf("snapshot project: %w",err) }
 if len(snapshot)==0 { return fmt.Errorf("project source snapshot is empty") }
 next,err:=repair.Apply(snapshot,patch); if err!=nil { return fmt.Errorf("patch rejected: %w",err) }
 if err:=ctx.Err(); err!=nil { return err }
 if err:=s.Files.Replace(ctx,p,next); err!=nil { return fmt.Errorf("persist approved patch: %w",err) }
 return nil
}

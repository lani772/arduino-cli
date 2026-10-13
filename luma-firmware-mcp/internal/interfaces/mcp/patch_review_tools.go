package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/application"
	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
	"github.com/lani772/arduino-cli/luma-firmware-mcp/internal/infrastructure/reviewstore"
)

type pendingPatch struct {
	Project project.Project
	FQBN string
	Patch repair.Patch
}

type ReviewJournal interface {
	Snapshot() reviewstore.State
	SaveProposal(string, reviewstore.Proposal) error
	BeginDecision(string, string) (reviewstore.Proposal, error)
	CompleteDecision(string, string, string) error
	RestoreProposal(string, reviewstore.Proposal, string) error
}

type patchReviewRegistry struct {
	mu sync.Mutex
	pending map[string]pendingPatch
	decided map[string]string
	journal ReviewJournal
}

func newPatchReviewRegistry(journal ReviewJournal) *patchReviewRegistry {
	state := &patchReviewRegistry{pending: map[string]pendingPatch{}, decided: map[string]string{}, journal: journal}
	if journal != nil {
		saved := journal.Snapshot()
		for id, p := range saved.Proposals {
			state.pending[id] = pendingPatch{Project:p.Project, FQBN:p.FQBN, Patch:p.Patch}
		}
		for id, decision := range saved.Decisions {
			state.decided[id] = decision
		}
	}
	return state
}

// RegisterPatchReviewTools exposes a review-first workflow without durable storage.
func RegisterPatchReviewTools(r *Registry, projects application.ProjectStore, approvals application.PatchApprovalService, builder application.BuildService) error {
	return RegisterPatchReviewToolsWithJournal(r, projects, approvals, builder, nil)
}

// RegisterPatchReviewToolsWithJournal enables durable proposal, decision, and audit state.
func RegisterPatchReviewToolsWithJournal(r *Registry, projects application.ProjectStore, approvals application.PatchApprovalService, builder application.BuildService, journal ReviewJournal) error {
	if r == nil || projects == nil || approvals.Files == nil || builder.CLI == nil {
		return fmt.Errorf("registry, project store, patch approval files, and build service are required")
	}
	state := newPatchReviewRegistry(journal)
	patchSchema := map[string]any{
		"type":"object",
		"properties":map[string]any{
			"id":map[string]any{"type":"string"},
			"description":map[string]any{"type":"string"},
			"diff":map[string]any{"type":"string"},
			"status":map[string]any{"type":"string","enum":[]string{"proposed"}},
			"files":map[string]any{"type":"array","items":map[string]any{"type":"object","properties":map[string]any{
				"path":map[string]any{"type":"string"},
				"operation":map[string]any{"type":"string","enum":[]string{"create","update","delete"}},
				"content":map[string]any{"type":"string"},
				"expected_sha":map[string]any{"type":"string"},
			},"required":[]string{"path","operation"},"additionalProperties":false}},
		},
		"required":[]string{"id","description","status","files"},
		"additionalProperties":false,
	}
	if err := r.Register(RegisteredTool{
		Definition: Tool{Name:"luma_repair_patch_propose", Description:"Validate and submit a structured firmware repair proposal for review. This tool does not write project files or compile firmware. The returned patch details should be reviewed before a separate decision.", InputSchema:objectSchema(map[string]any{
			"project_id":map[string]any{"type":"string"},
			"fqbn":map[string]any{"type":"string"},
			"patch":patchSchema,
		},"project_id","fqbn","patch")},
		Handler:func(ctx context.Context, raw json.RawMessage)(any,error) {
			var args struct {
				ProjectID string `json:"project_id"`
				FQBN string `json:"fqbn"`
				Patch repair.Patch `json:"patch"`
			}
			if err:=json.Unmarshal(raw,&args);err!=nil{return nil,fmt.Errorf("invalid proposal arguments: %w",err)}
			args.ProjectID=strings.TrimSpace(args.ProjectID)
			args.FQBN=strings.TrimSpace(args.FQBN)
			if args.ProjectID==""||args.FQBN=="" {return nil,fmt.Errorf("project_id and fqbn are required")}
			p,err:=projects.Get(ctx,args.ProjectID);if err!=nil{return nil,fmt.Errorf("load project: %w",err)}
			if err:=p.Validate();err!=nil{return nil,err}
			if err:=args.Patch.Validate();err!=nil{return nil,fmt.Errorf("invalid patch: %w",err)}
			if args.Patch.Status!=repair.StatusProposed{return nil,fmt.Errorf("only proposed patches can be submitted for review")}
			snapshot,err:=approvals.Files.Snapshot(ctx,p);if err!=nil{return nil,fmt.Errorf("snapshot project: %w",err)}
			if len(snapshot)==0{return nil,fmt.Errorf("project source snapshot is empty")}
			if _,err:=repair.Apply(snapshot,args.Patch);err!=nil{return nil,fmt.Errorf("patch rejected: %w",err)}
			state.mu.Lock()
			defer state.mu.Unlock()
			if _,exists:=state.pending[args.Patch.ID];exists{return nil,fmt.Errorf("patch id already pending")}
			if _,exists:=state.decided[args.Patch.ID];exists{return nil,fmt.Errorf("patch id has already been decided")}
			proposal:=pendingPatch{Project:p,FQBN:args.FQBN,Patch:args.Patch}
			if state.journal!=nil {
				if err:=state.journal.SaveProposal(args.Patch.ID,reviewstore.Proposal{Project:p,FQBN:args.FQBN,Patch:args.Patch});err!=nil{return nil,err}
			}
			state.pending[args.Patch.ID]=proposal
			return map[string]any{"status":"awaiting_explicit_decision","patch":args.Patch,"project_id":p.ID,"fqbn":args.FQBN,"files_will_be_written":false,"build_will_run":false},nil
		},
	});err!=nil{return err}

	return r.Register(RegisteredTool{
		Definition:Tool{Name:"luma_repair_patch_decide",Description:"Explicitly approve or reject a pending repair proposal. Rejection never writes files. Approval rechecks current file hashes, persists the patch, and only then starts a separate build.",InputSchema:objectSchema(map[string]any{
			"patch_id":map[string]any{"type":"string"},
			"approved":map[string]any{"type":"boolean"},
		},"patch_id","approved")},
		Handler:func(ctx context.Context,raw json.RawMessage)(any,error){
			var args struct {
				PatchID string `json:"patch_id"`
				Approved *bool `json:"approved"`
			}
			if err:=json.Unmarshal(raw,&args);err!=nil{return nil,fmt.Errorf("invalid decision arguments: %w",err)}
			args.PatchID=strings.TrimSpace(args.PatchID)
			if args.PatchID==""{return nil,fmt.Errorf("patch_id is required")}
			if args.Approved==nil{return nil,fmt.Errorf("approved must be explicitly true or false")}
			state.mu.Lock()
			proposal,ok:=state.pending[args.PatchID]
			if !ok {state.mu.Unlock();return nil,fmt.Errorf("pending patch not found")}
			if !*args.Approved {
				if state.journal!=nil {
					if _,err:=state.journal.BeginDecision(args.PatchID,"rejected");err!=nil{state.mu.Unlock();return nil,err}
					if err:=state.journal.CompleteDecision(args.PatchID,"rejected","");err!=nil{state.mu.Unlock();return nil,err}
				}
				delete(state.pending,args.PatchID)
				state.decided[args.PatchID]="rejected"
				state.mu.Unlock()
				return map[string]any{"patch_id":args.PatchID,"status":"rejected","files_written":false,"build_triggered":false},nil
			}
			if state.journal!=nil {
				if _,err:=state.journal.BeginDecision(args.PatchID,"approved");err!=nil{state.mu.Unlock();return nil,err}
			}
			delete(state.pending,args.PatchID)
			state.decided[args.PatchID]="processing"
			state.mu.Unlock()

			if err:=approvals.ApplyApproved(ctx,proposal.Project,proposal.Patch,true);err!=nil{
				if state.journal!=nil {
					if restoreErr:=state.journal.RestoreProposal(args.PatchID,reviewstore.Proposal{Project:proposal.Project,FQBN:proposal.FQBN,Patch:proposal.Patch},err.Error());restoreErr!=nil{return nil,fmt.Errorf("patch application failed (%v) and journal restoration failed (%w)",err,restoreErr)}
				}
				state.mu.Lock();delete(state.decided,args.PatchID);state.pending[args.PatchID]=proposal;state.mu.Unlock()
				return nil,fmt.Errorf("approved patch was not accepted: %w",err)
			}
			result,buildErr:=builder.Compile(ctx,proposal.Project,proposal.FQBN)
			detail:=""
			if buildErr!=nil{detail=buildErr.Error()}
			if state.journal!=nil {
				if err:=state.journal.CompleteDecision(args.PatchID,"approved",detail);err!=nil{return nil,fmt.Errorf("patch was applied but decision audit persistence failed: %w",err)}
			}
			state.mu.Lock();state.decided[args.PatchID]="approved";state.mu.Unlock()
			response:=map[string]any{"patch_id":args.PatchID,"status":"approved_and_applied","files_written":true,"build_triggered":true,"build":result}
			if buildErr!=nil{response["build_error"]=buildErr.Error()}
			return response,nil
		},
	})
}

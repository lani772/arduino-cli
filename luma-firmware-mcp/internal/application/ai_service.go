package application

import (
 "context"
 "encoding/json"
 "fmt"
 "strings"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/diagnostic"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
)
type AIRequest struct { Operation string; Instruction string; Source string; Diagnostics []diagnostic.Diagnostic }
type AIResponse struct { Content string; Model string; Provider string }
type AIProvider interface { Complete(context.Context, AIRequest) (AIResponse,error) }
type AIService struct { Provider AIProvider }
func (s AIService) GenerateFirmware(ctx context.Context,instruction string)(string,error){
 return s.complete(ctx,AIRequest{Operation:"generate_firmware",Instruction:instruction})
}
func (s AIService) Diagnose(ctx context.Context,ds []diagnostic.Diagnostic,source string)(repair.Patch,error){
 if s.Provider==nil{return repair.Patch{},fmt.Errorf("AI provider is required")}
 instruction := "Return only one JSON repair.Patch object. Include id, description, status=proposed and files[]. Each file change uses path, operation (create/update/delete), content for create/update, and expected_sha (SHA-256 of exact current file content) for update/delete. No markdown. Do not return display-only diffs."
 r,err:=s.Provider.Complete(ctx,AIRequest{Operation:"diagnose",Instruction:instruction,Source:source,Diagnostics:ds})
 if err!=nil{return repair.Patch{},err}
 if strings.TrimSpace(r.Content)==""{return repair.Patch{},fmt.Errorf("AI returned an empty repair proposal")}
 var patch repair.Patch
 if err:=json.Unmarshal([]byte(r.Content),&patch);err!=nil{return repair.Patch{},fmt.Errorf("AI repair proposal must be valid structured JSON: %w",err)}
 if patch.Status==""{patch.Status=repair.StatusProposed}
 if err:=patch.Validate();err!=nil{return repair.Patch{},fmt.Errorf("invalid AI repair patch: %w",err)}
 if len(patch.Files)==0{return repair.Patch{},fmt.Errorf("AI repair proposal has no structured file changes")}
 return patch,nil
}
func (s AIService) Repair(ctx context.Context,source string,p repair.Patch)(string,error){
 r,err:=s.complete(ctx,AIRequest{Operation:"repair",Source:source,Instruction:p.Description})
 return r,err
}
func(s AIService) complete(ctx context.Context,r AIRequest)(string,error){
 if s.Provider==nil{return "",fmt.Errorf("AI provider is required")}
 res,err:=s.Provider.Complete(ctx,r);if err!=nil{return "",err};if strings.TrimSpace(res.Content)==""{return "",fmt.Errorf("AI returned empty content")}
 return res.Content,nil
}

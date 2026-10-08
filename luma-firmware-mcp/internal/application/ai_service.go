package application

import (
 "context"
 "fmt"
 "strings"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/diagnostic"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
)

type AIRequest struct {
 Operation string
 Instruction string
 Source string
 Diagnostics []diagnostic.Diagnostic
}

type AIResponse struct {
 Content string
 Model string
 Provider string
}

type AIProvider interface { Complete(context.Context, AIRequest) (AIResponse,error) }

type AIService struct { Provider AIProvider }

func (s AIService) GenerateFirmware(ctx context.Context,instruction string)(string,error){
 return s.complete(ctx,AIRequest{Operation:"generate_firmware",Instruction:instruction})
}
func (s AIService) Diagnose(ctx context.Context,ds []diagnostic.Diagnostic,source string)(repair.Patch,error){
 r,err:=s.Provider.Complete(ctx,AIRequest{Operation:"diagnose",Source:source,Diagnostics:ds})
 if err!=nil{return repair.Patch{},err}
 if strings.TrimSpace(r.Content)==""{return repair.Patch{},fmt.Errorf("AI returned an empty repair proposal")}
 return repair.Patch{ID:"ai-diagnostic",Description:r.Content,Diff:r.Content,Status:repair.StatusProposed},nil
}
func (s AIService) Repair(ctx context.Context,source string,p repair.Patch)(string,error){
 r,err:=s.Provider.Complete(ctx,AIRequest{Operation:"repair",Source:source,Instruction:p.Description})
 if err!=nil{return "",err};if strings.TrimSpace(r.Content)==""{return "",fmt.Errorf("AI returned empty repaired source")}
 return r.Content,nil
}
func(s AIService) complete(ctx context.Context,r AIRequest)(string,error){
 if s.Provider==nil{return "",fmt.Errorf("AI provider is required")}
 res,err:=s.Provider.Complete(ctx,r);if err!=nil{return "",err};if strings.TrimSpace(res.Content)==""{return "",fmt.Errorf("AI returned empty content")}
 return res.Content,nil
}

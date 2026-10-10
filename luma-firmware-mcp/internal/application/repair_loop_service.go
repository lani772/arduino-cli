package application

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "sort"
 "strings"

 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/build"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/diagnostic"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
)
var ErrRepairAttemptsExhausted = errors.New("repair attempts exhausted")

// ProjectFiles isolates orchestration from filesystem details. Replace must be atomic.
type ProjectFiles interface {
 Snapshot(context.Context, project.Project) (map[string]string, error)
 Replace(context.Context, project.Project, map[string]string) error
}
type RepairLoopResult struct {
 FinalBuild build.Result
 Attempts int
 Diagnostics []diagnostic.Diagnostic
 Patches []repair.Patch
 Recovered bool
}
type RepairLoopService struct {
 Builder BuildService
 Engineer AIEngineer
 Parser DiagnosticParser
 Files ProjectFiles
 MaxAttempts int
}
func (s RepairLoopService) Run(ctx context.Context,p project.Project,fqbn string)(RepairLoopResult,error){
 var out RepairLoopResult
 if err:=p.Validate();err!=nil{return out,err}
 if strings.TrimSpace(fqbn)==""{return out,fmt.Errorf("board FQBN is required")}
 if s.Builder.CLI==nil{return out,fmt.Errorf("build service requires Arduino CLI")}
 if s.Engineer==nil{return out,fmt.Errorf("AI engineer is required")}
 if s.Parser==nil{return out,fmt.Errorf("diagnostic parser is required")}
 if s.Files==nil{return out,fmt.Errorf("project files adapter is required")}
 max:=s.MaxAttempts;if max<=0{max=3}
 original,err:=s.Files.Snapshot(ctx,p);if err!=nil{return out,fmt.Errorf("snapshot project: %w",err)}
 if len(original)==0{return out,fmt.Errorf("project source snapshot is empty")}
 working:=cloneSnapshot(original)
 for attempt:=0;attempt<=max;attempt++{
  if err:=ctx.Err();err!=nil{return out,err}
  result,buildErr:=(BuildService{CLI:s.Builder.CLI}).Compile(ctx,p,fqbn)
  out.FinalBuild=result
  if buildErr==nil&&result.Status==build.StatusPassed&&result.ExitCode==0{
   out.Recovered=out.Attempts>0
   return out,nil
  }
  raw:=result.Output;if buildErr!=nil{raw+="\n"+buildErr.Error()}
  diagnostics:=s.Parser.Parse(raw);out.Diagnostics=diagnostics
  if attempt==max{
   if restoreErr:=s.Files.Replace(ctx,p,original);restoreErr!=nil{return out,fmt.Errorf("%w; restoring original source failed: %v",ErrRepairAttemptsExhausted,restoreErr)}
   return out,fmt.Errorf("%w after %d attempts",ErrRepairAttemptsExhausted,max)
  }
  hasError:=false;for _,d:=range diagnostics{if d.IsError(){hasError=true;break}}
  if !hasError{return out,fmt.Errorf("build failed but no compiler error diagnostics were parsed; refusing speculative repair")}
  encoded,err:=json.Marshal(working);if err!=nil{return out,err}
  patch,err:=s.Engineer.Diagnose(ctx,diagnostics,string(encoded));if err!=nil{return out,fmt.Errorf("diagnose build failure: %w",err)}
  if err:=patch.Validate();err!=nil{return out,fmt.Errorf("AI returned invalid patch: %w",err)}
  patched,err:=repair.Apply(working,patch);if err!=nil{return out,fmt.Errorf("reject unsafe or stale repair patch: %w",err)}
  if err:=s.Files.Replace(ctx,p,patched);err!=nil{return out,fmt.Errorf("persist repair patch: %w",err)}
  working=patched;out.Patches=append(out.Patches,patch);out.Attempts++
 }
 return out,ErrRepairAttemptsExhausted
}
func cloneSnapshot(src map[string]string)map[string]string{dst:=make(map[string]string,len(src));for k,v:=range src{dst[k]=v};return dst}
func SnapshotText(src map[string]string)string{
 keys:=make([]string,0,len(src));for k:=range src{keys=append(keys,k)};sort.Strings(keys)
 var b strings.Builder
 for _,k:=range keys{b.WriteString("=== ");b.WriteString(k);b.WriteString(" ===\n");b.WriteString(src[k]);if !strings.HasSuffix(src[k],"\n"){b.WriteByte('\n')}}
 return b.String()
}

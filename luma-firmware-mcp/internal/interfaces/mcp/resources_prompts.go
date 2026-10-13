package mcp

import (
 "context"
 "encoding/json"
 "fmt"
 "sort"
 "strings"

 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/application"
)

type Resource struct { URI string `json:"uri"`; Name string `json:"name"`; Description string `json:"description,omitempty"`; MIMEType string `json:"mimeType,omitempty"` }
type ResourceContent struct { URI string `json:"uri"`; MIMEType string `json:"mimeType,omitempty"`; Text string `json:"text"` }
type ResourceHandler func(context.Context)(string,error)
type RegisteredResource struct { Definition Resource; Read ResourceHandler }
type ResourceRegistry struct { items map[string]RegisteredResource }
func NewResourceRegistry()*ResourceRegistry{return &ResourceRegistry{items:map[string]RegisteredResource{}}}
func(r *ResourceRegistry)Register(item RegisteredResource)error{
 if strings.TrimSpace(item.Definition.URI)==""||strings.TrimSpace(item.Definition.Name)==""||item.Read==nil{return fmt.Errorf("resource URI, name and reader are required")}
 if r.items==nil{r.items=map[string]RegisteredResource{}}
 if _,ok:=r.items[item.Definition.URI];ok{return fmt.Errorf("resource %q already registered",item.Definition.URI)}
 r.items[item.Definition.URI]=item;return nil
}
func(r *ResourceRegistry)List()[]Resource{
 keys:=make([]string,0,len(r.items));for k:=range r.items{keys=append(keys,k)};sort.Strings(keys)
 out:=make([]Resource,0,len(keys));for _,k:=range keys{out=append(out,r.items[k].Definition)};return out
}
func(r *ResourceRegistry)Read(ctx context.Context,uri string)(ResourceContent,error){
 item,ok:=r.items[uri];if !ok{return ResourceContent{},fmt.Errorf("unknown resource %q",uri)}
 text,err:=item.Read(ctx);if err!=nil{return ResourceContent{},err}
 return ResourceContent{URI:uri,MIMEType:item.Definition.MIMEType,Text:text},nil
}
type PromptArgument struct { Name string `json:"name"`; Description string `json:"description,omitempty"`; Required bool `json:"required,omitempty"` }
type Prompt struct { Name string `json:"name"`; Description string `json:"description,omitempty"`; Arguments []PromptArgument `json:"arguments,omitempty"` }
type PromptMessage struct { Role string `json:"role"`; Content map[string]any `json:"content"` }
type PromptResult struct { Description string `json:"description,omitempty"`; Messages []PromptMessage `json:"messages"` }
type PromptHandler func(map[string]string)(PromptResult,error)
type RegisteredPrompt struct { Definition Prompt; Get PromptHandler }
type PromptRegistry struct { items map[string]RegisteredPrompt }
func NewPromptRegistry()*PromptRegistry{return &PromptRegistry{items:map[string]RegisteredPrompt{}}}
func(r *PromptRegistry)Register(p RegisteredPrompt)error{
 if strings.TrimSpace(p.Definition.Name)==""||p.Get==nil{return fmt.Errorf("prompt name and handler are required")}
 if r.items==nil{r.items=map[string]RegisteredPrompt{}}
 if _,ok:=r.items[p.Definition.Name];ok{return fmt.Errorf("prompt %q already registered",p.Definition.Name)}
 r.items[p.Definition.Name]=p;return nil
}
func(r *PromptRegistry)List()[]Prompt{
 keys:=make([]string,0,len(r.items));for k:=range r.items{keys=append(keys,k)};sort.Strings(keys)
 out:=make([]Prompt,0,len(keys));for _,k:=range keys{out=append(out,r.items[k].Definition)};return out
}
func(r *PromptRegistry)Get(name string,args map[string]string)(PromptResult,error){
 p,ok:=r.items[name];if !ok{return PromptResult{},fmt.Errorf("unknown prompt %q",name)}
 for _,a:=range p.Definition.Arguments{if a.Required&&strings.TrimSpace(args[a.Name])==""{return PromptResult{},fmt.Errorf("prompt argument %q is required",a.Name)}}
 return p.Get(args)
}
func textPrompt(description,body string,args ...PromptArgument)RegisteredPrompt{
 return RegisteredPrompt{Definition:Prompt{Description:description,Arguments:args,Name:strings.ReplaceAll(strings.ToLower(strings.ReplaceAll(description," ","_")),"/","_")},Get:func(values map[string]string)(PromptResult,error){
  for k,v:=range values{body=strings.ReplaceAll(body,"{{"+k+"}}",v)}
  return PromptResult{Description:description,Messages:[]PromptMessage{{Role:"user",Content:map[string]any{"type":"text","text":body}}}},nil
 }}
}
func RegisterDefaultPrompts(r *PromptRegistry)error{
 prompts:=[]RegisteredPrompt{
  {Definition:Prompt{Name:"generate_esp32_lamp_firmware",Description:"Create a conservative ESP32 lamp firmware plan from a validated specification.",Arguments:[]PromptArgument{{Name:"specification",Description:"JSON firmware specification with explicit lamp GPIO assignments.",Required:true}}},Get:func(a map[string]string)(PromptResult,error){return PromptResult{Description:"Firmware generation instructions",Messages:[]PromptMessage{{Role:"user",Content:map[string]any{"type":"text","text":"Generate compilable ESP32 Arduino firmware using only the supplied specification. Preserve lamp IDs, names, rooms and GPIO assignments exactly. Do not invent GPIOs or credentials. Keep secrets out of source. Use explicit HIGH/LOW writes and safe startup. Return structured files and identify assumptions. Specification:\n"+a["specification"]}}}},nil}},
  {Definition:Prompt{Name:"diagnose_compiler_errors",Description:"Analyze compiler output and propose a minimal evidence-based fix.",Arguments:[]PromptArgument{{Name:"source",Description:"Relevant source code.",Required:true},{Name:"diagnostics",Description:"Full compiler output.",Required:true}}},Get:func(a map[string]string)(PromptResult,error){return PromptResult{Messages:[]PromptMessage{{Role:"user",Content:map[string]any{"type":"text","text":"Diagnose the compiler errors using only the supplied source and diagnostics. Separate confirmed causes from hypotheses. Propose the smallest patch; do not claim it compiles until rebuilt. Preserve behavior and do not add credentials.\nSOURCE:\n"+a["source"]+"\nDIAGNOSTICS:\n"+a["diagnostics"]}}}},nil}},
  {Definition:Prompt{Name:"propose_safe_repair",Description:"Propose a structured, reviewable firmware repair patch.",Arguments:[]PromptArgument{{Name:"source",Description:"Current project files or source.",Required:true},{Name:"diagnostics",Description:"Build diagnostics and observed behavior.",Required:true}}},Get:func(a map[string]string)(PromptResult,error){return PromptResult{Messages:[]PromptMessage{{Role:"user",Content:map[string]any{"type":"text","text":"Return a structured JSON repair proposal only. Use create/update/delete file operations. For updates and deletes include SHA-256 of the exact current file content as expected_sha. Do not use a display-only diff as an executable patch. Make the smallest safe change and do not apply it.\nSOURCE:\n"+a["source"]+"\nDIAGNOSTICS:\n"+a["diagnostics"]}}}},nil}},
 }
 for _,p:=range prompts{if err:=r.Register(p);err!=nil{return err}}
 return nil
}
func RegisterDefaultResources(r *ResourceRegistry,cli application.ArduinoCLI)error{
 if cli==nil{return fmt.Errorf("Arduino CLI adapter is required")}
 if err:=r.Register(RegisteredResource{Definition:Resource{URI:"luma://toolchain/arduino-cli",Name:"Arduino CLI toolchain status",Description:"Installed Arduino CLI version; does not expose environment variables or credentials.",MIMEType:"application/json"},Read:func(ctx context.Context)(string,error){v,err:=cli.Version(ctx);if err!=nil{return "",err};b,_:=json.Marshal(map[string]string{"tool":"arduino-cli","version":v});return string(b),nil}});err!=nil{return err}
 if err:=r.Register(RegisteredResource{Definition:Resource{URI:"luma://boards/detected",Name:"Detected Arduino boards",Description:"Current board discovery results from Arduino CLI.",MIMEType:"application/json"},Read:func(ctx context.Context)(string,error){v,err:=cli.ListBoards(ctx);if err!=nil{return "",err};b,err:=json.Marshal(v);return string(b),err}});err!=nil{return err}
 return r.Register(RegisteredResource{Definition:Resource{URI:"luma://firmware/specification-guidance",Name:"Firmware specification guidance",Description:"Safe firmware-specification constraints; not a live project record.",MIMEType:"text/plain"},Read:func(context.Context)(string,error){return "Supply project ID, microcontroller name, target, firmware version, and explicit lamp IDs/names/GPIOs. Default status GPIO is 13. GPIO assignments must be unique and must not collide with status GPIO. Common ESP32 GPIO 6-11 are reserved. Never place secrets in generated source. Validate the target board's actual pin restrictions before flashing.",nil}})
}

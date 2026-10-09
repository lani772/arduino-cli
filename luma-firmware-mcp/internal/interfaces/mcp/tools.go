package mcp

import (
 "context"
 "encoding/json"
 "fmt"
 "sort"
 "strings"

 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/application"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/firmware"
)

type Tool struct { Name string `json:"name"`; Description string `json:"description"`; InputSchema map[string]any `json:"inputSchema"` }
type ToolResult struct { Content []map[string]any `json:"content"`; IsError bool `json:"isError,omitempty"` }
type ToolHandler func(context.Context,json.RawMessage)(any,error)
type RegisteredTool struct { Definition Tool; Handler ToolHandler }
type Registry struct { tools map[string]RegisteredTool }
func NewRegistry()*Registry{return &Registry{tools:map[string]RegisteredTool{}}}
func(r *Registry)Register(t RegisteredTool)error{
 if strings.TrimSpace(t.Definition.Name)==""||t.Handler==nil{return fmt.Errorf("tool name and handler are required")}
 if r.tools==nil{r.tools=map[string]RegisteredTool{}}
 if _,ok:=r.tools[t.Definition.Name];ok{return fmt.Errorf("tool %q already registered",t.Definition.Name)}
 r.tools[t.Definition.Name]=t;return nil
}
func(r *Registry)List()[]Tool{
 names:=make([]string,0,len(r.tools));for name:=range r.tools{names=append(names,name)};sort.Strings(names)
 out:=make([]Tool,0,len(names));for _,name:=range names{out=append(out,r.tools[name].Definition)};return out
}
func(r *Registry)Call(ctx context.Context,name string,args json.RawMessage)(ToolResult,error){
 tool,ok:=r.tools[name];if !ok{return ToolResult{},fmt.Errorf("unknown tool %q",name)}
 value,err:=tool.Handler(ctx,args)
 if err!=nil{return ToolResult{Content:[]map[string]any{{"type":"text","text":err.Error()}},IsError:true},nil}
 b,err:=json.Marshal(value);if err!=nil{return ToolResult{},fmt.Errorf("encode tool result: %w",err)}
 return ToolResult{Content:[]map[string]any{{"type":"text","text":string(b)}}},nil
}
func objectSchema(properties map[string]any,required ...string)map[string]any{
 schema:=map[string]any{"type":"object","properties":properties,"additionalProperties":false}
 if len(required)>0{schema["required"]=required};return schema
}
func RegisterArduinoTools(r *Registry,cli application.ArduinoCLI)error{
 if cli==nil{return fmt.Errorf("Arduino CLI adapter is required")}
 if err:=r.Register(RegisteredTool{Definition:Tool{Name:"arduino_cli_version",Description:"Return the installed Arduino CLI version.",InputSchema:objectSchema(map[string]any{})},Handler:func(ctx context.Context,_ json.RawMessage)(any,error){return cli.Version(ctx)}});err!=nil{return err}
 if err:=r.Register(RegisteredTool{Definition:Tool{Name:"arduino_list_boards",Description:"List boards currently detected by Arduino CLI. This does not flash or modify a device.",InputSchema:objectSchema(map[string]any{})},Handler:func(ctx context.Context,_ json.RawMessage)(any,error){return cli.ListBoards(ctx)}});err!=nil{return err}
 return r.Register(RegisteredTool{Definition:Tool{Name:"arduino_compile",Description:"Compile an existing Arduino sketch for a specified FQBN. This does not upload or flash a device.",InputSchema:objectSchema(map[string]any{"source_path":map[string]any{"type":"string","description":"Path to an existing sketch directory or source file."},"fqbn":map[string]any{"type":"string","description":"Fully qualified board name, for example esp32:esp32:esp32."}},"source_path","fqbn")},Handler:func(ctx context.Context,raw json.RawMessage)(any,error){
  var p struct{SourcePath string `json:"source_path"`;FQBN string `json:"fqbn"`}
  if err:=json.Unmarshal(raw,&p);err!=nil{return nil,fmt.Errorf("invalid compile arguments: %w",err)}
  if strings.TrimSpace(p.SourcePath)==""||strings.TrimSpace(p.FQBN)==""{return nil,fmt.Errorf("source_path and fqbn are required")}
  result,err:=cli.Compile(ctx,p.SourcePath,p.FQBN)
  response:=map[string]any{"build":result}
  if err!=nil{response["error"]=err.Error()}
  return response,nil
 }})
}
func RegisterUploadTool(r *Registry, cli application.ArduinoCLI) error {
 if cli == nil { return fmt.Errorf("Arduino CLI adapter is required") }
 return r.Register(RegisteredTool{
  Definition: Tool{Name:"arduino_upload_firmware",Description:"Upload a compiled sketch to an explicitly selected serial port. Requires confirm_upload=true and an execution policy that permits device flashing; flashing is disabled by default.",InputSchema:objectSchema(map[string]any{
   "source_path":map[string]any{"type":"string","description":"Path to an existing Arduino sketch."},
   "fqbn":map[string]any{"type":"string","description":"Exact target board FQBN."},
   "port":map[string]any{"type":"string","description":"Exact serial port address returned by device discovery."},
   "confirm_upload":map[string]any{"type":"boolean","description":"Must be true to authorize this upload request."},
  },"source_path","fqbn","port","confirm_upload")},
  Handler:func(ctx context.Context,raw json.RawMessage)(any,error){
   var p struct{SourcePath string `json:"source_path"`;FQBN string `json:"fqbn"`;Port string `json:"port"`;Confirm bool `json:"confirm_upload"`}
   if err:=json.Unmarshal(raw,&p);err!=nil{return nil,fmt.Errorf("invalid upload arguments: %w",err)}
   if strings.TrimSpace(p.SourcePath)==""||strings.TrimSpace(p.FQBN)==""||strings.TrimSpace(p.Port)==""{return nil,fmt.Errorf("source_path, fqbn, and port are required")}
   if !p.Confirm{return nil,fmt.Errorf("upload requires confirm_upload=true")}
   result,err:=cli.Upload(ctx,p.SourcePath,p.FQBN,p.Port)
   response:=map[string]any{"upload":result}
   if err!=nil{response["error"]=err.Error()}
   return response,nil
  },
 })
}
func RegisterFirmwareValidationTool(r *Registry)error{
 properties:=map[string]any{
  "project_id":map[string]any{"type":"string"},"microcontroller_name":map[string]any{"type":"string"},"target":map[string]any{"type":"object"},
  "firmware_version":map[string]any{"type":"string"},"status_gpio":map[string]any{"type":"integer"},"offline_schedules":map[string]any{"type":"boolean"},"mqtt_enabled":map[string]any{"type":"boolean"},
  "lamps":map[string]any{"type":"array","items":map[string]any{"type":"object","properties":map[string]any{"id":map[string]any{"type":"string"},"name":map[string]any{"type":"string"},"room":map[string]any{"type":"string"},"gpio":map[string]any{"type":"integer"}},"required":[]string{"id","name","gpio"},"additionalProperties":false}},
 }
 schema:=objectSchema(properties,"project_id","microcontroller_name","target","firmware_version","lamps")
 return r.Register(RegisteredTool{Definition:Tool{Name:"luma_validate_firmware_spec",Description:"Validate a proposed LUMA firmware specification and its GPIO assignments without generating or flashing firmware.",InputSchema:schema},Handler:func(_ context.Context,raw json.RawMessage)(any,error){
  var spec firmware.Specification
  if err:=json.Unmarshal(raw,&spec);err!=nil{return nil,fmt.Errorf("invalid firmware specification: %w",err)}
  spec=spec.WithDefaults()
  if err:=spec.Validate();err!=nil{return map[string]any{"valid":false,"error":err.Error()},nil}
  return map[string]any{"valid":true,"status_gpio":spec.StatusGPIO,"lamp_count":len(spec.Lamps)},nil
 }})
}

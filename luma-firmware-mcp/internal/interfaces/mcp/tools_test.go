package mcp

import (
 "bytes"
 "context"
 "encoding/json"
 "errors"
 "strings"
 "testing"
)
func TestRegistryListIsSortedAndCallReturnsContent(t *testing.T){
 r:=NewRegistry()
 for _,name:=range []string{"z_tool","a_tool"}{name:=name
  err:=r.Register(RegisteredTool{Definition:Tool{Name:name,InputSchema:objectSchema(map[string]any{})},Handler:func(context.Context,json.RawMessage)(any,error){return map[string]string{"tool":name},nil}})
  if err!=nil{t.Fatal(err)}
 }
 listed:=r.List();if len(listed)!=2||listed[0].Name!="a_tool"{t.Fatalf("not sorted: %+v",listed)}
 result,err:=r.Call(context.Background(),"a_tool",nil);if err!=nil{t.Fatal(err)}
 if result.IsError||len(result.Content)!=1{t.Fatalf("unexpected result: %+v",result)}
}
func TestRegistryHandlerErrorIsToolError(t *testing.T){
 r:=NewRegistry();_ = r.Register(RegisteredTool{Definition:Tool{Name:"bad"},Handler:func(context.Context,json.RawMessage)(any,error){return nil,errors.New("failed safely")}})
 result,err:=r.Call(context.Background(),"bad",nil);if err!=nil{t.Fatal(err)}
 if !result.IsError{t.Fatal("expected isError")}
}
func TestRegistryRejectsDuplicate(t *testing.T){
 r:=NewRegistry();tool:=RegisteredTool{Definition:Tool{Name:"same"},Handler:func(context.Context,json.RawMessage)(any,error){return nil,nil}}
 if err:=r.Register(tool);err!=nil{t.Fatal(err)}
 if err:=r.Register(tool);err==nil{t.Fatal("expected duplicate rejection")}
}
func TestToolsCallViaStdio(t *testing.T){
 r:=NewRegistry()
 if err:=r.Register(RegisteredTool{Definition:Tool{Name:"echo",InputSchema:objectSchema(map[string]any{"value":map[string]any{"type":"string"}},"value")},Handler:func(_ context.Context,raw json.RawMessage)(any,error){var p map[string]string;if err:=json.Unmarshal(raw,&p);err!=nil{return nil,err};return p,nil}});err!=nil{t.Fatal(err)}
 input:=`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`+"\n"+`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"echo","arguments":{"value":"ok"}}}`+"\n"
 var output bytes.Buffer
 s:=NewStdio(strings.NewReader(input),&output);s.Registry=r
 if err:=s.Start(context.Background());err!=nil{t.Fatal(err)}
 lines:=strings.Split(strings.TrimSpace(output.String()),"\n");if len(lines)!=2{t.Fatalf("responses=%d",len(lines))}
 var call Response;if err:=json.Unmarshal([]byte(lines[1]),&call);err!=nil{t.Fatal(err)}
 if call.Error!=nil{t.Fatalf("call error: %+v",call.Error)}
 result:=call.Result.(map[string]any);content:=result["content"].([]any)
 if !strings.Contains(content[0].(map[string]any)["text"].(string),"ok"){t.Fatalf("unexpected tool content: %#v",content)}
}

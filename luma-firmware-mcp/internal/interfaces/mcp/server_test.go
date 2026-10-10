package mcp

import (
 "bytes"
 "context"
 "encoding/json"
 "strings"
 "testing"
)
func TestStdioServerInitializePingAndToolsList(t *testing.T){
 input:=strings.Join([]string{
  `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`,
  `{"jsonrpc":"2.0","method":"notifications/initialized"}`,
  `{"jsonrpc":"2.0","id":2,"method":"ping"}`,
  `{"jsonrpc":"2.0","id":3,"method":"tools/list"}`,
 },"\n")+"\n"
 var output bytes.Buffer
 if err:=NewStdio(strings.NewReader(input),&output).Start(context.Background());err!=nil{t.Fatal(err)}
 lines:=strings.Split(strings.TrimSpace(output.String()),"\n")
 if len(lines)!=3{t.Fatalf("expected 3 responses, got %d",len(lines))}
 var init Response
 if err:=json.Unmarshal([]byte(lines[0]),&init);err!=nil{t.Fatal(err)}
 if init.Error!=nil{t.Fatalf("initialize failed: %+v",init.Error)}
 result:=init.Result.(map[string]any)
 if result["protocolVersion"]!=ProtocolVersion{t.Fatalf("unexpected protocol version: %#v",result)}
 var tools Response
 if err:=json.Unmarshal([]byte(lines[2]),&tools);err!=nil{t.Fatal(err)}
 if tools.Error!=nil{t.Fatalf("tools/list failed: %+v",tools.Error)}
}
func TestStdioServerUnknownMethodAndParseError(t *testing.T){
 input:=`{"jsonrpc":"2.0","id":"x","method":"unknown"}`+"\n"+`{bad json}`+"\n"
 var output bytes.Buffer
 if err:=NewStdio(strings.NewReader(input),&output).Start(context.Background());err!=nil{t.Fatal(err)}
 lines:=strings.Split(strings.TrimSpace(output.String()),"\n")
 if len(lines)!=2{t.Fatalf("expected two error responses, got %d",len(lines))}
 var a,b Response
 _=json.Unmarshal([]byte(lines[0]),&a);_=json.Unmarshal([]byte(lines[1]),&b)
 if a.Error==nil||a.Error.Code!=-32601{t.Fatalf("unexpected method error: %+v",a.Error)}
 if b.Error==nil||b.Error.Code!=-32700{t.Fatalf("unexpected parse error: %+v",b.Error)}
}
func TestStdioServerRejectsMissingStreams(t *testing.T){
 if err:=(&StdioServer{}).Start(context.Background());err==nil{t.Fatal("expected missing stream error")}
}

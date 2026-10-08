package mcp

import (
 "bufio"
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "sync"
)

const ProtocolVersion = "2024-11-05"

type Server interface { Start(context.Context) error; Shutdown(context.Context) error }
type Request struct { JSONRPC string `json:"jsonrpc"`; ID json.RawMessage `json:"id,omitempty"`; Method string `json:"method"`; Params json.RawMessage `json:"params,omitempty"` }
type Response struct { JSONRPC string `json:"jsonrpc"`; ID json.RawMessage `json:"id"`; Result any `json:"result,omitempty"`; Error *RPCError `json:"error,omitempty"` }
type RPCError struct { Code int `json:"code"`; Message string `json:"message"`; Data any `json:"data,omitempty"` }
type StdioServer struct { In io.Reader; Out io.Writer; MaxMessageBytes int; writeMu sync.Mutex }
func NewStdio(in io.Reader,out io.Writer)*StdioServer{return &StdioServer{In:in,Out:out,MaxMessageBytes:4*1024*1024}}
func(s *StdioServer)Start(ctx context.Context)error{
 if s.In==nil||s.Out==nil{return errors.New("MCP stdio input and output are required")}
 limit:=s.MaxMessageBytes;if limit<=0{limit=4*1024*1024}
 scanner:=bufio.NewScanner(s.In);scanner.Buffer(make([]byte,4096),limit)
 for scanner.Scan(){
  if err:=ctx.Err();err!=nil{return err}
  var req Request
  if err:=json.Unmarshal(scanner.Bytes(),&req);err!=nil{
   if e:=s.write(Response{JSONRPC:"2.0",ID:json.RawMessage("null"),Error:&RPCError{Code:-32700,Message:"Parse error"}});e!=nil{return e};continue
  }
  if req.JSONRPC!="2.0"||req.Method==""{
   if len(req.ID)>0{if e:=s.write(Response{JSONRPC:"2.0",ID:req.ID,Error:&RPCError{Code:-32600,Message:"Invalid Request"}});e!=nil{return e}};continue
  }
  result,rpcErr:=s.handle(req)
  if len(req.ID)==0{continue}
  if e:=s.write(Response{JSONRPC:"2.0",ID:req.ID,Result:result,Error:rpcErr});e!=nil{return e}
 }
 if err:=scanner.Err();err!=nil{return fmt.Errorf("read MCP stdio: %w",err)}
 return nil
}
func(s *StdioServer)Shutdown(context.Context)error{return nil}
func(s *StdioServer)handle(req Request)(any,*RPCError){
 switch req.Method{
 case "initialize":
  var p struct{ProtocolVersion string `json:"protocolVersion"`}
  if len(req.Params)>0{if err:=json.Unmarshal(req.Params,&p);err!=nil{return nil,&RPCError{Code:-32602,Message:"Invalid params"}}}
  version:=p.ProtocolVersion;if version==""{version=ProtocolVersion}
  return map[string]any{"protocolVersion":version,"capabilities":map[string]any{"tools":map[string]any{"listChanged":false},"resources":map[string]any{},"prompts":map[string]any{}},"serverInfo":map[string]string{"name":"luma-firmware-mcp","version":"0.1.0"}},nil
 case "notifications/initialized": return nil,nil
 case "ping": return map[string]any{},nil
 case "tools/list": return map[string]any{"tools":[]any{}},nil
 case "resources/list": return map[string]any{"resources":[]any{}},nil
 case "prompts/list": return map[string]any{"prompts":[]any{}},nil
 default: return nil,&RPCError{Code:-32601,Message:"Method not found",Data:map[string]string{"method":req.Method}}
 }
}
func(s *StdioServer)write(v Response)error{
 s.writeMu.Lock();defer s.writeMu.Unlock()
 b,err:=json.Marshal(v);if err!=nil{return err}
 if _,err=s.Out.Write(append(b,'\n'));err!=nil{return fmt.Errorf("write MCP stdio: %w",err)}
 return nil
}

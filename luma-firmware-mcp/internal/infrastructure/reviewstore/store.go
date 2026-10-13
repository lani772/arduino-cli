package reviewstore

import (
 "encoding/json"
 "fmt"
 "os"
 "path/filepath"
 "strings"
 "sync"
 "time"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/project"
 "github.com/lani772/arduino-cli/luma-firmware-mcp/internal/domain/repair"
)

type Proposal struct { Project project.Project `json:"project"`; FQBN string `json:"fqbn"`; Patch repair.Patch `json:"patch"`; CreatedAt time.Time `json:"created_at"` }
type Event struct { PatchID string `json:"patch_id"`; ProjectID string `json:"project_id"`; Action string `json:"action"`; Actor string `json:"actor"`; At time.Time `json:"at"`; Detail string `json:"detail,omitempty"` }
type State struct { Proposals map[string]Proposal `json:"proposals"`; Decisions map[string]string `json:"decisions"`; Audit []Event `json:"audit"` }
type Store struct { Path string; mu sync.Mutex; state State }

func New(path string) (*Store,error) {
 if strings.TrimSpace(path)=="" { return nil,fmt.Errorf("review journal path is required") }
 abs,err:=filepath.Abs(path);if err!=nil{return nil,err}
 s:=&Store{Path:abs,state:State{Proposals:map[string]Proposal{},Decisions:map[string]string{},Audit:[]Event{}}}
 if err:=s.checkPath();err!=nil{return nil,err}
 data,err:=os.ReadFile(abs);if os.IsNotExist(err){return s,nil};if err!=nil{return nil,err}
 if err:=json.Unmarshal(data,&s.state);err!=nil{return nil,fmt.Errorf("decode review journal: %w",err)}
 if s.state.Proposals==nil{s.state.Proposals=map[string]Proposal{}};if s.state.Decisions==nil{s.state.Decisions=map[string]string{}}
 return s,nil
}
func(s *Store) checkPath()error{
 if err:=os.MkdirAll(filepath.Dir(s.Path),0700);err!=nil{return err}
 p:=filepath.Dir(s.Path);i,e:=os.Lstat(p);if e!=nil{return e};if !i.IsDir()||i.Mode()&os.ModeSymlink!=0{return fmt.Errorf("review journal parent must be a real directory")}
 i,e=os.Lstat(s.Path);if e==nil&&(i.Mode()&os.ModeSymlink!=0||!i.Mode().IsRegular()){return fmt.Errorf("review journal must be a regular file")};if e!=nil&&!os.IsNotExist(e){return e};return nil
}
func(s *Store) Snapshot()State{s.mu.Lock();defer s.mu.Unlock();return clone(s.state)}
func(s *Store) SaveProposal(id string,p Proposal)error{
 s.mu.Lock();defer s.mu.Unlock()
 if _,ok:=s.state.Proposals[id];ok{return fmt.Errorf("patch id already pending")};if _,ok:=s.state.Decisions[id];ok{return fmt.Errorf("patch id has already been decided")}
 n:=clone(s.state);n.Proposals[id]=p;n.Audit=append(n.Audit,Event{PatchID:id,ProjectID:p.Project.ID,Action:"proposed",Actor:"mcp-client",At:time.Now().UTC()})
 if err:=s.persist(n);err!=nil{return err};s.state=n;return nil
}
func(s *Store) BeginDecision(id,decision string)(Proposal,error){
 s.mu.Lock();defer s.mu.Unlock()
 p,ok:=s.state.Proposals[id];if !ok{return Proposal{},fmt.Errorf("pending patch not found")}
 if decision!="approved"&&decision!="rejected"{return Proposal{},fmt.Errorf("invalid decision")}
 n:=clone(s.state);delete(n.Proposals,id);n.Decisions[id]="processing";n.Audit=append(n.Audit,Event{PatchID:id,ProjectID:p.Project.ID,Action:decision+"_started",Actor:"mcp-client",At:time.Now().UTC()})
 if err:=s.persist(n);err!=nil{return Proposal{},err};s.state=n;return p,nil
}
func(s *Store) CompleteDecision(id,decision,detail string)error{
 s.mu.Lock();defer s.mu.Unlock();n:=clone(s.state);if n.Decisions[id]!="processing"{return fmt.Errorf("patch decision is not processing")}
 n.Decisions[id]=decision;n.Audit=append(n.Audit,Event{PatchID:id,Action:decision,Actor:"mcp-client",At:time.Now().UTC(),Detail:detail})
 if err:=s.persist(n);err!=nil{return err};s.state=n;return nil
}
func(s *Store) RestoreProposal(id string,p Proposal,detail string)error{
 s.mu.Lock();defer s.mu.Unlock();n:=clone(s.state);delete(n.Decisions,id);n.Proposals[id]=p;n.Audit=append(n.Audit,Event{PatchID:id,ProjectID:p.Project.ID,Action:"approval_failed_proposal_restored",Actor:"system",At:time.Now().UTC(),Detail:detail})
 if err:=s.persist(n);err!=nil{return err};s.state=n;return nil
}
func(s *Store) persist(n State)error{
 if err:=s.checkPath();err!=nil{return err};b,err:=json.MarshalIndent(n,"","  ");if err!=nil{return err}
 f,err:=os.CreateTemp(filepath.Dir(s.Path),".review-*.tmp");if err!=nil{return err};name:=f.Name();defer os.Remove(name)
 if err=f.Chmod(0600);err!=nil{f.Close();return err};if _,err=f.Write(b);err!=nil{f.Close();return err};if err=f.Sync();err!=nil{f.Close();return err};if err=f.Close();err!=nil{return err}
 return os.Rename(name,s.Path)
}
func clone(in State)State{n:=State{Proposals:map[string]Proposal{},Decisions:map[string]string{},Audit:append([]Event(nil),in.Audit...)};for k,v:=range in.Proposals{n.Proposals[k]=v};for k,v:=range in.Decisions{n.Decisions[k]=v};return n}

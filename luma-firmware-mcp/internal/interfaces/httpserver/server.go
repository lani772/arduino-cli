package httpserver

import (
 "encoding/json"
 "log"
 "net/http"
)

type Server struct { httpServer *http.Server; version string; logger *log.Logger }

func New(addr, version string, logger *log.Logger) *Server {
 mux := http.NewServeMux()
 s := &Server{httpServer:&http.Server{Addr:addr,Handler:mux},version:version,logger:logger}
 mux.HandleFunc("/healthz",s.health)
 return s
}

func (s *Server) Start() error { return s.httpServer.ListenAndServe() }

func (s *Server) health(w http.ResponseWriter,_ *http.Request) {
 w.Header().Set("Content-Type","application/json")
 _ = json.NewEncoder(w).Encode(map[string]string{"service":"luma-firmware-mcp","version":s.version,"status":"ready"})
}

package server

import (
    "log"
    "net"

    "github.com/albertofem/gloster/config"
    n "github.com/albertofem/gloster/node"
)

type Server struct {
    config      *config.Config
    node        n.Node
    exit        chan bool
}

func New(c *config.Config) *Server {
    return &Server{
        config:      c,
        exit:        make(chan bool),
        node:        c.Addr,
    }
}

func (s *Server) Init() {
    defer close(s.exit)

    s.startSocket()
    s.run()
}

func (s *Server) Shutdown() {
    s.exit <- true
}

func (s *Server) run() {
    for {
        select {
        case <-s.exit:
            return
        }
    }
}

func (s *Server) startSocket() {
    _, err := net.Listen("tcp", string(s.config.Addr.String()))

    if err != nil {
        log.Println("Error starting Socket Server: ", err)
        return
    }
}

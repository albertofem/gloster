package server

import (
    "log"
    "net"
    "fmt"
    "time"
    "math/rand"
    "net/rpc"

    "github.com/albertofem/gloster/node"
    "github.com/albertofem/gloster/database"
    "github.com/albertofem/gloster/config"
    "github.com/albertofem/gloster/rpcm"
    "github.com/albertofem/gloster/consensus"
)

const (
    Stopped     = "stopped"
    Follower    = "follower"
    Candidate   = "candidate"
    Leader      = "leader"
)

type Server struct {
    config              *config.Config
    node                node.Node
    state               ServerState
    rpcServer           *rpc.Server
    exitChan            chan bool
    connectionsChan     chan net.Conn
    electionTimeoutChan chan bool
}

type ServerState struct {
    currentState        string
    currentTerm         int
    votedFor            node.Node
    commandLog          []database.Command
    commitIndex         int
    lastApplied         int
    votesReceived       int

    // only leaders
    nextIndex           []int
    matchIndex          []int
}

func New(c *config.Config) *Server {
    serverState := ServerState{
        currentState:   Stopped,
        currentTerm:    0,
        votedFor:       nil,
        commandLog:     nil,
        commitIndex:    0,
        lastApplied:    0,
        nextIndex:      nil,
        matchIndex:     nil,
    }

    return &Server{
        config:                 c,
        node:                   c.Addr,
        state:                  serverState,
        exitChan:               make(chan bool),
        electionTimeoutChan:    make(chan bool),
    }
}

func (server *Server) Init() {
    defer close(server.exitChan)

    server.serve()
    server.run()
}

func (server *Server) Stop() {
    server.state.currentState = Stopped
}

func (server *Server) Shutdown() {
    server.exitChan <- true
}

func (server *Server) IsStopped() {
    return server.state.currentState == Stopped
}

func (server *Server) CurrentState() {
    return server.state.currentState
}

func (server *Server) Accept(connection net.Conn) {
    go server.rpcServer.ServeConn(connection)
}

func (server *Server) ResetElectionTimeout() {
    go func() {
        rand.Seed(time.Now().Unix())
        randomTimeout := rand.Intn(server.config.ElectionTimeout + 1 - server.config.ElectionTimeout) + server.config.ElectionTimeout

        time.Sleep(time.Duration(randomTimeout) * time.Millisecond)
        server.electionTimeoutChan <- true
    }()
}

func (server *Server) IsFollower() {
    return server.state.currentState == Follower
}

func (server *Server) IsCandidate() {
    return server.state.currentState == Candidate
}

func (server *Server) IncrementVote() {
    server.state.votesReceived++
}

func (server *Server) ResetElection() {
    server.state.votedFor = nil
    server.state.currentState = Follower
    server.state.votesReceived = 0
}

func (server *Server) State() {
    return server.state.currentState
}

func (server *Server) IsLeader() {
    return server.state.currentState == Leader
}

func (server *Server) LogState(state string) {
    fmt.Println("> (" + server.State() + ") " + state)
}

func (server *Server) SetTerm(term int) {
    server.state.currentTerm = term
}

func (server *Server) VoteFor(node *node.Node) {
    server.state.votedFor = node
}

func (server *Server) VoteForSelf() {
    server.state.votedFor = server.node
}

func (server *Server) Node() node.Node {
    return server.node
}

func (server *Server) Cluster() ([]node.Node) {
    return server.config.Cluster
}

func (server *Server) CurrentTerm() {
    return server.state.currentTerm
}

func (server *Server) HasVoted() {
    return server.state.votedFor != nil
}

func (server *Server) serve() {

}

func (server *Server) run() {
    server.beginFollow()

    for !server.IsStopped() {
        switch server.state() {
            case Follower:
                consensus.Follow(server, server.exitChan, server.electionTimeoutChan, server.connectionsChan)
            case Candidate:
                consensus.Candidate(server, server.exitChan, server.electionTimeoutChan, server.connectionsChan)
            case Leader:
                consensus.Lead(server, server.exitChan, server.connectionsChan)
        }
    }
}

func (server *Server) beginFollow() {
    server.state.currentState = Follower
}

func (server *Server) openConnections(listener net.Listener) {
    server.connectionsChan = make(chan net.Conn)

    go func() {
        for {
            connection, err := listener.Accept()

            if connection == nil {
                log.Fatal(err)
                continue
            }

            server.LogState("Accepted connection from " + connection.RemoteAddr().String())
            server.connectionsChan <-connection
        }
    }()
}

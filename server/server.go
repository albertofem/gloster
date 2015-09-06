package server

import (
    "log"
    "net"
    "github.com/albertofem/gloster/config"
    n "github.com/albertofem/gloster/node"
    "github.com/boltdb/bolt"
    "fmt"
    "github.com/albertofem/gloster/database"
    "time"
    "net/rpc"
    "github.com/albertofem/gloster/raft_rpc"
    "math/rand"
)

const (
    Stopped = "stopped"
    Follower = "follower"
    Candidate = "candidate"
    Leader = "leader"
)

type Server struct {
    config      *config.Config
    node        n.Node
    database    *bolt.DB
    exit        chan bool
    connections chan net.Conn
    State       ServerState
    electionTimeoutChan chan bool
    rpcServer   *rpc.Server
    votesReceived int
}

type ServerState struct {
    CurrentState string // follower, candidate, leader
    currentTerm int
    votedFor *n.Node
    commandLog []database.Command
    commitIndex int
    lastApplied int

    // only leaders
    nextIndex []int
    matchIndex []int
}

func New(c *config.Config) *Server {
    serverState := ServerState{
        CurrentState: Stopped,
        currentTerm: 0,
        votedFor: nil,
        commandLog: nil, // initialize this
        commitIndex: 0,
        lastApplied: 0,
        nextIndex: nil, // initialize this
        matchIndex: nil, // initialize this
    }

    return &Server{
        config: c,
        exit: make(chan bool),
        node: c.Addr,
        State: serverState,
        electionTimeoutChan: make(chan bool),
    }
}

func (s *Server) Init() {
    defer close(s.exit)

    s.serve()
    s.run()
}

func (s *Server) Shutdown() {
    s.logState("Shuting down...")
    s.exit <- true
}

func (s *Server) serve() {
    s.logState("Serving TCP on: " + s.node.String())

    glosterServer := new(Server)

    s.rpcServer = rpc.NewServer()
    s.rpcServer.Register(glosterServer)

    listener, err := net.Listen("tcp", string(s.config.Addr.String()))

    if err != nil {
        s.logState("Error starting Socket Server: " + err.Error())
        return
    }

    s.openConnections(listener)
}

func (s *Server) run() { // main server loop

    s.State.CurrentState = Follower

    currentState := s.State.CurrentState

    for currentState != Stopped {
        switch currentState {
            case Follower:
                s.follow()
            case Candidate:
                s.candidate()
            case Leader:
                s.lead()
        }

        currentState = s.State.CurrentState
    }
}

func (s *Server) follow() {

    s.logState("Following...")
    s.resetElectionTimeout()

    for s.State.CurrentState == Follower {
        select {
        case <-s.exit:
            s.State.CurrentState = Stopped
            return

        case connection := <-s.connections:
            s.accept(connection)

        case <-s.electionTimeoutChan:
            if (s.State.votedFor == nil) { // no voted casted, go for candidacy
                s.State.CurrentState = Candidate
            }
        }
    }
}

func (s *Server) candidate() {
    s.logState("Candidating...")

    s.State.currentTerm = 1
    s.State.votedFor = &s.node
    s.resetElectionTimeout()

    requestVoteResultChan := make(chan *rpc.Call, 10)

    // send request vote to all peers
    s.sendRequestVote(requestVoteResultChan)

    for s.State.CurrentState == Candidate {
        select {
        case <-s.exit:
            s.State.CurrentState = Stopped
            return

        case connection := <-s.connections:
            s.accept(connection)

        case <-s.electionTimeoutChan:
            if (s.votesReceived+1 >= len(s.config.Cluster)) {
                s.State.CurrentState = Leader
            } else { // reset election
                s.votesReceived = 0
                return
            }

        case requestResult := <-requestVoteResultChan:
            s.handleSendRequestVoteResult(requestResult)
        }
    }
}

func (s *Server) sendRequestVote(requestVoteResultChan chan *rpc.Call) {
    for _, node := range s.config.Cluster {
        client, err := rpc.Dial("tcp", node.String())

        if err != nil {
            s.logState("Cannot connect to peer: " + node.String())
            continue
        }

        requestVote := raft_rpc.RequestVote{
            Term: s.State.currentTerm,
            CandidateId: s.node.String(),
            LastLogIndex: s.State.lastApplied,
            LastLogTerm: s.State.currentTerm,
        }

        var requestVoteResult raft_rpc.RequestVoteResult

        client.Go("Server.RequestVote", requestVote, &requestVoteResult, requestVoteResultChan)
    }
}

func (s *Server) handleSendRequestVoteResult(c *rpc.Call) {
    var requestVoteResult *raft_rpc.RequestVoteResult

    requestVoteResult = c.Reply.(*raft_rpc.RequestVoteResult)

    if (requestVoteResult.VoteGranted == true) {
        s.logState("Voted for me! +1")
        s.votesReceived++
    }
}

func (s *Server) resetElectionTimeout() {
    go func() {
        rand.Seed(time.Now().Unix())
        randomTimeout := rand.Intn(s.config.ElectionTimeout + 1 - s.config.ElectionTimeout) + s.config.ElectionTimeout

        time.Sleep(time.Duration(randomTimeout) * time.Millisecond)
        s.electionTimeoutChan <- true
    }()
}

func (s *Server) RequestVote(v raft_rpc.RequestVote, voteResult *raft_rpc.RequestVoteResult) error {
    fmt.Printf("Server: %s", s)
    s.logState("Ready to cast vote")

    // already voted, ignore
    if s.State.votedFor != nil {
        s.logState("Already voted, ignoring")
        return nil
    }

    // node is outdated, ignore
    if v.LastLogIndex < s.State.lastApplied {
        s.logState("Node to vote for is outdated, ignoring")
        return nil
    }

    node, err := config.ParseNodeConfig(v.CandidateId)

    if err != nil {
        s.logState("Invalid candidate for voting: " + err.Error())
        return nil
    }

    s.State.votedFor = &node

    voteResult.VoteGranted = true
    voteResult.Term = s.State.currentTerm

    s.logState("Voted for node: " + v.CandidateId)

    return nil
}

func (s *Server) lead() {
    s.logState("Leading...")

    for s.State.CurrentState == Leader {
        select {
        case <-s.exit:
            s.State.CurrentState = Stopped
            return

        case connection := <-s.connections:
            s.accept(connection)
        }
    }
}

func (s *Server) openConnections(listener net.Listener) {
    s.connections = make(chan net.Conn)

    go func() {
        for {
            connection, err := listener.Accept()

            if connection == nil {
                log.Fatal(err)
                continue
            }

            s.logState("Accepted connection from " + connection.RemoteAddr().String())

            s.connections<-connection
        }
    }()
}

func (s *Server) accept(connection net.Conn) {
    go s.rpcServer.ServeConn(connection)
}

func (s *Server) handleSet(key string, value string) string {
    err := s.database.Update(func(tx *bolt.Tx) error {
        bucket, err := tx.CreateBucketIfNotExists([]byte("sets"))
        if err != nil {
            return err
        }

        err = bucket.Put([]byte(key), []byte(value))
        if err != nil {
            return err
        }
        return nil
    })

    if err != nil {
        log.Fatal(err)
    }

    return "(true)"
}

func (s *Server) handleGet(key string) string {
    value := ""

    err := s.database.View(func(tx *bolt.Tx) error {
        bucket := tx.Bucket([]byte("sets"))
        value = string(bucket.Get([]byte(key)))

        return nil
    })

    if err != nil {
        log.Fatal(err)
    }

    return value
}

func (s *Server) initializeDb() {
    db, err := bolt.Open(s.config.DatabasePath, 0600, nil)

    if err != nil {
        log.Fatal(err)
    }

    s.database = db
}

func (s *Server) logState(state string) {
    fmt.Println("> (" + s.State.CurrentState + ") " + state)
}
package server

import (
    "log"
    "net"

    "github.com/albertofem/gloster/config"
    n "github.com/albertofem/gloster/node"
    "github.com/mgutz/str"
    "github.com/boltdb/bolt"
    "fmt"
    "bufio"
    "strings"
    "github.com/albertofem/gloster/database"
    "time"
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
    state       ServerState
    electionTimeoutChan chan bool
}

type ServerState struct {
    currentState string // follower, candidate, leader
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
        currentState: Follower,
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
        state: serverState,
        electionTimeoutChan: make(chan bool),
    }
}

func (s *Server) Init() {
    defer close(s.exit)

    s.initializeDb();
    s.serve()
    s.run()
}

func (s *Server) Shutdown() {
    fmt.Println("Shuting down...")
    s.exit <- true
}

func (s *Server) serve() {
    server, err := net.Listen("tcp", string(s.config.Addr.String()))

    if err != nil {
        log.Println("Error starting Socket Server: ", err)
        return
    }

    s.openConnections(server)
}

func (s *Server) run() { // main server loop

    s.state.currentState = Follower

    currentState := s.state.currentState

    for currentState != Stopped {
        switch currentState {
            case Follower:
                s.follow()
            case Candidate:
                s.candidate()
            case Leader:
                s.lead()
        }

        currentState = s.state.currentState
    }
}

func (s *Server) follow() {

    s.logState("Following...")
    s.resetElectionTimeout()

    for s.state.currentState == Follower {
        select {
        case <-s.exit:
            s.state.currentState = Stopped
            return

        case client := <-s.connections:
            go s.accept(client)

        case <-s.electionTimeoutChan:
            s.state.currentState = Candidate
        }
    }

    // handle request vote (vote for peer)

    // if timeout -> go to candidate
}

func (s *Server) candidate() {
    s.logState("Candidating...")

    s.state.currentTerm = 1
    s.resetElectionTimeout()

    for s.state.currentState == Candidate {
        select {
        case <-s.exit:
            s.state.currentState = Stopped
            return

        case client := <-s.connections:
            go s.accept(client)

        case <-s.electionTimeoutChan:
            s.state.currentState = Follower
        }
    }

    // increment current term

    // vote for self

    // reset election timeout

    // send request vote to all peers

    // if requestvote -> ignore

    // if appendentries -> to follower

    // if timeout -> new election
}

func (s *Server) resetElectionTimeout() {
    s.logState("Election timeout reset")

    go func() {
        time.Sleep(time.Duration(s.config.ElectionTimeout) * time.Millisecond)
        s.electionTimeoutChan <- true
    }()
}

func (s *Server) lead() {
    // send empty append entries to all nodes (confirm as leader)

    // append entries heartbeat (define timeout)

    //
}

func (s *Server) State() string {
    return s.state.currentState
}

func (s *Server) openConnections(listener net.Listener) {
    s.connections = make(chan net.Conn)

    go func() {
        for {
            client, err := listener.Accept()

            if client == nil {
                log.Fatal(err)
                continue
            }

            fmt.Println("Accepted connection from " + client.RemoteAddr().String())

            s.connections<-client
        }
    }()
}

func (s *Server) accept(client net.Conn) {
    b := bufio.NewReader(client)

    for {
        line, err := b.ReadString('\n')

        if err != nil {
            break
        }

        command := str.ToArgv(strings.Trim(line, "\n"))
        response := ""

        switch command[0] {
            case "SET", "set":
                response = s.handleSet(command[1], command[2])
            case "GET", "get":
                response = s.handleGet(command[1])
            default:
                response = "Unrecognized command"
        }

        client.Write([]byte(response + "\n"))
    }
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
    db, err := bolt.Open("my.db", 0600, nil)

    if err != nil {
        log.Fatal(err)
    }

    s.database = db
}

func (s *Server) logState(state string) {
    fmt.Println("> - " + s.State() + " - " + state)
}
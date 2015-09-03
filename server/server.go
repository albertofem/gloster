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
)

type Server struct {
    config      *config.Config
    node        n.Node
    database    *bolt.DB
    exit        chan bool
    connections chan net.Conn
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

func (s *Server) run() {
    for {
        select {
        case <-s.exit:
            return

        case client := <-s.connections:
            go s.accept(client)
        }
    }
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
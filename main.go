package main

import (
    "flag"
    "log"
    "fmt"
    "runtime"
    "os"
    "os/signal"
    "syscall"

    "github.com/albertofem/gloster/config"
    "github.com/albertofem/gloster/server"
    "gopkg.in/alecthomas/kingpin.v1"
)

var (
    addr = kingpin.Arg("addr", "Node address <ip>:<port>").Required().String()
    cluster = kingpin.Arg("cluster", "Rest of cluster nodes, comma separated").Required().String()
    database = kingpin.Arg("database", "Path to database file").Required().String()
)


func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    kingpin.Parse();

    flag.Parse()

    log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

    config := config.NewConfig(*addr, *cluster, *database)

    fmt.Printf("Welcome to Gloster!, Listening on %d\n", config.Addr.Port)
    fmt.Printf("Using %d nodes\n", len(config.Cluster))

    for _,node := range config.Cluster {
        fmt.Printf("Node %s:%d\n", node.Host, node.Port)
    }

    server := server.New(config)
    c := make(chan os.Signal, 1)

    signal.Notify(
        c,
        os.Kill,
        os.Interrupt,
        syscall.SIGHUP,
        syscall.SIGINT,
        syscall.SIGTERM,
        syscall.SIGQUIT)

    go func() {
        <-c
        server.Shutdown()
    }()

    server.Init()
}

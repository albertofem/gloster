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
)

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    addr := flag.String("addr", "127.0.0.1:12000", "Port where Gloster will listen on")
    cluster := flag.String("cluster", "127.0.0.1:12000,127.0.0.1:12001,127.0.0.1:12002", "Cluster list, comma separated")

    flag.Parse()

    log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

    config := config.NewConfig(*addr, *cluster)

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

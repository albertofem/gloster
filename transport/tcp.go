package transport

import (
    "net"
    "log"
    "net/rpc"

    "github.com/albertofem/gloster/rpcm"
)

func ServeTcp() {
    glosterServer := new(rpcm.GlosterServer)
    glosterServer.SetServer(server)

    server.rpcServer = rpc.NewServer()
    server.rpcServer.Register(glosterServer)

    listener, err := net.Listen("tcp", string(server.config.Addr.String()))

    if err != nil {
        return
    }

    openConnections(listener)
}

func openConnections(listener net.Listener) {
    connectionsChan := make(chan net.Conn)

    go func() {
        for {
            connection, err := listener.Accept()

            if connection == nil {
                log.Fatal(err)
                continue
            }

            connectionsChan <-connection
        }
    }()

    return connectionsChan
}
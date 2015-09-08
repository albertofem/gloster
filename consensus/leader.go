package consensus

import (
    "net"

    "github.com/albertofem/gloster/server"
)

func Lead(
    server *server.Server,
    exitChan chan bool,
    connectionsChan chan net.Conn,
) {
    for server.IsLeader() {
        select {
        case <-exitChan:
            server.Stop()
            return

        case connection := <-connectionsChan:
            server.Accept(connectionsChan)
        }
    }
}
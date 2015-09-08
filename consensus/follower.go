package consensus

import (
    "net"

    "github.com/albertofem/gloster/server"
)

func Follow(
    server* server.Server,
    exitChan chan bool,
    electionTimeoutChan chan bool,
    connectionsChan chan net.Conn,
) {
    server.ResetElectionTimeout()

    for server.IsFollower() {
        select {
        case <-exitChan:
            server.Stop()
            return

        case connection := <-connectionsChan:
            server.Accept(connection)

        case <-electionTimeoutChan:
            onFollowerElectionTimeout(server)
        }
    }
}

func onFollowerElectionTimeout(server *server.Server) {

}
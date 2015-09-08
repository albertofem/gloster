package consensus

import (
    "net"

    "github.com/albertofem/gloster/server"
    "github.com/albertofem/gloster/rpcm"
)

func Candidate(
    server *server.Server,
    exitChan chan bool,
    electionTimeoutChan chan bool,
    connectionsChan chan net.Conn,
) {

    initializeCandidate(server)

    // send request vote to all peers
    requestVoteResultChan := rpcm.SendRequestVote(server)

    for server.IsCandidate() {
        select {
        case <-exitChan:
            server.Stop()
            return

        case connection := <-connectionsChan:
            server.Accept(connection)

        case <-electionTimeoutChan:
            onCandidateElectionTimeout(server)
            return

        case requestResult := <-requestVoteResultChan:
            rpcm.HandleSendRequestVoteResult(server, requestResult)
        }
    }
}

func onCandidateElectionTimeout(server *server.Server) {
    server.ResetElection()
}

func initializeCandidate(server *server.Server) {
    server.ResetElectionTimeout()
    server.SetTerm(1)
    server.VoteForSelf()
}
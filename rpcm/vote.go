package rpcm

import (
    "net/rpc"

    "github.com/albertofem/gloster/server"
    "github.com/albertofem/gloster/config"
)

type RequestVote struct {
    Term            int
    CandidateId     string
    LastLogIndex    int
    LastLogTerm     int
}

type RequestVoteResult struct {
    Term            int
    VoteGranted     bool
}

func HandleRequestVote(
    server *server.Server,
    requestVote RequestVote,
    requestVoteResult *RequestVoteResult,
) error {
    if server.HasVoted() != nil {
        return nil
    }

    node, err := config.ParseNodeConfig(requestVote.CandidateId)

    if err != nil {
        return nil
    }

    server.VoteFor(node)

    requestVoteResult.VoteGranted = true
    requestVoteResult.Term = server.CurrentTerm()

    return nil
}


func SendRequestVote(server *server.Server) (requestVoteResultChan chan *rpc.Call) {
    for _, node := range server.Cluster() {
        client, err := rpc.Dial("tcp", node.String())

        if err != nil {
            server.LogState("Cannot connect to peer: " + node.String())
            continue
        }

        requestVote := RequestVote{
            Term: server.CurrentTerm(),
            CandidateId: server.Node().String(),
        }

        var requestVoteResult RequestVoteResult

        client.Go("GlosterServer.RequestVote", requestVote, &requestVoteResult, requestVoteResultChan)
    }

    return requestVoteResultChan
}

func HandleSendRequestVoteResult(server *server.Server, call *rpc.Call) {
    var requestVoteResult *RequestVoteResult = call.Reply.(*RequestVoteResult)

    if (requestVoteResult.VoteGranted == true) {
        server.IncrementVote()
    }
}
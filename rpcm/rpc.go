package rpcm

import (
    "github.com/albertofem/gloster/server"
)

type GlosterServer struct {
    server *server.Server
}

func (rpcServer *GlosterServer) SetServer(server *server.Server) {
    rpcServer.server = server
}

func (rpcServer *GlosterServer) RequestVote(
    requestVote RequestVote,
    requestVoteResult *RequestVoteResult,
) {
    return HandleRequestVote(rpcServer.server, requestVote, requestVoteResult)
}
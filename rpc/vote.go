package rpc

type RequestVote struct {
    term int
    candidateId string
    lastLogIndex int
    lastLogTerm int
}

type RequestVoteResult struct {
    term int
    voteGrande bool
}



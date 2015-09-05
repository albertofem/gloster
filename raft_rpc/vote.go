package raft_rpc

type RequestVote struct {
    Term int
    CandidateId string
    LastLogIndex int
    LastLogTerm int
}

type RequestVoteResult struct {
    Term int
    VoteGranted bool
}
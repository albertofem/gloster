package raft_rpc
import "github.com/albertofem/gloster/database"

type AppendEntries struct {
    Term int
    LeaderId string
    PrevLogIndex int
    PrevLogTerm int
    Entries []database.Command
    LeaderCommit int
}

type AppendEntriesResult struct {
    Term int
    Success bool
}
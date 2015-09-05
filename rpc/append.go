package rpc
import "github.com/albertofem/gloster/database"

type AppendEntries struct {
    term int
    leaderId string
    prevLogIndex int
    prevLogTerm int
    entries []database.Command
    leaderCommit int
}

type AppendEntriesResult struct {
    term int
    success bool
}
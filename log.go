package lilraft

import "github.com/lilwulin/lilraft/protobuf"

// Log is used as a way to ensure consensus
// TODO: fill this
type Log struct {
	entries []*LogEntry
}

func newLog() *Log {
	return &Log{
		entries: make(*LogEntry, 0),
	}
}

func (log *Log) lastLogIndex() uint32 {
	return len(log.entries) - 1
}

// LogEntry is the entry in log, it wraps the LogEntry in raftpb.proto
type LogEntry struct {
	*protobuf.LogEntry
}

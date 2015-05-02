package lilraft

import (
	"bytes"
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"github.com/lilwulin/lilraft/protobuf"
)

// Log is used as a way to ensure consensus
// TODO: fill this
type Log struct {
	entries   []*LogEntry
	lastIndex uint64
}

func newLog() *Log {
	return &Log{
		entries: make([]*LogEntry, 0),
	}
}

// TODO: just add these temporarily
func (log *Log) lastIndexInc() {
	log.lastIndex++
}

func (log *Log) lastLogIndex() uint64 {
	return log.lastIndex
}

// LogEntry is the entry in log, it wraps the LogEntry in raftpb.proto
type LogEntry struct {
	*protobuf.LogEntry
}

func newLogEntry(index uint64, term uint64, command Command) (*LogEntry, error) {
	pbEntry := protobuf.LogEntry{
		Index:       proto.Uint64(index),
		Term:        proto.Uint64(term),
		CommandName: proto.String(command.Name()),
	}
	var bytesBuffer bytes.Buffer
	if err := json.NewEncoder(&bytesBuffer).Encode(command); err != nil {
		return nil, err
	}
	return &LogEntry{&pbEntry}, nil
}

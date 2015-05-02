package lilraft

import (
	"bytes"
	"encoding/json"

	"github.com/golang/protobuf/proto"
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

func (log *Log) appendEntry(logEtry *LogEntry) {
	log.lastIndexInc()
	log.entries = append(log.entries, logEtry)
}

// LogEntry is the entry in log, it wraps the LogEntry in raftpb.proto

func (log *Log) newLogEntry(term uint64, command Command) (*LogEntry, error) {
	// TODO: add more arguments later
	var bytesBuffer bytes.Buffer
	if err := json.NewEncoder(&bytesBuffer).Encode(command); err != nil {
		return nil, err
	}
	pbEntry := &LogEntry{
		Index:       proto.Uint64(log.lastLogIndex() + 1),
		Term:        proto.Uint64(term),
		CommandName: proto.String(command.Name()),
		Command:     bytesBuffer.Bytes(),
	}
	return pbEntry, nil
}

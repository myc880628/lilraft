package lilraft

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/golang/protobuf/proto"
)

// Log is used as a way to ensure consensus
// TODO: fill this, add mutex log maybe.
type Log struct {
	entries     []*LogEntry
	commitIndex uint64
	sync.RWMutex
}

func newLog() *Log {
	return &Log{
		entries: make([]*LogEntry, 0),
	}
}

// TODO: fill this func
func (log *Log) setCommitIndex(i uint64) {

}

func (log *Log) lastLogIndex() uint64 {
	log.RLock()
	defer log.RUnlock()
	if len(log.entries) > 0 {
		return *log.entries[len(log.entries)-1].Index
	} else {
		return 0
	}
}

func (log *Log) lastLogTerm() uint64 {
	log.RLock()
	defer log.RUnlock()
	if len(log.entries) > 0 {
		return *log.entries[len(log.entries)-1].Term
	} else {
		return 0
	}
}

func (log *Log) prevLogTerm(nextIndex uint64) uint64 {
	log.RLock()
	defer log.RUnlock()
	return *log.entries[nextIndex-1].Term
}

func (log *Log) prevLogIndex(nextIndex uint64) uint64 {
	log.RLock()
	defer log.RUnlock()
	return *log.entries[nextIndex-1].Index
}

func (log *Log) entriesAfer(index uint64) (le []*LogEntry) {
	log.RLock()
	defer log.RUnlock()
	le = log.entries[index:]
	return
}

func (log *Log) appendEntry(logEtry *LogEntry) {
	log.Lock()
	defer log.Unlock()
	log.entries = append(log.entries, logEtry)
}

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

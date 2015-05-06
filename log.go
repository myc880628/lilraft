package lilraft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
)

var errLogTruncEmpty = fmt.Errorf("truncate empty log")

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

// TODO: add more procedure
func (log *Log) setCommitIndex(i uint64) {
	log.Lock()
	log.commitIndex = i
	log.Unlock()
}

func (log *Log) getCommitIndex() uint64 {
	log.RLock()
	defer log.RUnlock()
	return log.commitIndex
}

// first index in log entries
func (log *Log) startIndex() uint64 {
	if len(log.entries) > 0 {
		return log.entries[0].GetIndex()
	}
	return 0
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
	return *log.entries[nextIndex-1-log.startIndex()].Term
}

func (log *Log) prevLogIndex(nextIndex uint64) uint64 {
	log.RLock()
	defer log.RUnlock()
	return *log.entries[nextIndex-1-log.startIndex()].Index
}

func (log *Log) entriesAfer(index uint64) (le []*LogEntry) {
	log.RLock()
	defer log.RUnlock()
	if index < log.startIndex() {
		le = nil
	} else {
		le = log.entries[(index - log.startIndex()):]
	}
	return
}

func (log *Log) appendEntry(logEntry *LogEntry) {
	log.Lock()
	defer log.Unlock()
	log.entries = append(log.entries, logEntry)
}

func (log *Log) appendEntries(index uint64, logEntries []*LogEntry) {
	startIndex := log.startIndex()
	log.Lock()
	defer log.Unlock()
	log.entries = append(log.entries[:(index-startIndex)+1], logEntries...)
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

func (log *Log) contains(index uint64, term uint64) bool {
	log.RLock()
	defer log.RUnlock()
	if log.lastLogIndex() < index {
		return false
	}
	if log.entries[index-log.startIndex()].GetTerm() == term {
		return true
	}
	return false
}

// func (log *Log) truncate(index uint64) error {
// 	length := len(log.entries)
// 	for log.lastLogIndex() >= index {
// 		if length > 0 {
// 			log.entries = log.entries[:length-1]
// 		} else {
// 			return errLogTruncEmpty
// 		}
// 	}
// 	return nil
// }

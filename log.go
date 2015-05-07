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
	commitIndex int64
	sync.RWMutex
}

func newLog() (l *Log) {
	l = &Log{
		entries:     make([]*LogEntry, 0),
		commitIndex: -1,
	}
	return
}

// TODO: add more procedure
func (log *Log) setCommitIndex(commitIndex int64) error {
	log.Lock()
	defer log.Unlock()
	for i := log.commitIndex + 1; i <= commitIndex; i++ {
		entryIndex := i - log.startIndex()
		entry := log.entries[entryIndex]
		command, err := newCommand(entry.GetCommandName(), entry.GetCommand())
		if err != nil {
			return err
		}
		command.Apply()
	}
	log.commitIndex = commitIndex
	return nil
}

// first index in log entries
func (log *Log) startIndex() int64 {
	if len(log.entries) > 0 {
		return log.entries[0].GetIndex()
	}
	return 0
}

func (log *Log) lastLogIndex() int64 {
	log.RLock()
	defer log.RUnlock()
	if len(log.entries) > 0 {
		return *log.entries[len(log.entries)-1].Index
	}
	return 0
}

func (log *Log) lastLogTerm() int64 {
	log.RLock()
	defer log.RUnlock()
	if len(log.entries) > 0 {
		return *log.entries[len(log.entries)-1].Term
	} else {
		return 0
	}
}

func (log *Log) prevLogTerm(nextIndex int64) int64 {
	log.RLock()
	defer log.RUnlock()
	return *log.entries[nextIndex-1-log.startIndex()].Term
}

func (log *Log) prevLogIndex(nextIndex int64) int64 {
	log.RLock()
	defer log.RUnlock()
	return *log.entries[nextIndex-1-log.startIndex()].Index
}

func (log *Log) entriesAfer(index int64) (le []*LogEntry) {
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

func (log *Log) appendEntries(index int64, logEntries []*LogEntry) {
	startIndex := log.startIndex()
	log.Lock()
	defer log.Unlock()
	log.entries = append(log.entries[:(index-startIndex)+1], logEntries...)
}

func (log *Log) newLogEntry(term int64, command Command) (*LogEntry, error) {
	// TODO: add more arguments later
	var bytesBuffer bytes.Buffer
	if err := json.NewEncoder(&bytesBuffer).Encode(command); err != nil {
		return nil, err
	}
	pbEntry := &LogEntry{
		Index:       proto.Int64(log.lastLogIndex() + 1),
		Term:        proto.Int64(term),
		CommandName: proto.String(command.Name()),
		Command:     bytesBuffer.Bytes(),
	}
	return pbEntry, nil
}

func (log *Log) contains(index int64, term int64) bool {
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

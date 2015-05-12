package lilraft

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
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
func (log *Log) setCommitIndex(commitIndex int64, context interface{}) error {
	log.Lock()
	defer log.Unlock()
	startIndex := log.startIndex()
	if startIndex < 0 {
		return fmt.Errorf("set commit index on empty log")
	}
	for i := log.commitIndex + 1; i <= commitIndex; i++ {
		entryIndex := i - startIndex
		entry := log.entries[entryIndex]
		if entry.GetCommandName() == cOldNewStr || entry.GetCommandName() == cNewStr {
			continue
		}
		command, err := newCommand(entry.GetCommandName(), entry.GetCommand())
		if err != nil {
			return err
		}
		command.Apply(context)
	}
	log.commitIndex = commitIndex
	return nil
}

// first index in log entries
func (log *Log) startIndex() int64 {
	if len(log.entries) > 0 {
		return log.entries[0].GetIndex()
	}
	return -1
}

func (log *Log) lastLogIndex() int64 {
	log.RLock()
	defer log.RUnlock()
	if len(log.entries) > 0 {
		return *log.entries[len(log.entries)-1].Index
	}
	return -1
}

func (log *Log) lastLogTerm() int64 {
	log.RLock()
	defer log.RUnlock()
	if len(log.entries) > 0 {
		return *log.entries[len(log.entries)-1].Term
	}
	return -1
}

// TODO: use this to prevent replicated commands
func (log *Log) lastLogEntry() *LogEntry {
	log.RLock()
	defer log.RUnlock()
	if len(log.entries) == 0 {
		return nil
	}
	return log.entries[len(log.entries)-1]
}

func (log *Log) prevLogTerm(index int64) int64 {
	log.RLock()
	defer log.RUnlock()
	startIndex := log.startIndex()
	if startIndex < 0 || index <= startIndex {
		return -1
	}
	return log.entries[index-startIndex-1].GetTerm()
}

func (log *Log) prevLogIndex(index int64) int64 {
	log.RLock()
	defer log.RUnlock()
	startIndex := log.startIndex()
	if startIndex < 0 || index <= startIndex {
		return -1
	}
	return log.entries[index-startIndex-1].GetIndex()
}

func (log *Log) entriesAfer(index int64) []*LogEntry {
	log.RLock()
	defer log.RUnlock()
	startIndex := log.startIndex()
	if index < startIndex {
		return nil
	}
	if index-startIndex >= int64(len(log.entries)) {
		return nil
	}
	return log.entries[index-startIndex:]
}

func (log *Log) appendEntry(logEntry *LogEntry) {
	log.Lock()
	defer log.Unlock()
	log.entries = append(log.entries, logEntry)
}

func (log *Log) appendEntries(s *Server, prevIndex int64, logEntries []*LogEntry) {
	log.Lock()
	defer log.Unlock()
	if len(log.entries) == 0 {
		log.entries = logEntries
		return
	}
	log.entries = log.entries[:prevIndex-log.startIndex()+1]
	// log.entries = append(log.entries[:(index-log.startIndex())+1], logEntries...)
	for _, entry := range logEntries {
		if entry.GetCommandName() == cOldNewStr {
			nodes := []Node{}
			if err := gob.NewDecoder(bytes.NewReader(entry.GetCommand())).Decode(&nodes); err != nil {
				logger.Println("decode Cold,new config err: ", err.Error())
			}
			s.config.setState(c_old_new)
			s.config.c_NewNode = make(nodeMap)
			for _, node := range nodes {
				s.config.c_NewNode[node.id()] = node
			}
		} else if entry.GetCommandName() == cNewStr {
			nodes := []Node{}
			if err := gob.NewDecoder(bytes.NewReader(entry.GetCommand())).Decode(&nodes); err != nil {
				logger.Println("decode Cnew config err: ", err.Error())
			}
			s.config.setState(c_old)
			s.config.c_OldNode = makeNodeMap(nodes...)
			s.config.c_NewNode = nil
		}
		log.entries = append(log.entries, entry)
	}
}

func (log *Log) newLogEntry(term int64, command Command) (*LogEntry, error) {
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

func (log *Log) newConfigEntry(term int64, nodes []Node, state string) (*LogEntry, error) {
	var bytesBuffer bytes.Buffer
	if err := gob.NewEncoder(&bytesBuffer).Encode(nodes); err != nil {
		return nil, err
	}
	pbEntry := &LogEntry{
		Index:       proto.Int64(log.lastLogIndex() + 1),
		Term:        proto.Int64(term),
		CommandName: proto.String(state),
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
	if index < 0 { // if index<0, means leader's request is empty
		return true
	}
	if len(log.entries) < 0 { // empty log, impossible to contain index
		return false
	}
	if log.entries[index-log.startIndex()].GetTerm() == term {
		return true
	}
	return false
}

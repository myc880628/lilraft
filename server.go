package lilraft

import (
	"math/rand"
	"sync"
	"time"
)

// state constant
const (
	follower = iota
	candidate
	leader
	stopped
)

var (
	minTimeout int32 = 150            // minimum timeout: 150 ms
	maxTimeout       = minTimeout * 2 //maximum timeout: 300ms
)

// mutexState wraps stateInt with a Mutex
type mutexState struct {
	stateInt uint8
	sync.RWMutex
}

func (mstate *mutexState) getState() uint8 {
	mstate.RLock()
	defer mstate.RUnlock()
	return mstate.stateInt
}

func (mstate *mutexState) setState(stateInt uint8) {
	mstate.Lock()
	mstate.stateInt = stateInt
	mstate.Unlock()
}

// Server is a concrete machine that process command, appendentries and requestvote, etc.
type Server struct {
	id uint16
	*mutexState
	context      interface{}
	electionTick <-chan time.Time
}

func randeomElectionTimeout() time.Duration {
	d := minTimeout + rand.Intn(maxTimeout-minTimeout)
	return time.Duration(d) * time.Millisecond
}

func resetElectionTimeout(s *Server) {
	s.electionTick = time.NewTicker(randeomElectionTimeout()).C
}

// NewServer can return a new server for clients
func NewServer(id uint16, context interface{}) (s *Server) {
	if id == nil {
		panic("lilraft: id is required")
	}
	if context == nil {
		panic("lilraft: contex is required")
	}
	s := &Server{
		id:      id,
		context: context,
	}
	rand.Seed(time.Now().Unix()) // random seed for election timeout
	resetElectionTimeout(s)
	return
}

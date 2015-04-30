package lilraft

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/lilwulin/lilraft/protobuf"
)

// state constant
const (
	follower = iota
	candidate
	leader
	stopped
)

const noleader = 0

var (
	minTimeout = 150            // minimum timeout: 150 ms
	maxTimeout = minTimeout * 2 //maximum timeout: 300ms
	// electionTimeout
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
	id              uint32
	leader          uint32
	currentTerm     uint32
	context         interface{}
	electionTimeout *time.Timer
	nodemap         nodeMap
	log             *Log
	httpClient      http.Client
	votes           uint32
	votefor         uint32
	voteCountChan   chan uint32
	config          *configuration
	*mutexState
}

func randomElectionTimeout() time.Duration {
	d := minTimeout + rand.Intn(maxTimeout-minTimeout)
	return time.Duration(d) * time.Millisecond
}

func (s *Server) resetElectionTimeout() {
	s.electionTimeout = time.NewTimer(randomElectionTimeout())
}

// NewServer can return a new server for clients
func NewServer(id uint32, context interface{}) (s *Server) {
	if context == nil {
		panic("lilraft: contex is required")
	}
	if id == 0 { // HINT: id可以在之后作为投票的标记
		panic("id must be > 0")
	}
	s = &Server{
		id:      id,
		leader:  noleader,
		context: context,
		// mutexState:  follower,
		nodemap:     make(nodeMap),
		currentTerm: 0,
		log:         newLog(),
		votes:       0,
		votefor:     0,
		// voteCountChan: make(chan uint32, 5),
	}
	s.mutexState.setState(follower)
	// http.Client can cache TCP connections
	s.httpClient.Transport = &http.Transport{DisableKeepAlives: false}
	rand.Seed(time.Now().Unix()) // random seed for election timeout
	s.resetElectionTimeout()
	return
}

// Start starts a server, remember to call NewServer before call this.
func (s *Server) Start() {
	go s.loop()
}

func (s *Server) loop() {
	for {
		switch s.getState() {
		case follower:
			s.followerloop()
		case candidate:
			s.candidateloop()
		case leader:
			s.leaderloop()
		case stopped:
			break
		default:
			panic("lilraft: impossible state")
		}
	}
}

// TODO: fill this.
func (s *Server) followerloop() {
	for s.getState() == follower {
		select {
		case <-s.electionTimeout.C:
			s.setState(candidate)
			return
			// case <-
		}
	}
}

// TODO: fill this.
func (s *Server) candidateloop() {
	s.leader = noleader
	s.currentTerm++ // enter candidate state. Server increments its term
	s.resetElectionTimeout()
	s.votes = 1 // candidate vote for himself
	s.requestVotes()
	for s.getState() == candidate {
		select {
		case <-s.electionTimeout.C:
			// Candidate reset timeout and start a new election
			s.resetElectionTimeout()
			s.votes = 1 // candidate vote for himself
			s.requestVotes()
		case <-s.voteCountChan:
			s.votes += 1
		}
	}
}

// TODO: fill this.
func (s *Server) leaderloop() {

}

func (s *Server) requestVotes() {
	// if re-relect, reset the voteCountChan
	s.voteCountChan = make(chan uint32, 5)
	for _, node := range s.nodemap {
		go func() {
			pb := &protobuf.RequestVoteRequest{
				CandidateID:  proto.Uint32(s.id),
				Term:         proto.Uint32(s.currentTerm),
				LastLogIndex: proto.Uint64(s.log.lastLogIndex()),
			}
			responseProto, err := (*node).rpcRequestVote(s, pb)
			if err != nil {
				fmt.Println(err.Error())
			}
			if responseProto.GetVoteGranted() {
				s.voteCountChan <- 1
			}
		}()
	}
}

package lilraft

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
)

// state constant
const (
	follower = iota
	candidate
	leader
	stopped
)

const noleader = 0
const notVotedYet = 0

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

// vote is passed through the server's voteChan
type vote struct {
	id      uint32
	granted bool
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
	//----from the paper------
	currentTerm      uint64
	votefor          uint32
	log              *Log
	commitIndex      uint64
	lastAppliedIndex uint64
	nextIndex        []uint64
	matchIndex       []uint64
	//------------------------
	id               uint32
	leader           uint32
	context          interface{}
	electionTimeout  *time.Timer
	nodemap          nodeMap
	httpClient       http.Client
	voteChan         chan *vote
	config           *configuration
	nodesVoteGranted map[uint32]bool
	commandChan      chan wrappedCommand
	mutexState
}

type wrappedCommand struct {
	command Command
	errChan chan error
}

func randomElectionTimeout() time.Duration {
	d := minTimeout + rand.Intn(maxTimeout-minTimeout)
	return time.Duration(d) * time.Millisecond
}

func (s *Server) resetElectionTimeout() {
	s.electionTimeout = time.NewTimer(randomElectionTimeout())
}

// NewServer can return a new server for clients
func NewServer(id uint32, context interface{}, config *configuration) (s *Server) {
	if context == nil {
		panic("lilraft: contex is required")
	}
	if id == 0 { // HINT: id可以在之后作为投票的标记
		panic("lilraft: id must be > 0")
	}
	if config == nil {
		panic("lilraft: configuration unset")
	}
	s = &Server{
		id:      id,
		leader:  noleader,
		context: context,
		// mutexState:  follower,
		nodemap:     make(nodeMap),
		currentTerm: 0,
		log:         newLog(),
		votefor:     0,
		// voteCountChan: make(chan uint32, 5),
		commandChan: make(chan wrappedCommand),
	}
	s.mutexState.setState(stopped)
	// http.Client can cache TCP connections
	s.httpClient.Transport = &http.Transport{DisableKeepAlives: false}
	s.httpClient.Timeout = time.Duration(minTimeout/2) * time.Millisecond
	rand.Seed(time.Now().Unix()) // random seed for election timeout
	s.resetElectionTimeout()
	return
}

// Start starts a server, remember to call NewServer before call this.
func (s *Server) Start() {
	s.setState(follower)
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
	s.nodesVoteGranted = make(map[uint32]bool)
	s.voteForItself()
	s.requestVotes()
	if s.electionPass() {
		// TODO: Add more operation
		s.setState(leader)
		s.votefor = notVotedYet
		return
	}
	for s.getState() == candidate {
		select {
		case <-s.electionTimeout.C:
			// Candidate start a new election
			return
		}
	}
}

// TODO: fill this.
func (s *Server) leaderloop() {
	for s.getState() == leader {
		select {
		case wrappedCommand := <-s.commandChan:
			newEntry, err := s.log.newLogEntry(s.currentTerm, wrappedCommand.command)
			if err != nil {
				wrappedCommand.errChan <- err
			}
			s.log.appendEntry(newEntry)

		}

	}
}

// Exec executes client's command. client should
// use concrete commands that implement Command interface
func (s *Server) Exec(command Command) error {
	errChan := make(chan error)
	s.commandChan <- wrappedCommand{
		errChan: errChan,
		command: command,
	}
	return <-errChan
}

// RegisterCommand registers client's commands that
// implement Command interface. Only registered commands
// can be decoded and executed by other nodes.
// TODO: fill this
func (s *Server) RegisterCommand(command Command) {
	if command == nil {
		panic("lilraft: Command cannot be nil")
	} else if commandType[command.Name()] != nil {
		panic("lilraft: Command exists!")
	}
	commandType[command.Name()] = command
}

func (s *Server) voteForItself() {
	s.votefor = s.id
	s.nodesVoteGranted[s.id] = true
}

func (s *Server) requestVotes() {
	// if re-relect, reset the voteChan
	s.voteChan = make(chan *vote, 10)
	allNodes := s.config.allNodes()
	for id, node := range allNodes {
		responded := false
		go func() { // send vote request simultaneously
			for !responded { // not responded, keep trying
				if id == s.id { // if it's the candidiate itself
					return
				}
				pb := &RequestVoteRequest{
					CandidateID:  proto.Uint32(s.id),
					Term:         proto.Uint64(s.currentTerm),
					LastLogIndex: proto.Uint64(s.log.lastLogIndex()),
				}
				responseProto, err := (*node).rpcRequestVote(s, pb)
				if err != nil {
					fmt.Println(err.Error())
					responded = false
					// does not timeout and try to send request over and over agian
					continue
				}
				responded = true
				s.voteChan <- &vote{
					id:      id,
					granted: responseProto.GetVoteGranted(),
				}
			}
		}()
	}

	for {
		select {
		case vote := <-s.voteChan:
			s.nodesVoteGranted[vote.id] = vote.granted
			if len(s.nodesVoteGranted) == len(allNodes) { // if all nodes have voted
				return
			}
		}
	}
}

func (s *Server) electionPass() bool {
	if s.config.getState() == c_old {
		votesCount := 0
		for _, granted := range s.nodesVoteGranted {
			if granted {
				votesCount++
			}
		}
		if votesCount >= len(s.nodesVoteGranted)/2+1 {
			return true
		}
	} else if s.config.getState() == c_old_new { // candidate must be approved by Cold and Cold,new.
		voteCountOld := 0
		voteCountOldNew := 0
		for id := range s.config.c_OldNode {
			if s.nodesVoteGranted[id] == true {
				voteCountOld++
			}
		}
		for id := range s.config.c_NewNode {
			if s.nodesVoteGranted[id] == true {
				voteCountOldNew++
			}
		}
		if voteCountOld >= len(s.config.c_OldNode)/2+1 &&
			voteCountOldNew >= len(s.config.c_NewNode)/2+1 {
			return true
		}
	} else {
		panic("lilraft: config state unknown")
	}
	return false
}

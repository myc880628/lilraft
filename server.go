package lilraft

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
)

// Logger
var logger = log.New(os.Stdout, "[lilraft]", log.Lmicroseconds)

// Error
var deponseError = fmt.Errorf("leader got deposed")

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

type nextIndex struct {
	m map[uint32]uint64
	sync.RWMutex
}

func (ni *nextIndex) get(id uint32) uint64 {
	ni.RLock()
	defer ni.RUnlock()
	if index, ok := ni.m[id]; !ok {
		panic("lilraft: wrong peer in []NextIndex")
	} else {
		return index
	}
}

func (ni *nextIndex) set(id uint32, index uint64) {
	ni.Lock()
	ni.m[id] = index
	ni.Unlock()
}

func (ni *nextIndex) inc(id uint32) {
	ni.Lock()
	ni.m[id]++
	ni.Unlock()
}

func (ni *nextIndex) dec(id uint32) {
	ni.Lock()
	if ni.m[id] > 0 {
		ni.m[id]--
	}
	ni.Unlock()
}

// Server is a concrete machine that process command, appendentries and requestvote, etc.
type Server struct {
	//----from the paper------
	votefor          uint32
	currentTerm      uint64
	commitIndex      uint64
	lastAppliedIndex uint64
	matchIndex       []uint64
	nextIndex        nextIndex
	log              *Log
	//------------------------
	id                   uint32
	leader               uint32
	context              interface{}
	electionTimeout      *time.Timer
	nodemap              nodeMap
	httpClient           http.Client
	config               *configuration
	getVoteRequestChan   chan wrappedVoteRequest
	getAppendEntriesChan chan wrappedAppendRequest
	// responseChan
	//-------leader only------------
	commandChan chan wrappedCommand
	//-----candidate only-----------
	getVoteResponseChan chan wrappedVoteResponse
	nodesVoteGranted    map[uint32]bool
	//-----------------------------
	mutexState
}

type wrappedVoteResponse struct {
	id uint32
	*RequestVoteResponse
}

type wrappedCommand struct {
	command Command
	errChan chan error
}

type wrappedAppendRequest struct {
	request      *AppendEntriesRequest
	responseChan chan *AppendEntriesResponse
}

type wrappedVoteRequest struct {
	request      *RequestVoteRequest
	responseChan chan *RequestVoteResponse
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
		id:                   id,
		leader:               noleader,
		context:              context,
		currentTerm:          0,
		log:                  newLog(),
		votefor:              0,
		config:               config,
		nodemap:              make(nodeMap),
		commandChan:          make(chan wrappedCommand),
		getVoteRequestChan:   make(chan wrappedVoteRequest),
		getAppendEntriesChan: make(chan wrappedAppendRequest),
	}
	s.mutexState.setState(stopped)
	// http.Client can cache TCP connections
	s.httpClient.Transport = &http.Transport{DisableKeepAlives: false}
	s.httpClient.Timeout = time.Duration(minTimeout/2) * time.Millisecond
	time.Sleep(1 * time.Second)
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
			logger.Printf("node %d become candidate\n", s.id)
			return
		case wrappedVoteRequest := <-s.getVoteRequestChan:
			voteRequest := wrappedVoteRequest.request
			resp := &RequestVoteResponse{
				Term:        proto.Uint64(s.currentTerm),
				VoteGranted: proto.Bool(false),
			}
			if s.votefor != notVotedYet {
				resp.VoteGranted = proto.Bool(false)
			} else if s.currentTerm > voteRequest.GetTerm() {
				resp.VoteGranted = proto.Bool(false)
			} else if voteRequest.GetLastLogTerm() > s.log.lastLogTerm() {
				s.votefor = voteRequest.GetCandidateID()
				resp.VoteGranted = proto.Bool(true)
			} else if voteRequest.GetLastLogTerm() == s.log.lastLogTerm() &&
				voteRequest.GetLastLogIndex() >= s.log.lastLogIndex() {
				s.votefor = voteRequest.GetCandidateID() // 更新term的时候再将votefor更新为notVotedYet
				resp.VoteGranted = proto.Bool(true)
			}
			wrappedVoteRequest.responseChan <- resp
		case wrappedAppendRequest := <-s.getAppendEntriesChan:
			appendRequest := wrappedAppendRequest.request
			resp := s.handleAppendEntriesRequest(appendRequest)
			wrappedAppendRequest.responseChan <- resp
		}
	}
}

func (s *Server) handleAppendEntriesRequest(appendRequest *AppendEntriesRequest) *AppendEntriesResponse {
	resp := &AppendEntriesResponse{Term: proto.Uint64(s.currentTerm)}
	if s.currentTerm > appendRequest.GetTerm() {
		resp.Success = proto.Bool(false)
		return resp
	}
	if s.log.contains(appendRequest.GetPrevLogIndex(), appendRequest.GetPrevLogTerm()) {
		s.log.appendEntries(appendRequest.GetPrevLogIndex(), appendRequest.GetEntries())
		resp.Success = proto.Bool(true)
	} else {
		resp.Success = proto.Bool(false)
	}
	return resp
}

// TODO: fill this.
func (s *Server) candidateloop() {
	s.nodesVoteGranted = make(map[uint32]bool)
	s.leader = noleader
	s.currentTerm++ // enter candidate state. Server increments its term
	s.resetElectionTimeout()
	s.requestVotes()
	s.voteForItself()
	for s.getState() == candidate {
		select {
		case <-s.electionTimeout.C:
			// Candidate timeouts and start a new election
			return
		case resp := <-s.getVoteResponseChan:
			if rTerm := resp.GetTerm(); rTerm > s.currentTerm {
				s.currentTerm = rTerm
				s.stepDown()
				return
			}
			if resp.GetVoteGranted() {
				s.nodesVoteGranted[resp.id] = true
			}
			if s.electionPass() {
				// TODO: delete these logger
				for id, _ := range s.nodesVoteGranted {
					logger.Printf("node %d vote for node %d", id, s.id)
				}
				logger.Printf("node %d became leader\n", s.id)
				s.setState(leader)
				s.votefor = notVotedYet
				s.leader = s.id
				return
			}
		}
	}
}

// TODO: fill this
func (s *Server) stepDown() {
	s.setState(follower)
	s.votefor = notVotedYet
}

// TODO: fill this.
func (s *Server) leaderloop() {
	// 记得初始化nextIndex
	for s.getState() == leader {
		select {
		case wrappedCommand := <-s.commandChan:
			newEntry, err := s.log.newLogEntry(s.currentTerm, wrappedCommand.command)
			if err != nil {
				wrappedCommand.errChan <- err
			}
			s.log.appendEntry(newEntry)
			err = s.sendAppendEntries()
			wrappedCommand.errChan <- err
			if err == deponseError { // if leader has been deposed
				s.setState(follower)
				return
			}
		}
	}
}

// TODO: add more procedure
func (s *Server) sendAppendEntries() error {
	theOtherNodes := s.theOtherNodes()
	successChan := make(chan bool, len(theOtherNodes))
	errChan := make(chan error, len(theOtherNodes))
	for id, node := range theOtherNodes {
		go func() {
			for {
				pbAER := AppendEntriesRequest{
					LeaderID:     proto.Uint32(s.id),
					Term:         proto.Uint64(s.currentTerm),
					PrevLogIndex: proto.Uint64(s.log.prevLogIndex(s.nextIndex.get(id))),
					PrevLogTerm:  proto.Uint64(s.log.prevLogTerm(s.nextIndex.get(id))),
					CommitIndex:  proto.Uint64(s.log.commitIndex),
					Entries:      s.log.entriesAfer(s.nextIndex.get(id)),
				}
				reponse, err := node.rpcAppendEntries(s, &pbAER)
				if err != nil {
					continue
				}
				if rTerm := reponse.GetTerm(); rTerm > s.currentTerm {
					errChan <- deponseError
					return
				}
				if !reponse.GetSuccess() {
					s.nextIndex.dec(id)
					logger.Printf("append entries not succeedded, node %d dec ni to %d\n",
						id, s.nextIndex.get(id))
					continue
				}
				logger.Printf("node %d append entries succeedded\n", id)
				s.nextIndex.inc(id)
				successChan <- true
				return
			}
		}()
	}
	successCount := 0
	go func() {
		for {
			select {
			case <-successChan:
				successCount++
				if successCount == len(theOtherNodes)/2 {
					// PAPER: Agreement(forelectionsandentrycommitment)requires
					// separate majorities from both the old and new
					// configurations.
					// so this needs to be changed later
					errChan <- nil
				}
				if successCount == len(theOtherNodes) {
					return // all nodes have accept request
				}
			}
		}
	}()
	return <-errChan
}

func (s *Server) theOtherNodes() nodeMap {
	allNodeMap := make(nodeMap)
	for id, node := range s.config.c_NewNode {
		if id == s.id {
			continue
		}
		allNodeMap[id] = node
	}

	for id, node := range s.config.c_OldNode {
		if id == s.id {
			continue
		}
		allNodeMap[id] = node
	}
	return allNodeMap
}

// Exec executes client's command. client should
// use concrete and registered commands that
// implement Command interface
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
	theOtherNodes := s.theOtherNodes()
	// if re-relect, reset the voteChan
	s.getVoteResponseChan = make(chan wrappedVoteResponse, len(theOtherNodes))
	for id, node := range theOtherNodes {
		responded := false
		go func() { // send vote request simultaneously
			for !responded { // not responded, keep trying
				pb := &RequestVoteRequest{
					CandidateID:  proto.Uint32(s.id),
					Term:         proto.Uint64(s.currentTerm),
					LastLogIndex: proto.Uint64(s.log.lastLogIndex()),
					LastLogTerm:  proto.Uint64(s.log.lastLogTerm()),
				}
				responseProto, err := node.rpcRequestVote(s, pb)
				if err != nil {
					// logger.Println("vote response error:", err.Error())
					responded = false
					// does not timeout and try to send request over and over agian
					continue
				}
				responded = true
				s.getVoteResponseChan <- wrappedVoteResponse{
					id:                  id,
					RequestVoteResponse: responseProto,
				}
			}
		}()
	}
}

func (s *Server) electionPass() bool {
	if s.config.getState() == c_old {
		votesCount := 0
		for id := range s.config.c_OldNode {
			if s.nodesVoteGranted[id] == true {
				votesCount++
			}
		}
		if votesCount >= len(s.config.c_OldNode)/2+1 {
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

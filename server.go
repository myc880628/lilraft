package lilraft

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
)

func init() {
	gob.Register(&HTTPNode{})
}

// Logger
var logger = log.New(os.Stdout, "[lilraft]", log.Lmicroseconds)

// Error
var (
	errDepose   = fmt.Errorf("leader got deposed")
	errNoLeader = fmt.Errorf("cannot find leader")
)

// state constant
const (
	follower = iota
	candidate
	leader
	stopped
	noleader
	cOld
	cOldNew
)

var (
	minTimeout = 150            // minimum timeout: 150 ms
	maxTimeout = minTimeout * 2 //maximum timeout: 300ms
)

func randomElectionTimeout() time.Duration {
	d := minTimeout + rand.Intn(maxTimeout-minTimeout)
	return time.Duration(d) * time.Millisecond
}

func hearbeatInterval() time.Duration {
	return time.Duration(minTimeout/10) * time.Millisecond
}

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
	m map[int32]int64
	sync.RWMutex
}

func (ni *nextIndex) get(id int32) int64 {
	ni.RLock()
	defer ni.RUnlock()
	if index, ok := ni.m[id]; !ok {
		panic("lilraft: wrong node in []NextIndex")
	} else {
		return index
	}
}

func (ni *nextIndex) set(id int32, index int64) {
	ni.Lock()
	ni.m[id] = index
	ni.Unlock()
}

func (ni *nextIndex) inc(id int32) {
	ni.Lock()
	ni.m[id]++
	ni.Unlock()
}

func (ni *nextIndex) dec(id int32) {
	ni.Lock()
	if ni.m[id] > 0 {
		ni.m[id]--
	}
	ni.Unlock()
}

// Server is a concrete machine that process command, appendentries and requestvote, etc.
type Server struct {
	//----from the paper------
	voted            bool
	currentTerm      int64
	lastAppliedIndex int64
	matchIndex       []int64
	nextIndex        nextIndex
	log              *Log
	//------------------------
	id                   int32
	leader               int32
	context              interface{}
	electionTimeout      *time.Timer
	httpClient           http.Client
	config               *configuration
	getVoteRequestChan   chan wrappedVoteRequest
	getAppendEntriesChan chan wrappedAppendRequest
	getConfigChan        chan wrappedSetConfig
	// responseChan
	//-------leader only------------
	deposeChan  chan int64
	commandChan chan wrappedCommand
	//-----candidate only-----------
	getVoteResponseChan chan wrappedVoteResponse
	nodesVoteGranted    map[int32]bool
	//-----------------------------
	path string
	mutexState
}

type wrappedVoteResponse struct {
	id int32
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

type wrappedSetConfig struct {
	nodes   []Node
	errChan chan error
}

func (s *Server) resetElectionTimeout() {
	s.electionTimeout = time.NewTimer(randomElectionTimeout())
}

// Exec executes client's command. client should
// use concrete and registered commands that
// implement Command interface
func (s *Server) Exec(command Command) error {
	errChan := make(chan error)
	logger.Println(command)
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

// SetConfig allows client to change the nodes of configuration
func (s *Server) SetConfig(nodes ...Node) error {
	errChan := make(chan error)
	s.getConfigChan <- wrappedSetConfig{
		nodes:   nodes,
		errChan: errChan,
	}
	return <-errChan
}

// NewServer can return a new server for clients
func NewServer(id int32, context interface{}, config *configuration, path string) (s *Server) {
	if context == nil {
		panic("lilraft: contex is required")
	}
	if id <= 0 {
		panic("lilraft: id must be > 0")
	}
	if config == nil {
		panic("lilraft: configuration unset")
	}
	s = &Server{
		id:                   id,
		leader:               noleader,
		context:              context,
		path:                 path,
		currentTerm:          0,
		log:                  newLog(),
		voted:                false,
		config:               config,
		nextIndex:            nextIndex{m: make(map[int32]int64)},
		commandChan:          make(chan wrappedCommand),
		getVoteRequestChan:   make(chan wrappedVoteRequest),
		getAppendEntriesChan: make(chan wrappedAppendRequest),
		getConfigChan:        make(chan wrappedSetConfig),
	}
	s.mutexState.setState(stopped)
	// http.Client can cache TCP connections
	s.httpClient.Transport = &http.Transport{DisableKeepAlives: false}
	s.httpClient.Timeout = time.Duration(minTimeout/2) * time.Millisecond
	rand.Seed(time.Now().Unix()) // random seed for election timeout
	s.resetElectionTimeout()
	err := os.MkdirAll(path, 0700)
	if err != nil && !os.IsExist(err) {
		panic(err.Error())
	}
	if err := s.log.recover(s.logPath(), s.context); err != nil {
		logger.Println("server recover", s.logPath(), "error: ", err.Error())
	}
	return
}

// Start starts a server, remember to call NewServer before call this.
func (s *Server) Start() {
	s.setState(follower) // all servers start as follower
	go s.loop()
}

// Stop stops a server by setting server's state(a integer) to stopped
func (s *Server) Stop() {
	s.setState(stopped)
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
			return
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
			wrappedVoteRequest.responseChan <- s.handleVoteRequest(voteRequest)
		case wrappedAppendRequest := <-s.getAppendEntriesChan:
			appendRequest := wrappedAppendRequest.request
			wrappedAppendRequest.responseChan <- s.handleAppendEntriesRequest(appendRequest)
		case wrappedCommand := <-s.commandChan:
			wrappedCommand.errChan <- s.redirectClientCommand(wrappedCommand.command)
		case wrappedSetConfig := <-s.getConfigChan:
			wrappedSetConfig.errChan <- s.redirectClientConfig(wrappedSetConfig.nodes)
		}
	}
}

func (s *Server) redirectClientCommand(command Command) error {
	if s.leader == noleader {
		return errNoLeader
	}
	logger.Printf("node %d redirecting client to node %d", s.id, s.leader)
	var bytesBuffer bytes.Buffer
	if err := json.NewEncoder(&bytesBuffer).Encode(command); err != nil {
		return err
	}
	redirectedCommand := &RedirectedCommand{
		CommandName: proto.String(command.Name()),
		Command:     bytesBuffer.Bytes(),
	}
	return s.theOtherNodes()[s.leader].rpcCommand(s, redirectedCommand)
}

func (s *Server) redirectClientConfig(nodes []Node) error {
	if s.leader == noleader {
		return errNoLeader
	}
	logger.Printf("node %d redirecting config to node %d", s.id, s.leader)

	return s.theOtherNodes()[s.leader].rpcSetConfig(s, nodes)
}

func (s *Server) candidateloop() {
	s.nodesVoteGranted = make(map[int32]bool)
	s.leader = noleader
	s.updateCurrentTerm(s.currentTerm + 1) // enter candidate state. Server increments its term
	s.resetElectionTimeout()
	s.requestVotes()
	s.voteForItself()
	for s.getState() == candidate {
		select {
		case <-s.electionTimeout.C:
			// Candidate timeouts and start a new election
			return
		case resp := <-s.getVoteResponseChan:
			if resp.GetTerm() > s.currentTerm {
				s.updateCurrentTerm(resp.GetTerm())
				s.stepDown()
				return
			}
			if resp.GetVoteGranted() {
				s.nodesVoteGranted[resp.id] = true
			}
			if s.electionPass() {
				// TODO: delete these logger
				for id := range s.nodesVoteGranted {
					logger.Printf("node %d vote for node %d", id, s.id)
				}
				logger.Printf("node %d became leader\n", s.id)
				s.setState(leader)
				s.voted = false
				s.leader = s.id
				return
			}
		case wrappedVoteRequest := <-s.getVoteRequestChan:
			voteRequest := wrappedVoteRequest.request
			if voteRequest.GetTerm() > s.currentTerm {
				s.voted = false // candidate has voted for itself, we need to change that
				s.stepDown()
				// candidate become follower
				wrappedVoteRequest.responseChan <- s.handleVoteRequest(voteRequest)
				return
			}
			// candidate does not step down
			wrappedVoteRequest.responseChan <- &RequestVoteResponse{
				Term:        proto.Int64(s.currentTerm),
				VoteGranted: proto.Bool(false),
			}
		case wrappedAppendRequest := <-s.getAppendEntriesChan:
			appendRequest := wrappedAppendRequest.request
			if appendRequest.GetTerm() >= s.currentTerm {
				s.stepDown()
				wrappedAppendRequest.responseChan <- s.handleAppendEntriesRequest(appendRequest)
				return
			}
			wrappedAppendRequest.responseChan <- &AppendEntriesResponse{
				Term:    proto.Int64(s.currentTerm),
				Success: proto.Bool(false),
			}
		case wrappedCommand := <-s.commandChan:
			wrappedCommand.errChan <- fmt.Errorf("leader doesn't exist")
		case wrappedSetConfig := <-s.getConfigChan:
			wrappedSetConfig.errChan <- fmt.Errorf("leader doesn't exist")
		}
	}
}

func (s *Server) leaderloop() {
	theOtherNdoes := s.theOtherNodes()
	s.deposeChan = make(chan int64, len(theOtherNdoes))
	// initialize nextIndex
	for id := range theOtherNdoes {
		s.nextIndex.set(id, s.log.lastLogIndex())
	}
	hearbeatChan := make(chan struct{})
	heartBeat := time.NewTicker(hearbeatInterval())
	defer heartBeat.Stop()
	go func() {
		for _ = range heartBeat.C {
			// s.broadcastHeartbeats()
			hearbeatChan <- struct{}{}
		}
	}()

	for s.getState() == leader {
		select {
		case wrappedCommand := <-s.commandChan:
			newEntry, err := s.log.newLogEntry(s.currentTerm, wrappedCommand.command)
			if err != nil {
				wrappedCommand.errChan <- err
			}
			s.log.appendEntry(newEntry)
			err = s.sendAppendEntries()
			if err != nil {
				logger.Println("append entries error: ", err.Error())
				wrappedCommand.errChan <- err
				continue
			}
			wrappedCommand.errChan <- s.log.setCommitIndex(s.log.lastLogIndex(), s.context)
		case wrappedSetConfig := <-s.getConfigChan:
			go func() {
				wrappedSetConfig.errChan <- s.processSetConfig(&wrappedSetConfig)
			}()
		case wrappedVoteRequest := <-s.getVoteRequestChan:
			voteRequest := wrappedVoteRequest.request
			if voteRequest.GetTerm() > s.currentTerm {
				s.stepDown()
			}
			// now leader stepped down and become follower
			wrappedVoteRequest.responseChan <- s.handleVoteRequest(voteRequest)
		case wrappedAppendRequest := <-s.getAppendEntriesChan:
			appendRequest := wrappedAppendRequest.request
			if appendRequest.GetTerm() > s.currentTerm {
				s.stepDown()
			}
			// now leader stepped down and become follower
			wrappedAppendRequest.responseChan <- s.handleAppendEntriesRequest(appendRequest)
		case responseTerm := <-s.deposeChan:
			s.stepDown()
			s.updateCurrentTerm(responseTerm)
		case <-hearbeatChan:
			s.broadcastHeartbeats()
		}
	}
}

// TODO: refactor this
func (s *Server) broadcastHeartbeats() {
	theOtherNdoes := s.theOtherNodes()
	for id, node := range theOtherNdoes {
		go func(id int32, node Node) {
			response, err := s.sendAppendRequestTo(id, node)
			if err != nil {
				logger.Println(err.Error())
			}
			if rTerm := response.GetTerm(); rTerm > s.currentTerm {
				s.deposeChan <- rTerm
			}
			if response.GetSuccess() {
				s.nextIndex.set(id, s.log.lastLogIndex())
			} else {
				s.nextIndex.dec(id)
			}
		}(id, node)
	}
}

func (s *Server) sendAppendRequestTo(id int32, node Node) (*AppendEntriesResponse, error) {
	appendRequest := AppendEntriesRequest{
		LeaderID:     proto.Int32(s.id),
		Term:         proto.Int64(s.currentTerm),
		PrevLogIndex: proto.Int64(s.log.prevLogIndex(s.nextIndex.get(id))),
		PrevLogTerm:  proto.Int64(s.log.prevLogTerm(s.nextIndex.get(id))),
		CommitIndex:  proto.Int64(s.log.commitIndex),
		Entries:      s.log.entriesAfer(s.nextIndex.get(id)),
	}
	return node.rpcAppendEntries(s, &appendRequest)
}

func (s *Server) sendAppendEntries() error {
	theOtherNodes := s.theOtherNodes()
	oldNode := s.config.cOldNode
	newNode := s.config.cNewNode
	successChan := make(chan int32, len(theOtherNodes))
	errChan := make(chan error, len(theOtherNodes))
	for id, node := range theOtherNodes {
		go func(id int32, node Node) {
			for {
				reponse, err := s.sendAppendRequestTo(id, node)
				if err != nil {
					continue
				}
				if rTerm := reponse.GetTerm(); rTerm > s.currentTerm {
					errChan <- errDepose
					s.deposeChan <- rTerm // leader stale
					return
				}
				if !reponse.GetSuccess() {
					s.nextIndex.dec(id)
					logger.Printf("append entries not succeedded, node %d dec ni to %d\n",
						id, s.nextIndex.get(id))
					continue
				}
				logger.Printf("node %d append entries succeedded\n", id)
				s.nextIndex.set(id, s.log.lastLogIndex())
				successChan <- id
				return
			}
		}(id, node)
	}
	oldSuccessCount := 1
	newSuccessCount := 1
	if newNode != nil && newNode[s.id] == nil {
		newSuccessCount = 0
	}
	go func() {
		acceptCount := 0
		for {
			select {
			case id := <-successChan:
				acceptCount++
				if oldNode[id] != nil {
					oldSuccessCount++
				}
				if newNode[id] != nil {
					newSuccessCount++
				}
				if oldSuccessCount >= len(oldNode)/2+1 && newSuccessCount >= len(newNode)/2+1 {
					// PAPER: Agreement(for elections and entry commitment)requires
					// separate majorities from both the old and new
					// configurations.
					errChan <- nil
				}
				if acceptCount >= len(theOtherNodes) {
					return // all nodes have accept request
				}
			}
		}
	}()
	return <-errChan
}

func (s *Server) theOtherNodes() nodeMap {
	nMap := make(nodeMap)
	for id, node := range s.config.cNewNode {
		if id == s.id {
			continue
		}
		nMap[id] = node
	}

	for id, node := range s.config.cOldNode {
		if id == s.id {
			continue
		}
		nMap[id] = node
	}
	return nMap
}

func (s *Server) updateCurrentTerm(term int64) {
	s.currentTerm = term
	s.voted = false // follower vote exactly once on one specific term
}

func (s *Server) handleAppendEntriesRequest(appendRequest *AppendEntriesRequest) *AppendEntriesResponse {
	resp := &AppendEntriesResponse{
		Term:    proto.Int64(s.currentTerm),
		Success: proto.Bool(false)}
	if s.currentTerm > appendRequest.GetTerm() {
		return resp
	}
	if s.currentTerm < appendRequest.GetTerm() {
		s.updateCurrentTerm(appendRequest.GetTerm())
	}
	if s.log.contains(appendRequest.GetPrevLogIndex(), appendRequest.GetPrevLogTerm()) {
		s.log.appendEntries(s, appendRequest.GetPrevLogIndex(), appendRequest.GetEntries())
		resp.Success = proto.Bool(true)

		if s.leader == noleader {
			s.leader = appendRequest.GetLeaderID()
		}

		if appendRequest.GetCommitIndex() > s.log.commitIndex {
			s.log.setCommitIndex(appendRequest.GetCommitIndex(), s.context)
		}

		s.resetElectionTimeout()
	}
	return resp
}

func (s *Server) handleVoteRequest(voteRequest *RequestVoteRequest) *RequestVoteResponse {
	resp := &RequestVoteResponse{
		Term:        proto.Int64(s.currentTerm),
		VoteGranted: proto.Bool(false),
	}
	if s.voted == true { // follower vote exactly once on one specific term
		return resp
	} else if s.currentTerm >= voteRequest.GetTerm() {
		return resp
	} else if voteRequest.GetLastLogTerm() >= s.log.lastLogTerm() && voteRequest.GetLastLogIndex() >= s.log.lastLogIndex() {
		// candidate's log is at leat up-to-date as Server s.
		// logger.Printf("node %d get vote")
		resp.VoteGranted = proto.Bool(true)
		s.voted = true
	}
	return resp
}

func (s *Server) voteForItself() {
	s.voted = true
	s.nodesVoteGranted[s.id] = true
}

func (s *Server) requestVotes() {
	theOtherNodes := s.theOtherNodes()
	// if re-relect, reset the voteChan
	s.getVoteResponseChan = make(chan wrappedVoteResponse, len(theOtherNodes))
	for id, node := range theOtherNodes {
		go func(id int32, node Node) { // send vote request simultaneously
			for { // not responded, keep trying
				logger.Printf("node %d send to node %d\n", s.id, id)
				pb := &RequestVoteRequest{
					CandidateID:  proto.Int32(s.id),
					Term:         proto.Int64(s.currentTerm),
					LastLogIndex: proto.Int64(s.log.lastLogIndex()),
					LastLogTerm:  proto.Int64(s.log.lastLogTerm()),
				}
				responseProto, err := node.rpcRequestVote(s, pb)
				if err == nil {
					if responseProto.GetVoteGranted() {
						logger.Printf("node %d got node %d granted\n", s.id, id)
					} else {
						logger.Printf("node %d got node %d reject\n", s.id, id)
					}
					s.getVoteResponseChan <- wrappedVoteResponse{
						id:                  id,
						RequestVoteResponse: responseProto,
					}
					return
				}
				logger.Println("vote response error:", err.Error())
			}
		}(id, node)
	}
}

func (s *Server) electionPass() bool {
	if s.config.getState() == cOld {
		votesCount := 0
		for id := range s.config.cOldNode {
			if s.nodesVoteGranted[id] == true {
				votesCount++
			}
		}
		if votesCount >= len(s.config.cOldNode)/2+1 {
			return true
		}
	} else if s.config.getState() == cOldNew { // candidate must be approved by Cold and Cold,new.
		voteCountOld := 0
		voteCountOldNew := 0
		for id := range s.config.cOldNode {
			if s.nodesVoteGranted[id] == true {
				voteCountOld++
			}
		}
		for id := range s.config.cNewNode {
			if s.nodesVoteGranted[id] == true {
				voteCountOldNew++
			}
		}
		if voteCountOld >= len(s.config.cOldNode)/2+1 && voteCountOldNew >= len(s.config.cNewNode)/2+1 {
			// if pass both old and new config
			return true
		}
	} else {
		panic("lilraft: config state unknown")
	}
	return false
}

// candidate or leader will call this if they detect legit leader
func (s *Server) stepDown() {
	s.setState(follower)
	s.leader = noleader
	logger.Printf("node %d step down to follower\n", s.id)
}

func (s *Server) processSetConfig(wrappedSetConfig *wrappedSetConfig) error {
	nodes := wrappedSetConfig.nodes
	cOldNewEntry, err := s.log.newConfigEntry(s.currentTerm, nodes, cOldNewStr)
	if err != nil {
		return err
	}
	s.config.setState(cOldNew)
	s.config.cNewNode = makeNodeMap(nodes...)
	for _, node := range nodes {
		if _, ok := s.nextIndex.m[node.id()]; !ok {
			s.nextIndex.set(node.id(), s.log.lastLogIndex())
		}
	}
	s.log.appendEntry(cOldNewEntry)
	err = s.sendAppendEntries()
	if err != nil {
		logger.Println("append Coldnew config entries error: ", err.Error())
		return err
	}
	if err := s.log.setCommitIndex(s.log.lastLogIndex(), s.context); err != nil {
		return err
	}
	// now cOldNew has send to majorities of c_new and cOld and commit ColdNew
	cNewEntry, err := s.log.newConfigEntry(s.currentTerm, nodes, cNewStr)
	if err != nil {
		return err
	}
	s.log.appendEntry(cNewEntry)
	s.config.setState(cOld)
	s.config.cOldNode = makeNodeMap(nodes...)
	s.config.cNewNode = nil
	err = s.sendAppendEntries()
	if err != nil {
		logger.Println("append Cnew config entries error: ", err.Error())
		return err
	}
	if err := s.log.setCommitIndex(s.log.lastLogIndex(), s.context); err != nil {
		return err
	}
	// now c_new has been commited
	if s.config.cOldNode[s.id] == nil {
		s.setState(stopped)
	}
	return nil
}

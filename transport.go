package lilraft

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"
)

const (
	idPath            = "/lilraft/id"
	appendEntriesPath = "/lilraft/appendentries"
	requestVotePath   = "/lilraft/requestvote"
	commandPath       = "/lilraft/command"
	setConfigPath     = "/lilraft/setconfig"
)

func init() {
	gob.Register(RetValErr{})
}

// SetHTTPTransport accept a http multiplexer to handle other peers'
// RPCs like RequestVote and AppendEntry, etc.
// Run this before running server.Start().
func (s *Server) SetHTTPTransport(mux *http.ServeMux, port int) {
	mux.HandleFunc(idPath, idHandleFunc(s))
	mux.HandleFunc(appendEntriesPath, appendEntriesHandler(s))
	mux.HandleFunc(requestVotePath, requestVoteHandler(s))
	mux.HandleFunc(commandPath, commandHandler(s))
	mux.HandleFunc(setConfigPath, setConfigHandler(s))
	go http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func idHandleFunc(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprint(s.id)))
	}
}

func appendEntriesHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		appendEntriesRequest := &AppendEntriesRequest{}
		if err := Decode(r.Body, appendEntriesRequest); err != nil {
			logger.Printf("node %d decode appendEntriesRequest err: %s \n", s.id, err.Error())
			return
		}
		// logger.Printf("node %d: append entries in comming!", s.id)
		respChan := make(chan *AppendEntriesResponse)
		s.getAppendEntriesChan <- wrappedAppendRequest{
			request:      appendEntriesRequest,
			responseChan: respChan,
		}
		responseProto := <-respChan
		if err := Encode(w, responseProto); err != nil {
			logger.Println("encode error: ", err.Error())
		}
	}
}

func requestVoteHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		voteRequest := &RequestVoteRequest{}
		if err := Decode(r.Body, voteRequest); err != nil {
			logger.Printf("node %d decode voteRequest error: %s\n", s.id, err.Error())
			return
		}
		logger.Printf("node %d: Vote Request in comming! node %d term %d asks for vote, lastLogIndex %d, lastLogTerm %d",
			s.id, voteRequest.GetCandidateID(),
			voteRequest.GetTerm(),
			voteRequest.GetLastLogIndex(),
			voteRequest.GetLastLogTerm(),
		)
		respChan := make(chan *RequestVoteResponse)
		s.getVoteRequestChan <- wrappedVoteRequest{
			request:      voteRequest,
			responseChan: respChan,
		}
		responseProto := <-respChan
		if err := Encode(w, responseProto); err != nil {
			logger.Println("encode error: ", err.Error())
		}
	}
}

type RetValErr struct {
	Val interface{}
	Err error
}

// TODO: fill this func
func commandHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		rCommand := &RedirectedCommand{}
		if err := Decode(r.Body, rCommand); err != nil {
			logger.Printf("node %d decode voteRequest error: %s\n", s.id, err.Error())
			w.Write([]byte("failed"))
			return
		}
		logger.Println("node ", s.id, " get command")
		command, err := newCommand(rCommand.GetCommandName(), rCommand.GetCommand())
		if err != nil {
			logger.Printf("node %d generate new command error: %s\n", s.id, err.Error())
			w.Write([]byte("failed"))
			return
		}

		val, err := s.Exec(command)
		var bytesBuffer bytes.Buffer
		// var retValErr RetValErr
		gob.NewEncoder(&bytesBuffer).Encode(RetValErr{
			Val: val,
			Err: err,
		})
		b := bytesBuffer.Bytes()
		println("send bytes back: ", b)
		w.Write(b)
	}
}

// TODO: fill this func
func setConfigHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		logger.Println("config in comming")
		// httpNodes := []*HTTPNode{}
		nodes := []Node{}
		if err := gob.NewDecoder(r.Body).Decode(&nodes); err != nil {
			logger.Printf("node %d decode config error: %s\n", s.id, err.Error())
		}
		if err := s.SetConfig(nodes...); err != nil {
			w.Write([]byte("failed"))
			return
		}
		w.Write([]byte("success"))
	}
}

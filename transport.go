package lilraft

import (
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
		logger.Printf("node %d: append entries in comming!\n", s.id)
		appendEntriesRequest := &AppendEntriesRequest{}
		if err := Decode(r.Body, appendEntriesRequest); err != nil {
			logger.Println("Decode appendEntriesRequest err: ", err.Error())
		}
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
			logger.Println("decode voteRequest error: ", err.Error())
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

// TODO: fill this func
func commandHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// TODO: fill this func
func setConfigHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

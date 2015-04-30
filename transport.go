package lilraft

import (
	"fmt"
	"net/http"
)

const (
	idPath            = "lilraft/id"
	appendEntriesPath = "lilraft/appendentries"
	requestVotePath   = "lilraft/requestvote"
	commandPath       = "lilraft/command"
	setConfigPath     = "lilraft/setconfig"
)

// SetHTTPTransport accept a http multiplexer to handle other peers'
// RPCs like RequestVote and AppendEntry, etc.
// Run this before running server.Start().
func (s *Server) SetHTTPTransport(mux *http.ServeMux) {
	mux.HandleFunc(idPath, idHandleFunc(s))
	mux.HandleFunc(appendEntriesPath, appendEntriesHandleFunc(s))
	mux.HandleFunc(requestVotePath, requestVoteHandleFunc(s))
	mux.HandleFunc(commandPath, commandHandleFunc(s))
	mux.HandleFunc(setConfigPath, setConfigHandleFunc(s))
}

func idHandleFunc(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprint(s.id)))
	}
}

// TODO: fill this func
func appendEntriesHandleFunc(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// TODO: fill this func
func requestVoteHandleFunc(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// TODO: fill this func
func commandHandleFunc(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// TODO: fill this func
func setConfigHandleFunc(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

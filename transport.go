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

// RunHTTPTransport accept a server and http multiplexer to handle other peers'
// RPCs like RequestVote and AppendEntry, etc.
// Run this before running NewServer.
func (s *Server) RunHTTPTransport(mux *http.ServeMux) {
	mux.HandleFunc(idPath, idHandleFunc(s))
	mux.HandleFunc(appendEntriesPath, appendEntriesHandleFunc(s))
	mux.HandleFunc(requestVotePath, requestVoteHandleFunc(s))
	mux.HandleFunc(commandPath, commandHandleFunc(s))
	mux.HandleFunc(setConfigPath, setConfigHandleFunc(s))
}

func idHandleFunc(s *server) http.HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprint(s.id)))
	}
}

// TODO: fill this func
func appendEntriesHandleFunc(s *server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// TODO: fill this func
func requestVoteHandleFunc(s *server) http.HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// TODO: fill this func
func commandHandleFunc(s *server) http.HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// TODO: fill this func
func setConfigHandleFunc(s *server) http.HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

package lilraft

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
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
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func idHandleFunc(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprint(s.id)))
	}
}

// TODO: fill this func
func appendEntriesHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// TODO: fill this func
func requestVoteHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var voteRequest *RequestVoteRequest
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Println("Read vote request error")
		}
		if err = proto.Unmarshal(data, voteRequest); err != nil {
			logger.Println("unmarshal RequestVoteRequest error")
		}
		respChan := make(chan *RequestVoteResponse)
		s.getVoteRequestChan <- wrappedVoteRequest{
			request:      voteRequest,
			responseChan: respChan,
		}
		<-respChan
		// TODO: Add more procedure
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

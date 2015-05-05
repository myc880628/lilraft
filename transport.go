package lilraft

import (
	"fmt"
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
	go http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
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
		logger.Println("Vote Request in comming!")
		voteRequest := &RequestVoteRequest{}
		if err := Decode(r.Body, voteRequest); err != nil {
			logger.Println("decode voteRequest error: ", err.Error())
		}
		respChan := make(chan *RequestVoteResponse)
		s.getVoteRequestChan <- wrappedVoteRequest{
			request:      voteRequest,
			responseChan: respChan,
		}
		responseProto := <-respChan
		responseBytes, err := proto.Marshal(responseProto)
		if err != nil {
			logger.Println("marshal response proto error")
		}
		if _, err = w.Write(responseBytes); err != nil {
			logger.Println("write to node: ", err)
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

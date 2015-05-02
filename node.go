package lilraft

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/golang/protobuf/proto"
)

// Node represent the remote machine ndoe
type Node interface {
	id() uint32
	rpcAppendEntries(*AppendEntriesRequest) *AppendEntriesResponse
	rpcRequestVote(*Server, *RequestVoteRequest) (*RequestVoteResponse, error)
}

// nodeMap wraps some useful function with a map
//  for server to send rpc to remote machine.
type nodeMap map[uint32]*Node

type httpNode struct {
	remoteId uint32
	url      *url.URL
}

func (node *httpNode) id() uint32 {
	return node.remoteId
}

func (node *httpNode) rpcRequestVote(server *Server, rvr *RequestVoteRequest) (*RequestVoteResponse, error) {
	rvrbytes, err := proto.Marshal(rvr)
	url := fmt.Sprintf("%s%s", node.url.String(), requestVotePath)
	if err != nil {
		return nil, err
	}
	var bytesBuffer bytes.Buffer
	if _, err = bytesBuffer.Write(rvrbytes); err != nil {
		return nil, err
	}

	// Send request to node
	httpResponse, err := server.httpClient.Post(url, "application/protobuf", &bytesBuffer)
	if err != nil {
		return nil, err
	}
	reponseData, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}
	responseProto := &RequestVoteResponse{}
	if err = proto.Unmarshal(reponseData, responseProto); err != nil {
		return nil, err
	}
	return responseProto, nil
}

// TODO: fill this
func (node *httpNode) rpcAppendEntries(aer *AppendEntriesRequest) *AppendEntriesResponse {
	return nil
}

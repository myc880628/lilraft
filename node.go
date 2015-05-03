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
	rpcAppendEntries(*Server, *AppendEntriesRequest) (*AppendEntriesResponse, error)
	rpcRequestVote(*Server, *RequestVoteRequest) (*RequestVoteResponse, error)
}

// nodeMap wraps some useful function with a map
//  for server to send rpc to remote machine.
type nodeMap map[uint32]Node

type HTTPNode struct {
	remoteId uint32
	url      *url.URL
}

func (node *HTTPNode) id() uint32 {
	return node.remoteId
}

// TODO: refactor this maybe.
func (node *HTTPNode) rpcRequestVote(server *Server, rvr *RequestVoteRequest) (*RequestVoteResponse, error) {
	rvrbytes, err := proto.Marshal(rvr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s%s", node.url.String(), requestVotePath)
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

func (node *HTTPNode) rpcAppendEntries(server *Server, aer *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	aerBytes, err := proto.Marshal(aer)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s%s", node.url.String(), appendEntriesPath)
	var bytesBuffer bytes.Buffer
	if _, err = bytesBuffer.Write(aerBytes); err != nil {
		return nil, err
	}

	// Send request to nodes
	httpResponse, err := server.httpClient.Post(url, "application/protobuf", &bytesBuffer)
	if err != nil {
		return nil, err
	}
	responseData, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}
	responseProto := &AppendEntriesResponse{}
	if err = proto.Unmarshal(responseData, responseProto); err != nil {
		return nil, err
	}
	return responseProto, nil
}

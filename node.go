package lilraft

import (
	"net/url"

	"github.com/lilwulin/lilraft/protobuf"
)

// Node represent the remote machine ndoe
type Node interface {
	rpcAppendEntries(protobuf.AppendEntriesRequest) protobuf.AppendEntriesResponse
	rpcRequestVote(protobuf.RequestVoteRequest) protobuf.RequestVoteResponse
}

// nodeMap wraps some useful function with a map
//  for server to send rpc to remote machine.
type nodeMap map[uint32]*Node

func (nmap *nodeMap) requestVotes() {

}

type httpNode struct {
	url *url.URL
}

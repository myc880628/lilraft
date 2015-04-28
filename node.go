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

type httpNode struct {
	url *url.URL
}

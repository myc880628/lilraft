package lilraft

import (
	"net/url"

	"github.com/lilwulin/lilraft/protobuf"
)

type Node interface {
	// TODO: Add response
	rpcAppendEntries(protobuf.AppendEntriesRequest) protobuf.AppendEntriesResponse
	rpcRequestVote(protobuf.RequestVoteRequest) protobuf.RequestVoteResponse
}

type HttpNode struct {
	url *url.URL
}

package lilraft

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"

	"github.com/golang/protobuf/proto"
)

// Node represent the remote machine ndoe
type Node interface {
	id() int32
	rpcAppendEntries(*Server, *AppendEntriesRequest) (*AppendEntriesResponse, error)
	rpcRequestVote(*Server, *RequestVoteRequest) (*RequestVoteResponse, error)
}

// nodeMap wraps some useful function with a map
//  for server to send rpc to remote machine.
type nodeMap map[int32]Node

func makeNodeMap(nodes ...Node) (nm nodeMap) {
	nm = nodeMap{}
	for _, node := range nodes {
		nm[node.id()] = node
	}
	return
}

type HTTPNode struct {
	ID  int32
	URL *url.URL
}

func NewHTTPNode(id int32, rawurl string) (httpNode *HTTPNode) {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	if id <= 0 {
		panic("node id should be > 0")
	}
	u.Path = ""
	logger.Printf("node url: %s\n", u.String())
	return &HTTPNode{
		ID:  id,
		URL: u,
	}
}

func (node *HTTPNode) id() int32 {
	return node.ID
}

// TODO: refactor these rpc maybe.
func (node *HTTPNode) rpcRequestVote(server *Server, rvr *RequestVoteRequest) (*RequestVoteResponse, error) {
	var bytesBuffer bytes.Buffer
	if err := Encode(&bytesBuffer, rvr); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s%s", node.URL.String(), requestVotePath)
	// Send request to node
	httpResponse, err := server.httpClient.Post(url, "application/protobuf", &bytesBuffer)
	if err != nil {
		return nil, err
	}
	// fmt.Println(httpResponse.Status)
	responseProto := &RequestVoteResponse{}
	if err = Decode(httpResponse.Body, responseProto); err != nil {
		return nil, err
	}
	return responseProto, nil
}

func (node *HTTPNode) rpcAppendEntries(server *Server, aer *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	var bytesBuffer bytes.Buffer
	if err := Encode(&bytesBuffer, aer); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s%s", node.URL.String(), appendEntriesPath)
	// Send request to nodes
	httpResponse, err := server.httpClient.Post(url, "application/protobuf", &bytesBuffer)
	if err != nil {
		return nil, err
	}
	responseProto := &AppendEntriesResponse{}
	if err = Decode(httpResponse.Body, responseProto); err != nil {
		return nil, err
	}
	return responseProto, nil
}

func Decode(body io.Reader, pb proto.Message) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	if err = proto.Unmarshal(data, pb); err != nil {
		return err
	}
	return nil
}

func Encode(w io.Writer, pb proto.Message) error {
	data, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	if _, err = w.Write(data); err != nil {
		return err
	}
	return nil
}

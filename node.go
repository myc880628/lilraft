package lilraft

import (
	"bytes"
	"encoding/gob"
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
	rpcCommand(*Server, *RedirectedCommand) (interface{}, error) // followers use this to redirect client's commands to leader
	rpcSetConfig(*Server, []Node) error                          // followers use this to redirect client's new configuration
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
	URL string
}

func NewHTTPNode(id int32, rawurl string) (httpNode *HTTPNode) {
	_, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	if id <= 0 {
		panic("node id should be > 0")
	}
	logger.Printf("node url: %s\n", rawurl)
	return &HTTPNode{
		ID:  id,
		URL: rawurl,
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
	url := fmt.Sprintf("%s%s", node.URL, requestVotePath)
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
	url := fmt.Sprintf("%s%s", node.URL, appendEntriesPath)
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

type httpCommand struct {
	// http transport needs command's name to decode command
	// to a specific type of command, so I wrapped command
	// with command's name
	name    string
	command Command
}

func (node *HTTPNode) rpcCommand(server *Server, rCommand *RedirectedCommand) (interface{}, error) {
	var bytesBuffer bytes.Buffer
	if err := Encode(&bytesBuffer, rCommand); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s%s", node.URL, commandPath)
	httpResponse, err := server.httpClient.Post(url, "application/protobuf", &bytesBuffer)
	if err != nil {
		return nil, err
	}
	// var resp []byte
	// if _, err := httpResponse.Body.Read(resp); err != nil {
	// 	return nil, err
	// }
	// if string(resp) == "failed" {
	// 	return nil, fmt.Errorf("node %d exec command failed", node.id())
	// }
	// return nil

	resp, _ := ioutil.ReadAll(httpResponse.Body)
	println("get response : ", resp)
	var retValErr RetValErr
	gob.NewDecoder(bytes.NewReader(resp)).Decode(&retValErr)
	return retValErr.Val, retValErr.Err
}

// TODO: this needs to be refactor
func (node *HTTPNode) rpcSetConfig(server *Server, nodes []Node) error {
	var bytesBuffer bytes.Buffer
	if err := gob.NewEncoder(&bytesBuffer).Encode(nodes); err != nil {
		return err
	}
	url := fmt.Sprintf("%s%s", node.URL, setConfigPath)
	httpResponse, err := server.httpClient.Post(url, "application/protobuf", &bytesBuffer)
	if err != nil {
		return err
	}
	var resp []byte
	if _, err := httpResponse.Body.Read(resp); err != nil {
		return err
	}
	if string(resp) == "failed" {
		return fmt.Errorf("node %d exec command failed", node.id())
	}
	return nil
}

func Decode(r io.Reader, pb proto.Message) error {
	var length int
	_, err := fmt.Fscanf(r, "%8x\n", &length)
	if err != nil {
		return err
	}
	data := make([]byte, length)
	_, err = io.ReadFull(r, data)
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
	if _, err = fmt.Fprintf(w, "%8x\n", len(data)); err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

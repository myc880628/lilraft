// Code generated by protoc-gen-go.
// source: raftpb.proto
// DO NOT EDIT!

/*
Package protobuf is a generated protocol buffer package.

It is generated from these files:
	raftpb.proto

It has these top-level messages:
	LogEntry
	AppendEntriesRequest
	AppendEntriesResponse
	RequestVoteRequest
	RequestVoteResponse
*/
package protobuf

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type LogEntry struct {
	Index            *uint64 `protobuf:"varint,1,req" json:"Index,omitempty"`
	Term             *uint64 `protobuf:"varint,2,req" json:"Term,omitempty"`
	CommandName      *string `protobuf:"bytes,3,req" json:"CommandName,omitempty"`
	Command          []byte  `protobuf:"bytes,4,opt" json:"Command,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *LogEntry) Reset()         { *m = LogEntry{} }
func (m *LogEntry) String() string { return proto.CompactTextString(m) }
func (*LogEntry) ProtoMessage()    {}

func (m *LogEntry) GetIndex() uint64 {
	if m != nil && m.Index != nil {
		return *m.Index
	}
	return 0
}

func (m *LogEntry) GetTerm() uint64 {
	if m != nil && m.Term != nil {
		return *m.Term
	}
	return 0
}

func (m *LogEntry) GetCommandName() string {
	if m != nil && m.CommandName != nil {
		return *m.CommandName
	}
	return ""
}

func (m *LogEntry) GetCommand() []byte {
	if m != nil {
		return m.Command
	}
	return nil
}

type AppendEntriesRequest struct {
	LeaderID         *uint64     `protobuf:"varint,1,req" json:"LeaderID,omitempty"`
	Term             *uint64     `protobuf:"varint,2,req" json:"Term,omitempty"`
	PrevLogIndex     *uint64     `protobuf:"varint,3,req" json:"PrevLogIndex,omitempty"`
	PrevLogTerm      *uint64     `protobuf:"varint,4,req" json:"PrevLogTerm,omitempty"`
	CommitIndex      *uint64     `protobuf:"varint,5,req" json:"CommitIndex,omitempty"`
	Entries          []*LogEntry `protobuf:"bytes,6,rep" json:"Entries,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *AppendEntriesRequest) Reset()         { *m = AppendEntriesRequest{} }
func (m *AppendEntriesRequest) String() string { return proto.CompactTextString(m) }
func (*AppendEntriesRequest) ProtoMessage()    {}

func (m *AppendEntriesRequest) GetLeaderID() uint64 {
	if m != nil && m.LeaderID != nil {
		return *m.LeaderID
	}
	return 0
}

func (m *AppendEntriesRequest) GetTerm() uint64 {
	if m != nil && m.Term != nil {
		return *m.Term
	}
	return 0
}

func (m *AppendEntriesRequest) GetPrevLogIndex() uint64 {
	if m != nil && m.PrevLogIndex != nil {
		return *m.PrevLogIndex
	}
	return 0
}

func (m *AppendEntriesRequest) GetPrevLogTerm() uint64 {
	if m != nil && m.PrevLogTerm != nil {
		return *m.PrevLogTerm
	}
	return 0
}

func (m *AppendEntriesRequest) GetCommitIndex() uint64 {
	if m != nil && m.CommitIndex != nil {
		return *m.CommitIndex
	}
	return 0
}

func (m *AppendEntriesRequest) GetEntries() []*LogEntry {
	if m != nil {
		return m.Entries
	}
	return nil
}

type AppendEntriesResponse struct {
	Term             *uint64 `protobuf:"varint,1,req" json:"Term,omitempty"`
	Success          *bool   `protobuf:"varint,2,req" json:"Success,omitempty"`
	Reason           *string `protobuf:"bytes,3,req" json:"Reason,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *AppendEntriesResponse) Reset()         { *m = AppendEntriesResponse{} }
func (m *AppendEntriesResponse) String() string { return proto.CompactTextString(m) }
func (*AppendEntriesResponse) ProtoMessage()    {}

func (m *AppendEntriesResponse) GetTerm() uint64 {
	if m != nil && m.Term != nil {
		return *m.Term
	}
	return 0
}

func (m *AppendEntriesResponse) GetSuccess() bool {
	if m != nil && m.Success != nil {
		return *m.Success
	}
	return false
}

func (m *AppendEntriesResponse) GetReason() string {
	if m != nil && m.Reason != nil {
		return *m.Reason
	}
	return ""
}

type RequestVoteRequest struct {
	CandidateID      *uint64 `protobuf:"varint,1,req" json:"CandidateID,omitempty"`
	Term             *uint64 `protobuf:"varint,2,req" json:"Term,omitempty"`
	LastLogIndex     *uint64 `protobuf:"varint,3,req" json:"LastLogIndex,omitempty"`
	LastLogTerm      *uint64 `protobuf:"varint,4,req" json:"LastLogTerm,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *RequestVoteRequest) Reset()         { *m = RequestVoteRequest{} }
func (m *RequestVoteRequest) String() string { return proto.CompactTextString(m) }
func (*RequestVoteRequest) ProtoMessage()    {}

func (m *RequestVoteRequest) GetCandidateID() uint64 {
	if m != nil && m.CandidateID != nil {
		return *m.CandidateID
	}
	return 0
}

func (m *RequestVoteRequest) GetTerm() uint64 {
	if m != nil && m.Term != nil {
		return *m.Term
	}
	return 0
}

func (m *RequestVoteRequest) GetLastLogIndex() uint64 {
	if m != nil && m.LastLogIndex != nil {
		return *m.LastLogIndex
	}
	return 0
}

func (m *RequestVoteRequest) GetLastLogTerm() uint64 {
	if m != nil && m.LastLogTerm != nil {
		return *m.LastLogTerm
	}
	return 0
}

type RequestVoteResponse struct {
	Term             *uint64 `protobuf:"varint,1,req" json:"Term,omitempty"`
	VoteGranted      *bool   `protobuf:"varint,2,req" json:"VoteGranted,omitempty"`
	Reason           *string `protobuf:"bytes,3,req" json:"Reason,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *RequestVoteResponse) Reset()         { *m = RequestVoteResponse{} }
func (m *RequestVoteResponse) String() string { return proto.CompactTextString(m) }
func (*RequestVoteResponse) ProtoMessage()    {}

func (m *RequestVoteResponse) GetTerm() uint64 {
	if m != nil && m.Term != nil {
		return *m.Term
	}
	return 0
}

func (m *RequestVoteResponse) GetVoteGranted() bool {
	if m != nil && m.VoteGranted != nil {
		return *m.VoteGranted
	}
	return false
}

func (m *RequestVoteResponse) GetReason() string {
	if m != nil && m.Reason != nil {
		return *m.Reason
	}
	return ""
}

func init() {
}

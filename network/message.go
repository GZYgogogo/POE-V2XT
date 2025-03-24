package network

import "github.com/anthdm/projectx/core"

type GetBlocksMessage struct {
	From uint32
	// If To is 0 the maximum blocks will be returned.
	To uint32
}

type BlocksMessage struct {
	Blocks []*core.Block
}

type GetStatusMessage struct{}

type StatusMessage struct {
	// the id of the server
	ID            string
	Version       uint32
	CurrentHeight uint32
}

// 序列化为Digest
type RequestMessage struct {
	Timestamp  int64  `json:"timestamp"`
	ClientID   string `json:"clientID"` // 发送请求的NODE ID
	Block      string `json:"block"`    // 序列化后的Block
	SequenceID int64  `json:"sequenceID"`
}

type ReplyMessage struct {
	ViewID    int64  `json:"viewID"`
	Timestamp int64  `json:"timestamp"`
	ClientID  string `json:"clientID"`
	NodeID    string `json:"nodeID"`
	Result    string `json:"result"`
}

type PrePrepareMessage struct {
	ViewID     int64           `json:"viewID"`
	SequenceID int64           `json:"sequenceID"`
	Digest     string          `json:"digest"` // marshal of requestMsg is hashed
	RequestMsg *RequestMessage `json:"requestMsg"`
}

// prepareMsg
type VoteMessage struct {
	ViewID               int64           `json:"viewID"`
	SequenceID           int64           `json:"sequenceID"`
	Digest               string          `json:"digest"`
	NodeID               string          `json:"nodeID"`
	Resulte              bool            `json:"resulte"` // block是否正确
	RequestMsg           *RequestMessage `json:"requestMsg"`
	MessageConsensusType `json:"msgType"`
}

type MessageConsensusType int

const (
	PrepareMsg MessageConsensusType = iota
	CommitMsg
)

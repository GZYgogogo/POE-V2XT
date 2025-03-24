package network

import (
	"errors"
	"time"
)

type PBFTNode struct {
	NodeID   string // 节点ID
	View     *View  // 当前视图编号
	Sequence int64  // 当前序列号
	// 增加写锁
	CommittedMsgs []*RequestMessage
	CurrentState  *State
	// MsgBuffer    *MsgBuffer
}

type MsgBuffer struct {
	ReqMsgs        []*RequestMessage
	PrePrepareMsgs []*PrePrepareMessage
	PrepareMsgs    []*VoteMessage
	CommitMsgs     []*VoteMessage
}

type View struct {
	ID      int64
	Primary string
}

// type MsgBuffer struct {
// 	ReqMsgs        []*RequestMessage
// 	PrePrepareMsgs []*PrePrepareMessage
// 	PrepareMsgs    []*VoteMessage
// 	CommitMsgs     []*VoteMessage
// }

const ResolvingTimeDuration = time.Millisecond * 1000 // 1 second.

func InitPBFT(nodeId string) *PBFTNode {
	const viewID = 10000000000 // temporary.

	node := &PBFTNode{
		NodeID: nodeId,
		View: &View{
			ID:      viewID,
			Primary: "Main_NODE",
		},
		// Consensus-related struct
		CurrentState:  nil,
		CommittedMsgs: make([]*RequestMessage, 0),
		// MsgBuffer: &MsgBuffer{
		// 	ReqMsgs:        make([]*RequestMessage, 0),
		// 	PrePrepareMsgs: make([]*PrePrepareMessage, 0),
		// 	PrepareMsgs:    make([]*VoteMessage, 0),
		// 	CommitMsgs:     make([]*VoteMessage, 0),
		// },
	}

	return node
}

func (node *PBFTNode) GetReq(reqMsg *RequestMessage) (*PrePrepareMessage, error) {

	// Create a new state for the new consensus.
	err := node.createStateForNewConsensus()
	// fmt.Printf("%s state: %+v\n", node.NodeID, node.CurrentState)
	if err != nil {
		return nil, err
	}

	// Start the consensus process.
	prePrepareMsg, err := node.CurrentState.StartConsensus(reqMsg)
	if err != nil {
		return nil, err
	}
	node.Sequence = prePrepareMsg.SequenceID
	// Send getPrePrepare message
	if prePrepareMsg == nil {
		// node.Broadcast(prePrepareMsg, "/preprepare")
		// LogStage("Pre-prepare", true)
		return nil, errors.New("pre-prepare message is nil")
	}

	return prePrepareMsg, err
}

// 当节点的CurrentState为nil时，可以调用GetPrePrepare
func (node *PBFTNode) GetPrePrepare(prePrepareMsg *PrePrepareMessage) (*VoteMessage, error) {

	// Create a new state for the new consensus.
	err := node.createStateForNewConsensus()
	// fmt.Printf("%s state: %+v\n", node.NodeID, node.CurrentState)
	if err != nil {
		return nil, err
	}

	// PrePrepare验证消息是否一直，并且验证block有效性
	prePareMsg, err := node.CurrentState.PrePrepare(prePrepareMsg)

	if err != nil {
		return nil, err
	}
	if prePareMsg == nil {
		return nil, errors.New("prepare message is nil")
	}

	prePareMsg.NodeID = node.NodeID
	prePareMsg.RequestMsg = prePrepareMsg.RequestMsg
	// fmt.Printf("[prepare-Vote]: %+v\n", prePareMsg)
	node.Sequence = prePareMsg.SequenceID
	return prePareMsg, nil
}

func (node *PBFTNode) GetPrepare(prepareMsg *VoteMessage) (*VoteMessage, error) {
	commitMsg, err := node.CurrentState.Prepare(prepareMsg, node.NodeID)
	if err != nil {
		return nil, err
	}

	if commitMsg == nil {
		return nil, errors.New("commit message is nil")
	}
	commitMsg.NodeID = node.NodeID
	commitMsg.RequestMsg = prepareMsg.RequestMsg
	return commitMsg, nil
}

func (node *PBFTNode) GetCommit(commitMsg *VoteMessage) (*ReplyMessage, error) {
	// 已投票的会置CurrentState为nil
	if node.CurrentState == nil {
		return nil, errors.New("already committed")
	}
	replyMsg, committedMsg, err := node.CurrentState.Commit(commitMsg)
	// fmt.Printf("replyMessage => %+v; committedMsg => %+v; err => %v\n", replyMsg, committedMsg, err)
	if err != nil {
		return nil, err
	}

	if replyMsg != nil {
		if committedMsg != nil {
			// Attach node ID to the message
			// Save the last version of committed messages to node.
			node.CommittedMsgs = append(node.CommittedMsgs, committedMsg)
			// node.Reply(replyMsg)
			replyMsg.NodeID = node.NodeID
			return replyMsg, nil
		}
	}

	return nil, errors.New("committed message or the message is nil")
}

func (node *PBFTNode) createStateForNewConsensus() error {
	// Check if there is an ongoing consensus process.
	if node.CurrentState != nil {
		return errors.New("another consensus is ongoing")
	}

	// Get the last sequence ID
	var lastSequenceID int64
	if len(node.CommittedMsgs) == 0 {
		lastSequenceID = -1
	} else {
		lastSequenceID = node.CommittedMsgs[len(node.CommittedMsgs)-1].SequenceID
	}

	// Create a new state for this new consensus process in the Primary
	node.CurrentState = CreateState(node.View.ID, lastSequenceID)

	// LogStage("Create the replica status", true)

	return nil
}

// func LogMsg(msg interface{}) {
// 	switch msg.(type) {
// 	case *RequestMsg:
// 		reqMsg := msg.(*RequestMsg)
// 		fmt.Printf("[REQUEST] ClientID: %s, Timestamp: %d, Operation: %s\n", reqMsg.ClientID, reqMsg.Timestamp, reqMsg.Operation)
// 	case *consensus.PrePrepareMsg:
// 		prePrepareMsg := msg.(*consensus.PrePrepareMsg)
// 		fmt.Printf("[PREPREPARE] ClientID: %s, Operation: %s, SequenceID: %d\n", prePrepareMsg.RequestMsg.ClientID, prePrepareMsg.RequestMsg.Operation, prePrepareMsg.SequenceID)
// 	case *consensus.VoteMsg:
// 		voteMsg := msg.(*consensus.VoteMsg)
// 		if voteMsg.MsgType == consensus.PrepareMsg {
// 			fmt.Printf("[PREPARE] NodeID: %s\n", voteMsg.NodeID)
// 		} else if voteMsg.MsgType == consensus.CommitMsg {
// 			fmt.Printf("[COMMIT] NodeID: %s\n", voteMsg.NodeID)
// 		}
// 	}
// }

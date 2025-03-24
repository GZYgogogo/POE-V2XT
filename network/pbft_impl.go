package network

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"time"

	"github.com/anthdm/projectx/types"
)

type State struct {
	ViewID         int64
	MsgLogs        *MsgLogs
	LastSequenceID int64
	CurrentStage   int
}

type MsgLogs struct {
	ReqMsg      *RequestMessage
	PrepareMsgs map[string]*VoteMessage
	CommitMsgs  map[string]*VoteMessage
}

const (
	Idle        int = iota // Node is created successfully, but the consensus process is not started yet.
	PrePrepared            // The ReqMsgs is processed successfully. The node is ready to head to the Prepare stage.
	Prepared               // Same with `prepared` stage explained in the original paper.
	Committed              // Same with `committed-local` stage explained in the original paper.
)

// f: # 可以容忍的最大恶意节点数
// f = (n­1) / 3
// n = 4, in this case.
const f = 1

func CreateState(viewID int64, lastSequenceID int64) *State {
	return &State{
		ViewID: viewID,
		MsgLogs: &MsgLogs{
			ReqMsg:      nil,
			PrepareMsgs: make(map[string]*VoteMessage),
			CommitMsgs:  make(map[string]*VoteMessage),
		},
		LastSequenceID: lastSequenceID,
		CurrentStage:   Idle,
	}
}

func (state *State) StartConsensus(request *RequestMessage) (*PrePrepareMessage, error) {
	// `sequenceID` will be the index of this message.
	sequenceID := time.Now().UnixNano()

	// Find the unique and largest number for the sequence ID
	if state.LastSequenceID != -1 {
		for state.LastSequenceID >= sequenceID {
			sequenceID += 1
		}
	}

	// Assign a new sequence ID to the request message object.
	request.SequenceID = sequenceID
	// Save ReqMsgs to its logs.
	state.MsgLogs.ReqMsg = request
	// fmt.Printf("start consensus: %+v\n", request)
	// Get the digest of the request message
	digest, err := digest(request)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Change the stage to pre-prepared.
	state.CurrentStage = PrePrepared

	return &PrePrepareMessage{
		ViewID:     state.ViewID,
		SequenceID: sequenceID,
		Digest:     digest,
		RequestMsg: request,
	}, nil
}

func (state *State) PrePrepare(prePrepareMsg *PrePrepareMessage) (*VoteMessage, error) {

	// Get ReqMsgs and save it to its logs like the primary.
	state.MsgLogs.ReqMsg = prePrepareMsg.RequestMsg
	// fmt.Printf("state: %+v, request: %+v\n", state, prePrepareMsg.RequestMsg)
	// Verify if v, n(a.k.a. sequenceID), d are correct.
	if !state.verifyMsg(prePrepareMsg.ViewID, prePrepareMsg.SequenceID, prePrepareMsg.Digest) {
		return nil, errors.New("pre-prepare message is corrupted")
	}

	// Change the stage to pre-prepared.
	state.CurrentStage = PrePrepared

	return &VoteMessage{
		ViewID:               state.ViewID,
		SequenceID:           prePrepareMsg.SequenceID,
		Digest:               prePrepareMsg.Digest,
		MessageConsensusType: PrepareMsg,
	}, nil
}

func (state *State) Prepare(prepareMsg *VoteMessage, nodeID string) (*VoteMessage, error) {
	if !state.verifyMsg(prepareMsg.ViewID, prepareMsg.SequenceID, prepareMsg.Digest) {
		return nil, errors.New("prepare message is corrupted")
	}

	// Append msg to its logs
	state.MsgLogs.PrepareMsgs[prepareMsg.NodeID] = prepareMsg

	if state.prepared() {
		// Change the stage to prepared.
		state.CurrentStage = Prepared

		return &VoteMessage{
			ViewID:               state.ViewID,
			SequenceID:           prepareMsg.SequenceID,
			Digest:               prepareMsg.Digest,
			MessageConsensusType: CommitMsg,
		}, nil
	}
	return nil, nil
}

func (state *State) Commit(commitMsg *VoteMessage) (*ReplyMessage, *RequestMessage, error) {
	if !state.verifyMsg(commitMsg.ViewID, commitMsg.SequenceID, commitMsg.Digest) {
		return nil, nil, errors.New("commit message is corrupted")
	}

	// Append msg to its logs
	state.MsgLogs.CommitMsgs[commitMsg.NodeID] = commitMsg

	// Print current voting status
	// fmt.Printf("[Commit-Vote]: %d\n", len(state.MsgLogs.CommitMsgs))

	if state.committed() {
		// This node executes the requested operation locally and gets the result.
		result := "Executed"

		// Change the stage to prepared.
		state.CurrentStage = Committed

		return &ReplyMessage{
			ViewID:    state.ViewID,
			Timestamp: state.MsgLogs.ReqMsg.Timestamp,
			ClientID:  state.MsgLogs.ReqMsg.ClientID,
			Result:    result,
		}, state.MsgLogs.ReqMsg, nil
	}

	return nil, nil, fmt.Errorf("Not enough commit messages.")
}

func (state *State) prepared() bool {
	if state.MsgLogs.ReqMsg == nil {
		return false
	}

	if len(state.MsgLogs.PrepareMsgs) < 2*f {
		return false
	}

	return true
}

func (state *State) committed() bool {
	if !state.prepared() {
		return false
	}

	if len(state.MsgLogs.CommitMsgs) < 2*f {
		return false
	}

	return true
}

func (state *State) verifyMsg(viewID int64, sequenceID int64, digestGot string) bool {
	// Wrong view. That is, wrong configurations of peers to start the consensus.
	if state.ViewID != viewID {
		return false
	}

	// Check if the Primary sent fault sequence number. => Faulty primary.
	// TODO: adopt upper/lower bound check.
	if state.LastSequenceID != -1 {
		if state.LastSequenceID >= sequenceID {
			return false
		}
	}
	// fmt.Printf("state.MsgLogs.ReqMsg: %+v\n", state.MsgLogs.ReqMsg)
	digest, err := digest(state.MsgLogs.ReqMsg)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Check digest.
	if digestGot != digest {
		return false
	}

	return true
}

func digest(object interface{}) (string, error) {
	msg, err := json.Marshal(object)

	if err != nil {
		return "", err
	}

	return Hash(msg).String(), nil
}

func Hash(data []byte) types.Hash {
	return types.Hash(sha256.Sum256(data))
}

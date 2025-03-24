package network

import (
	"testing"
)

func TestStartPBFT(t *testing.T) {
	// 初始化区块链
	// bc := &core.Blockchain{
	// 	BlockStore: make(map[types.Hash]*core.Block),
	// 	TxStore:    make(map[types.Hash]*core.Transaction),
	// }

	// // 创建PBFT节点
	// node := &PBFTNode{
	// 	ID:         1,
	// 	Blockchain: bc,
	// 	Logs:       make(map[int64]*ConsensusLog),
	// 	// msgStore:   make(map[MsgType]map[int64]map[int]*Message),
	// }

	// // 创建测试区块
	// header := &core.Header{
	// 	PrevBlockHash: types.Hash{},
	// 	Timestamp:     time.Now().Unix(),
	// 	Height:        1,
	// }
	// block := &core.Block{
	// 	Header: header,
	// 	// Hash:   header.CalculateHash(),
	// }

	// // 启动共识流程
	// seq := node.NextSequence()
	// // 日志记录block共识
	// log := node.GetLog(seq)
	// log.Block = block

	// // 模拟消息处理
	// // prepareMsg := &Message{
	// // 	Type:      pbft.MsgPrepare,
	// // 	View:      node.View,
	// // 	Sequence:  seq,
	// // 	BlockHash: block.Hash(core.BlockHasher{}),
	// // 	NodeID:    node.ID,
	// // }
	// // node.HandleMessage(prepareMsg)

	// // 创建检查点
	// node.CreateCheckpoint(seq)
}

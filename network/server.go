package network

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"os"

	"encoding/json"

	"github.com/anthdm/projectx/core"
	"github.com/anthdm/projectx/crypto"
	"github.com/anthdm/projectx/types"

	"time"

	"github.com/go-kit/log"
)

var defaultBlockTime = time.Millisecond * 800

// var defaultBlockTime = time.Second * 1

// 信誉三元组
type TrustOpinion struct {
	Belief      float64 // 信任度 b
	Disbelief   float64 // 不信任度 d
	Uncertainty float64 // 不确定度 u
}

type ServerOpts struct {
	ID            string
	Transport     Transport
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transports    []Transport
	PrivateKey    *crypto.PrivateKey
	blockTime     time.Duration
	Reputation    float64 // 信誉值
	Trust         TrustOpinion
}

type Server struct {
	ServerOpts
	blockTime   time.Duration
	memPool     *TxPool
	blockchain  *core.Blockchain
	isVaildator bool //是验证器节点还是普通节点
	rpcCh       chan RPC
	txChan      chan *core.Transaction
	quitCh      chan struct{} //退出信号

	PBFTNode *PBFTNode
}

func NewServer(opts ServerOpts) (*Server, error) {
	if opts.blockTime == time.Duration(0) {
		opts.blockTime = defaultBlockTime
	}
	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
	}
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "ID", opts.ID)
	}
	blockchain, err := core.NewBlockchain(opts.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}
	s := &Server{
		ServerOpts:  opts,
		isVaildator: opts.PrivateKey != nil,
		blockchain:  blockchain,
		blockTime:   opts.blockTime,
		memPool:     NewTxPool(100),
		rpcCh:       make(chan RPC), //无缓冲
		txChan:      make(chan *core.Transaction),
		quitCh:      make(chan struct{}, 1),

		PBFTNode: InitPBFT(opts.ID),
	}

	//If we dont got any processor from server options, we going to
	//use the server as default
	if s.RPCProcessor == nil {
		s.RPCProcessor = s
	}
	go s.boostrapNodes()
	if s.isVaildator {
		go s.VaildatorLoop()
	}
	s.initTransports()
	return s, nil
}

func (s *Server) Start() {
	// the channel of all transports connected to server are listened
	time.Sleep(time.Second * 1)
	// go s.syncLoop()
free:
	//一直检查Server中来往的rpc，如果没有判断是否要退出
	for {
		select {
		case tx := <-s.txChan:
			if err := s.processTransaction(tx); err != nil {
				s.Logger.Log("process TX error", err)
			}
		case rpc := <-s.rpcCh:
			// handle rpc message
			// type of msg is *DecodedMessage
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("error", err)
			}
			// process decoded message
			if err = s.RPCProcessor.ProcessMessage(msg); err != nil {
				if err != core.ErrBlockKnown {
					s.Logger.Log("error", err)
				}
			}

		// quit signal
		case <-s.quitCh:
			break free
		}
	}
	fmt.Println("Server showdown")
}

func (s *Server) VaildatorLoop() {
	//创建一个定时器
	ticker := time.NewTicker(s.blockTime)

	s.Logger.Log(
		"msg", "starting vaildator loop",
		"blockTime", s.blockTime,
	)

	for {
		<-ticker.C
		if err := s.createNewBlock(); err != nil {
			s.Logger.Log("error", err)
		}
	}
}

// server connect to all nodes in transports list
func (s *Server) boostrapNodes() error {
	for _, tr := range s.Transports {
		if s.Transport.Addr() != tr.Addr() {
			if err := s.Transport.Connect(tr); err != nil {
				s.Logger.Log("error", "couldn't connect to remote", err)
				return err
			}
		}
	}
	return nil
}

// 同步循环，保证节点之间的区块一致
func (s *Server) syncLoop() {
	ticker := time.NewTicker(1 * time.Second)

	for {
		<-ticker.C
		for _, tr := range s.Transports {
			if s.Transport.Addr() != tr.Addr() {
				if err := s.sendGetStatusMessage(tr); err != nil { // && tr.CheckConnection(s.Transport.Addr())
					s.Logger.Log("error", "sendGetStatusMessage", err)
				}
			}
		}
	}
}

// 传输层！！！！
// Server come true Processor interface
// 函数的参数是再复制一份数据。参数使用pointer是防止传入的参数数据过大。
func (s *Server) ProcessMessage(msg *DecodedMessage) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	case *core.Block:
		return s.processBlock(t)
	case *GetStatusMessage:
		return s.processGetStatusMessage(msg.From, t)
	case *StatusMessage:
		return s.processStatusMessage(msg.From, t)
	case *GetBlocksMessage:
		return s.processGetBlocksMessage(msg.From, t)
	case *BlocksMessage:
		return s.processBlocksMessage(msg.From, t)
	case *ReplyMessage:
		return s.processReplyMessage(msg.From, t)
	// case *RequestMessage:
	// 	return s.processRequestMessage(msg.From, t)
	case *PrePrepareMessage:
		return s.processPrePrepareMessage(msg.From, t)
	case *VoteMessage:
		return s.processVoteMessage(msg.From, t)
	}

	return nil
}

// func (s *Server) processRequestMessage(from string, data *RequestMessage) error {
// 	if s.Transport.Addr() != "Main_NODE" {
// 		return nil
// 	}
// 	// fmt.Printf("%s receive request message from %s\n", s.Transport.Addr(), from)
// 	// panic("here")
// 	prePrepare, err := s.PBFTNode.GetReq(data)
// 	if err != nil {
// 		return err
// 	}
// 	// 广播prePrepare消息
// 	buf := new(bytes.Buffer)
// 	if err := gob.NewEncoder(buf).Encode(prePrepare); err != nil {
// 		return err
// 	}
// 	msg := NewMessage(MessageTypePrePrepare, buf.Bytes())
// 	// fmt.Printf("%s broadcast preprepare message to %+v\n", s.Transport.Addr(), s.Transports)
// 	return s.Transport.Broadcast(msg.Bytes())
// }

func (s *Server) processPrePrepareMessage(from string, data *PrePrepareMessage) error {
	// fmt.Printf("%s receive preprepare message from %s\n", s.Transport.Addr(), from)

	// TODO：GetPrePrepare内验证block的合法性，不合法不同意，就不要继续发送了
	Block := new(core.Block)
	err := json.Unmarshal([]byte(data.RequestMsg.Block), Block)
	if err != nil {
		return err
	}
	if err := s.blockchain.Validator.ValidateBlock(Block); err != nil {
		return err
	}
	voteMessage, err := s.PBFTNode.GetPrePrepare(data)

	if err != nil {
		return err
	}
	// 广播Prepare的投票消息
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(voteMessage); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeVote, buf.Bytes())
	return s.Transport.Broadcast(msg.Bytes())
}

func (s *Server) processVoteMessage(from string, data *VoteMessage) error {

	buf := new(bytes.Buffer)
	var msg *Message
	switch data.MessageConsensusType {
	case PrepareMsg:
		//全部正确投票
		// fmt.Printf("%s receive vote message from %s => %+v\n", s.Transport.Addr(), from, data)
		// 收集到足够多的投票才能生成commit消息
		commitMessage, err := s.PBFTNode.GetPrepare(data)
		// fmt.Printf("commitMessage => %+v\n", commitMessage)
		if err != nil {
			return err
		}

		// 广播commit消息
		if err := gob.NewEncoder(buf).Encode(commitMessage); err != nil {
			return err
		}
		msg = NewMessage(MessageTypeVote, buf.Bytes())
		return s.Transport.Broadcast(msg.Bytes())
	case CommitMsg:
		/// 收集到足够多的投票才能生成commit消息
		// fmt.Printf("%s receive CommitMsg vote from %s\n", s.Transport.Addr(), from)
		replyMessage, err := s.PBFTNode.GetCommit(data)

		if err != nil {
			return err
		}

		// fmt.Printf("replyMessage:%+v\n", replyMessage)
		if err := gob.NewEncoder(buf).Encode(replyMessage); err != nil {
			return err
		}
		msg = NewMessage(MessageTypeReply, buf.Bytes())
		//TODO: 发送reply消息之前就可以执行tx，并上链
		block := new(core.Block)
		err = json.Unmarshal([]byte(data.RequestMsg.Block), block)
		if err != nil {
			return err
		}
		err = s.blockchain.AddBlock(block)
		if err != nil {
			return err
		}
		// 清理共识的CurrentState，为下一轮共识做准备
		// fmt.Printf("%s PBFTNode : %+v\n", s.Transport.Addr(), s.PBFTNode.CurrentState.LastSequenceID)
		s.PBFTNode.CurrentState = nil

		// fmt.Printf("%s PBFTNode: %+v,Sequence %+v\n", s.Transport.Addr(), s.PBFTNode, s.PBFTNode.Sequence)
		return s.Transport.SendMessage("LOCAL", msg.Bytes())
	default:
		return errors.New("unknown message consensus type")
	}
}

// 告诉主节点已经执行完毕
func (s *Server) processReplyMessage(from string, data *ReplyMessage) error {
	// fmt.Printf("%s receive reply message from %s => block had added to chain\n", s.Transport.Addr(), from)
	// s.Logger.Log("msg", "received reply message", "from", from)
	// fmt.Println("receive reply message from ", from)
	// TODO: 重新开始一轮共识，删除本轮的CurrentState
	// s.PBFTNode.CurrentState = &State{}
	return nil
}

// send status message, if height lower than current height, we need to get block information
// TODO: Remove the logic from the main function to here
// Normally Transport which is our own transport should do the trick.
func (s *Server) sendGetStatusMessage(tr Transport) error {
	var (
		getStatusMsg = new(GetStatusMessage)
		buf          = new(bytes.Buffer)
	)

	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())

	// fmt.Printf("=> %s sending GetStatus message to %s\n", s.Transport.Addr(), tr.Addr())
	if err := s.Transport.SendMessage(tr.Addr(), msg.Bytes()); err != nil {
		return err
	}

	//statusMessage
	return nil
}

func (s *Server) processTransaction(tx *core.Transaction) error {
	// hash := tx.Hash(core.TxHasher{})
	//验证交易
	if err := tx.Verify(); err != nil {
		return err
	}
	//将交易是否已经加入内存池中，为了课题展现快速共识，暂时注释
	// if s.memPool.Contains(hash) {
	// 	return nil
	// }
	tx.SetFirstSceen(time.Now().Unix())

	//为效率，现在有专门的节点负责广播交易
	// go s.BroadcastTx(tx)

	s.memPool.Add(tx)

	return nil
}

func (s *Server) processBlock(b *core.Block) error {
	if err := s.blockchain.AddBlock(b); err != nil {
		return err
	}
	go s.BroadcastBlock(b)

	return nil
}

func (s *Server) processGetStatusMessage(form string, data *GetStatusMessage) error {
	// fmt.Printf("=> %s receiving GetStatus message from %s => %+v\n", s.Transport.Addr(), form, data)
	StatusMessage := &StatusMessage{
		ID:            s.ID,
		CurrentHeight: s.blockchain.Height(),
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(StatusMessage); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeStatus, buf.Bytes())
	return s.Transport.SendMessage(form, msg.Bytes())
}

func (s *Server) processStatusMessage(form string, data *StatusMessage) error {
	fmt.Printf("=> %s receiving Status message from %s => %+v\n", s.Transport.Addr(), form, data)
	if s.blockchain.Height() >= data.CurrentHeight {
		s.Logger.Log("msg", "cannot sync blockHeight to low", "ourHeight", s.blockchain.Height(), "remoteHeight", data.CurrentHeight)
		return nil
	}

	//TODO: in this case, we 100% sure that the node has blocks heighter than us
	ourHeight := s.blockchain.Height()
	getBlockMessage := &GetBlocksMessage{
		From: ourHeight + 1,
		To:   0,
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(getBlockMessage); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())

	return s.Transport.SendMessage(form, msg.Bytes())
}

func (s *Server) processGetBlocksMessage(from string, data *GetBlocksMessage) error {

	s.Logger.Log("msg", "received getBlocks message", "from", from)

	var (
		blocks    = []*core.Block{}
		ourHeight = s.blockchain.Height()
	)

	if data.To == 0 {
		for i := int(data.From); i <= int(ourHeight); i++ {
			block, err := s.blockchain.GetBlock(uint32(i))
			if err != nil {
				return err
			}

			blocks = append(blocks, block)
		}
	}

	blocksMsg := &BlocksMessage{
		Blocks: blocks,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(blocksMsg); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeBlocks, buf.Bytes())

	return s.Transport.SendMessage(from, msg.Bytes())
}

func (s *Server) processBlocksMessage(from string, data *BlocksMessage) error {
	for _, block := range data.Blocks {
		if err := s.blockchain.AddBlock(block); err != nil {
			s.Logger.Log("error", err.Error())
			return err
		}
	}
	return nil
}

// 获取消息bug，获取到了不属于自己的消息
func (s *Server) initTransports() {
	go func() {
		for rpc := range s.Transport.Consume() {
			s.rpcCh <- rpc
		}
	}()
	go func() {
		for tx := range s.Transport.ConsumeTx() {
			s.txChan <- tx
		}
	}()
}

func (s *Server) Broadcast(payload []byte) error {
	for _, tr := range s.Transports {
		if err := tr.Broadcast(payload); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) BroadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(b); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeBlock, buf.Bytes())
	return s.Broadcast(msg.Bytes())
}

func (s *Server) BroadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(tx); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeTx, buf.Bytes())
	return s.Broadcast(msg.Bytes())
}

func (s *Server) createNewBlock() error {
	currentHeader, err := s.blockchain.GetHeader(s.blockchain.Height())
	if err != nil {
		return err
	}

	txx := s.memPool.Pending()

	block, err := core.NewBlockFromPrevHeader(currentHeader, txx)
	if err != nil {
		return err
	}
	if err := block.Sign(*s.PrivateKey); err != nil {
		return err
	}
	if err := s.blockchain.AddBlock(block); err != nil {
		return err
	}

	//TODO:pending pool of tx should only reflect on vaildator nodes,
	//Right now "normal nodes" does not have their pending pool.
	s.memPool.ClearPending()
	// go s.BroadcastBlock(block)
	go func(error) {
		err = s.StartConsensus(block)
		// s.Logger.Log("error", err)
		// return
	}(err)
	return err
}

func (s *Server) StartConsensus(block *core.Block) error {
	s.PBFTNode.CurrentState = nil
	jsonBlock, err := json.Marshal(block)
	if err != nil {
		return err
	}
	requestMsg := &RequestMessage{
		Timestamp: time.Now().Unix(),
		Block:     string(jsonBlock),
	}

	prePrepare, err := s.PBFTNode.GetReq(requestMsg)
	if err != nil {
		return err
	}
	// 广播prePrepare消息
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(prePrepare); err != nil {
		return err
	}

	msg := NewMessage(MessageTypePrePrepare, buf.Bytes())
	return s.Transport.Broadcast(msg.Bytes())
}

func genesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Timestamp: 0000000,
		Height:    0,
	}
	block, _ := core.NewBlock(header, nil)
	return block
}

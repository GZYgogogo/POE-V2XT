package core

import (
	"fmt"
	"math"

	"github.com/anthdm/projectx/crypto"
	"github.com/anthdm/projectx/types"
)

type Transaction struct {
	Data      []byte
	From      crypto.PublicKey
	Signature *crypto.Signature

	// cached version of the tx data hash
	hash       types.Hash
	firstSceen int64

	IsUrgent bool    // 是否为紧急交易，紧急交易才有Poe
	Tc       float64 // Tc 表示紧急交易链确认一个区块的平均时延，
	Te       float64 // Te 表示交易的期待时延
	Ts       float64 // Ts 表示交易的产生时间
	Ta       float64 // Ta 表示交易到达RSU的时间。
	Poe      float64
}

func (tx *Transaction) SetFirstSceen(t int64) {
	tx.firstSceen = t
}

func (tx *Transaction) FirstSceen() int64 {
	return tx.firstSceen
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data: data,
	}
}

func (tx *Transaction) Hash(hasher Hasher[*Transaction]) types.Hash {
	if tx.hash.IsZero() {
		tx.hash = hasher.Hash(tx)
	}
	return tx.hash
}

func (tx *Transaction) Sign(privKey crypto.PrivateKey) error {
	sig, err := privKey.Sign(tx.Data)
	if err != nil {
		return err
	}

	tx.From = privKey.PublicKey()
	tx.Signature = sig

	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("transaction has no signature")
	}

	if !tx.Signature.Verify(tx.From, tx.Data) {
		return fmt.Errorf("invalid transaction signature")
	}

	return nil
}

func (tx *Transaction) Decode(dec Decoder[*Transaction]) error {
	return dec.Decode(tx)
}

func (tx *Transaction) Encode(enc Encoder[*Transaction]) error {
	return enc.Encode(tx)
}

func (t *Transaction) CalculatePoe() {
	// ω 代表车辆在此期间已经累计申请的紧急交易数量，设置为一轮共识中提交的紧急交易数量
	// θ 代表已申请紧急交易数量的影响权重 0.4
	w := 1.0
	θ := 0.4
	t.Poe = t.CalculateE() * math.Exp(w*θ)
}

// E 表示车辆对该笔交易紧急度的期望
func (t *Transaction) CalculateE() float64 {
	return math.Exp(t.Tc / (t.Te + t.Ts - t.Ta))
}

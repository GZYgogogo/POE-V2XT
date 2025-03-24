package network

import "github.com/anthdm/projectx/core"

type Transport interface {
	Consume() <-chan RPC
	Connect(Transport) error
	SendMessage(string, []byte) error
	SendTransaction(*core.Transaction) error
	ConsumeTx() <-chan *core.Transaction
	Broadcast([]byte) error
	Addr() string
}

package network

type Transport interface {
	Consume() <-chan RPC
	Connect(Transport) error
	SendMessage(string, []byte) error
	Broadcast([]byte) error
	Addr() string
}

package network

import (
	"bytes"
	"fmt"
	"sync"
)

type LocalTransport struct {
	addr      string
	consumeCh chan RPC
	lock      sync.RWMutex
	peers     map[string]*LocalTransport
}

func NewLocalTransport(addr string) *LocalTransport {
	return &LocalTransport{
		addr:      addr,
		consumeCh: make(chan RPC, 1024),
		peers:     make(map[string]*LocalTransport),
	}
}

func (t *LocalTransport) Consume() <-chan RPC {
	return t.consumeCh
}

func (t *LocalTransport) Connect(tr Transport) error {
	trans := tr.(*LocalTransport)
	t.lock.Lock()
	defer t.lock.Unlock()

	t.peers[tr.Addr()] = trans

	return nil
}

func (t *LocalTransport) SendMessage(to string, payload []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if t.addr == to {
		// return fmt.Errorf("%s: could not send message to itself %s", t.addr, to)
		return nil
	}

	peer, ok := t.peers[to]
	if !ok {
		return fmt.Errorf("%s: could not send message to unknown peer %s", t.addr, to)
	}
	// fmt.Printf("%s 的 peers: %+v\n", t.Addr(), t.peers)
	// fmt.Printf("get peer %+v\n", peer)
	// fmt.Printf("%s send message  to %s\n", t.Addr(), to)
	peer.consumeCh <- RPC{
		From:    t.addr,
		Payload: bytes.NewReader(payload),
	}
	// fmt.Println("send message done", len(peer.consumeCh))
	return nil
}

func (t *LocalTransport) Broadcast(payload []byte) error {
	for _, peer := range t.peers {
		// fmt.Printf("send message from %s to %s\n", t.Addr(), peer.Addr())
		if err := t.SendMessage(peer.Addr(), payload); err != nil {
			return err
		}
	}
	return nil
}

func (t *LocalTransport) Addr() string {
	return t.addr
}

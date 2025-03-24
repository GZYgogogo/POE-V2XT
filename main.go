// package main

// import (
// 	"bytes"
// 	"encoding/gob"
// 	"fmt"
// 	"log"

// 	"github.com/anthdm/projectx/core"
// 	"github.com/anthdm/projectx/crypto"
// 	"github.com/anthdm/projectx/network"
// )

// var transports = []network.Transport{
// 	network.NewLocalTransport("LOCAL"),
// 	network.NewLocalTransport("REMOTE_1"),
// 	network.NewLocalTransport("REMOTE_2"),
// 	network.NewLocalTransport("REMOTE_3"),
// }

// var remoteTransports = []network.Transport{
// 	network.NewLocalTransport("REMOTE_1"),
// 	network.NewLocalTransport("REMOTE_2"),
// 	network.NewLocalTransport("REMOTE_3"),
// }

// func main() {
// 	initRemoteServers(remoteTransports)
// 	localNode := transports[0]
// 	// trLate := transports[1]
// 	// remoteNodeA := transports[1]
// 	// remoteNodeC := transports[3]

// 	// go func() {
// 	// 	for {
// 	// 		if err := sendTransaction(remoteNodeA, localNode.Addr()); err != nil {
// 	// 			logrus.Error(err)
// 	// 		}
// 	// 		time.Sleep(2 * time.Second)
// 	// 	}
// 	// }()

// 	// go func() {
// 	// 	time.Sleep(6 * time.Second)
// 	// 	lateServer := makeServer(string(trLate.Addr()), trLate, nil)
// 	// 	go lateServer.Start()
// 	// }()

// 	privKey := crypto.GeneratePrivateKey()
// 	localServer := makeServer("LOCAL", localNode, &privKey)
// 	localServer.Start()
// }

// func initRemoteServers(trs []network.Transport) {
// 	for i := 0; i < len(trs); i++ {
// 		id := fmt.Sprintf("REMOTE_%d", i)
// 		s := makeServer(id, trs[i], nil)
// 		go s.Start()
// 	}
// }

// func makeServer(id string, tr network.Transport, pk *crypto.PrivateKey) *network.Server {
// 	// newTransport := []network.Transport{}
// 	// for _, transport := range transports {
// 	// 	if transport.Addr() != tr.Addr() {
// 	// 		// fmt.Printf("%s 加入至 %s \n", transport.Addr(), tr.Addr())
// 	// 		newTransport = append(newTransport, transport)
// 	// 	}
// 	// }
// 	opts := network.ServerOpts{
// 		Transport:  tr,
// 		PrivateKey: pk,
// 		ID:         id,
// 		Transports: transports,
// 	}

// 	s, err := network.NewServer(opts)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	return s
// }

// func sendGetStatusMessage(tr network.Transport, to network.NetAddr) error {
// 	var (
// 		getStatusMsg = new(network.GetStatusMessage)
// 		buf          = new(bytes.Buffer)
// 	)

// 	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
// 		return err
// 	}
// 	msg := network.NewMessage(network.MessageTypeGetStatus, buf.Bytes())

// 	return tr.SendMessage(to, msg.Bytes())
// }

// func sendTransaction(tr network.Transport, to network.NetAddr) error {
// 	privKey := crypto.GeneratePrivateKey()
// 	// data := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0e}
// 	data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d, 0x05, 0x0a, 0x0f}
// 	tx := core.NewTransaction(data)
// 	tx.Sign(privKey)
// 	buf := &bytes.Buffer{}
// 	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
// 		return err
// 	}

// 	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

// 	return tr.SendMessage(to, msg.Bytes())
// }

package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/anthdm/projectx/cfg"
	"github.com/anthdm/projectx/core"
	"github.com/anthdm/projectx/crypto"
	"github.com/anthdm/projectx/network"
	"github.com/sirupsen/logrus"
)

var transports = []network.Transport{
	network.NewLocalTransport("LOCAL"),
	// network.NewLocalTransport("Main_NODE"),
	network.NewLocalTransport("REMOTE_0"),
	network.NewLocalTransport("REMOTE_1"),
	network.NewLocalTransport("REMOTE_2"),
	// network.NewLocalTransport("REMOTE_3"),
	// network.NewLocalTransport("REMOTE_4"),
	// network.NewLocalTransport("REMOTE_5"),
	// network.NewLocalTransport("REMOTE_6"),
	// network.NewLocalTransport("REMOTE_7"),
	// network.NewLocalTransport("REMOTE_8"),
	// network.NewLocalTransport("REMOTE_9"),
	// network.NewLocalTransport("REMOTE_10"),
	// network.NewLocalTransport("REMOTE_11"),
	// network.NewLocalTransport("REMOTE_12"),
}

func main() {
	cfg.Init()
	localNode := transports[0]
	// main := transports[1]
	remoteNode0 := transports[1]
	// remoteNodeC := transports[3]

	//循环发送交易
	go func() {
		for {
			if err := sendTransaction(remoteNode0, localNode.Addr()); err != nil {
				logrus.Error(err)
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()
	initRemoteServers(transports[1:])
	privKey := crypto.GeneratePrivateKey()
	localServer := makeServer("LOCAL", localNode, &privKey)
	go localServer.Start()
	// mainServer := makeServer("Main_NODE", main, nil)
	// go mainServer.Start()

	select {}
}

func initRemoteServers(trs []network.Transport) {
	for i := 0; i < len(trs); i++ {
		id := fmt.Sprintf("REMOTE_%d", i)
		s := makeServer(id, trs[i], nil)
		go s.Start()
	}
}

// func makeClientServer(id string, tr network.Transport, pk *crypto.PrivateKey) *network.Server {
// 	newTransport := []network.Transport{}
// 	for _, transport := range transports {
// 		if transport.Addr() != tr.Addr() && transport.Addr() != "LOCAL" {
// 			newTransport = append(newTransport, transport)
// 		}
// 	}
// 	opts := network.ServerOpts{
// 		Transport:  tr,
// 		PrivateKey: pk,
// 		ID:         id,
// 		Transports: newTransport,
// 	}

// 	s, err := network.NewServer(opts)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	return s
// }

func makeServer(id string, tr network.Transport, pk *crypto.PrivateKey) *network.Server {
	newTransport := []network.Transport{}
	for _, transport := range transports {
		if transport.Addr() != tr.Addr() {
			newTransport = append(newTransport, transport)
		}
	}
	opts := network.ServerOpts{
		Transport:  tr,
		PrivateKey: pk,
		ID:         id,
		Transports: newTransport,
	}

	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	return s
}

func sendGetStatusMessage(tr network.Transport, to string) error {
	var (
		getStatusMsg = new(network.GetStatusMessage)
		buf          = new(bytes.Buffer)
	)

	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}
	msg := network.NewMessage(network.MessageTypeGetStatus, buf.Bytes())

	return tr.SendMessage(to, msg.Bytes())
}

func sendTransaction(tr network.Transport, to string) error {
	privKey := crypto.GeneratePrivateKey()

	// data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d, 0x05, 0x0a, 0x0f}

	// 定义要生成的字节长度
	// length := 16

	// // 生成随机字节切片
	// randomBytes := make([]byte, length)
	// _, err := rand.Read(randomBytes)
	// if err != nil {
	// 	return err
	// }
	// tx := core.NewTransaction(randomBytes)
	tx := core.NewTransaction([]byte{0x01})
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	return tr.SendMessage(to, msg.Bytes())
}

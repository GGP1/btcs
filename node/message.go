package node

import (
	"github.com/GGP1/btcs/encoding/gob"
)

const (
	messageLength = 12

	// https://developer.bitcoin.org/reference/p2p_networking.html
	msgAddr      message = "addr"
	msgBlock     message = "block"
	msgGetAddr   message = "getaddr"
	msgGetBlocks message = "getblocks"
	msgGetData   message = "getdata"
	msgInv       message = "inv"
	msgPing      message = "ping"
	msgPong      message = "pong"
	msgTx        message = "tx"
	msgVersion   message = "version"
)

type (
	message string

	addr struct {
		Addresses []string
	}

	blockData struct {
		AddrFrom string
		Block    []byte
	}

	getaddr struct {
		AddrFrom string
	}

	getblocks struct {
		AddrFrom string
	}

	getdata struct {
		AddrFrom string
		Type     string
		ID       []byte
	}

	inv struct {
		AddrFrom string
		Type     string
		Items    [][]byte
	}

	ping struct {
		AddrFrom string
	}

	pong struct {
		AddrFrom string
	}

	transaction struct {
		AddrFrom    string
		Transaction []byte
	}

	version struct {
		AddrFrom   string
		Version    int
		BestHeight int32
	}
)

func newMessage(cmd message, payload any) ([]byte, error) {
	encPayload, err := gob.Encode(payload)
	if err != nil {
		return nil, err
	}

	var bytes [messageLength]byte
	for i, c := range cmd {
		bytes[i] = byte(c)
	}

	return append(bytes[:], encPayload...), nil
}

func bytesToMessage(bytes []byte) message {
	cmd := make([]byte, 0, messageLength)

	for _, b := range bytes {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}

	return message(cmd)
}

func getPayload[T any](request []byte) (T, error) {
	return gob.Decode[T](request[messageLength:])
}

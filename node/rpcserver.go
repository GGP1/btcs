package node

import (
	"net"
	"net/rpc"
	"os"

	"github.com/GGP1/btcs/block"
	"github.com/GGP1/btcs/encoding/base58"
	"github.com/GGP1/btcs/logger"
	"github.com/GGP1/btcs/tx"
	"github.com/GGP1/btcs/tx/utxo"
	"github.com/GGP1/btcs/wallet"
)

// RPCAddress where the node will be listening for rpc calls.
const RPCAddress = "0.0.0.0:8338"

// GetTransactionResponse is the structure of the GetTransaction rpc call response.
type GetTransactionResponse struct {
	Tx    tx.Tx
	Block block.Block
}

// SendTxParams contains the parameters used for the SendTx rpc call.
type SendTxParams struct {
	AccountName string
	To          string
	Amount      int
	Fee         int
}

// RunRPCServer starts the node's rpc server.
//
// The listener is returned to call Close when done.
func (n *Node) RunRPCServer() (net.Listener, error) {
	ln, err := net.Listen("tcp", RPCAddress)
	if err != nil {
		return nil, err
	}

	server := rpc.NewServer()
	if err := server.Register(n); err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-n.interrupt:
				return

			default:
				conn, err := ln.Accept()
				if err != nil {
					logger.Error("RPC: ", err)
					continue
				}
				go server.ServeConn(conn)
			}
		}
	}()
	logger.Info("Starting RPC server at ", RPCAddress)

	return ln, nil
}

// AddNode adds a node to the peer list.
func (n *Node) AddNode(address string, reply *int) error {
	*reply = n.peers.Add(address)
	return n.sendVersion(address)
}

// DisconnectNode removes a node from the peer list.
func (n *Node) DisconnectNode(address string, reply *int) error {
	*reply = n.peers.Remove(address)
	return nil
}

// GetRawMempool returns the node's mempool transaction ids.
func (n *Node) GetRawMempool(_ struct{}, reply *[]string) error {
	txIDs := make([]string, 0, n.txPool.Count())
	err := n.txPool.ForEach(func(txID string, _ tx.Tx) error {
		txIDs = append(txIDs, txID)
		return nil
	})
	if err != nil {
		return err
	}

	*reply = txIDs
	return nil
}

// GetBestHeight returns the node's blockchain best height.
func (n *Node) GetBestHeight(_ struct{}, reply *int32) error {
	bestHeight, err := n.blockchain.BestHeight()
	if err != nil {
		return err
	}
	*reply = bestHeight
	return nil
}

// GetBlock returns a block given a hash.
func (n *Node) GetBlock(hash []byte, reply *block.Block) error {
	block, err := n.blockchain.Block(hash)
	if err != nil {
		return err
	}
	*reply = block
	return nil
}

// GetLastBlock returns the last block (tip) of a chain.
func (n *Node) GetLastBlock(_ struct{}, reply *block.Block) error {
	block, err := n.blockchain.LastBlock()
	if err != nil {
		return err
	}
	*reply = block
	return nil
}

// GetPeerInfo returns data about each connected node.
func (n *Node) GetPeerInfo(_ struct{}, reply *[]string) error {
	*reply = n.peers.List()
	return nil
}

// GetTransaction returns a transaction given an id.
func (n *Node) GetTransaction(id []byte, reply *GetTransactionResponse) error {
	block, tx, err := n.blockchain.FindTransaction(id)
	if err != nil {
		return err
	}
	*reply = GetTransactionResponse{
		Block: block,
		Tx:    tx,
	}
	return nil
}

// GetAddressUTXOs returns the UTXOs corresponding to an address.
func (n *Node) GetAddressUTXOs(address string, reply *[]tx.Output) error {
	utxoSet := &utxo.Set{Blockchain: n.blockchain}
	pubKeyHash := base58.Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	utxos, err := utxoSet.FindUTXOs(pubKeyHash)
	if err != nil {
		return err
	}

	*reply = utxos
	return nil
}

// GetAddressesUTXOs returns the UTXOs corresponding to a set of addresses.
func (n *Node) GetAddressesUTXOs(addresses []string, reply *map[string][]tx.Output) error {
	utxoSet := &utxo.Set{Blockchain: n.blockchain}
	utxosMap := make(map[string][]tx.Output)

	for _, address := range addresses {
		pubKeyHash := base58.Decode([]byte(address))
		pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

		utxos, err := utxoSet.FindUTXOs(pubKeyHash)
		if err != nil {
			return err
		}

		utxosMap[address] = utxos
	}

	*reply = utxosMap
	return nil
}

// ListBlocks returns all the blocks from the blockchain.
func (n *Node) ListBlocks(_ struct{}, reply *[]block.Block) error {
	bci := n.blockchain.NewIterator()
	count, err := bci.BlocksCount()
	if err != nil {
		return err
	}

	blocks := make([]block.Block, 0, count)
	err = bci.ForEach(func(block block.Block) error {
		blocks = append(blocks, block)
		return nil
	})
	if err != nil {
		return err
	}

	*reply = blocks
	return nil
}

// SendPing sends a ping request to all the other peers.
func (n *Node) SendPing(_ struct{}, reply *struct{}) error {
	return n.peers.ForEach(func(addr string) error {
		return n.sendPing(addr)
	})
}

// SendTx sends sends a transaction to another node and returns the transaction id.
func (n *Node) SendTx(params SendTxParams, reply *[]byte) error {
	wallet, err := wallet.Load()
	if err != nil {
		return err
	}
	defer wallet.Save()

	utxoSet := &utxo.Set{Blockchain: n.blockchain}
	tx, err := utxo.NewTx(
		wallet.Account(params.AccountName),
		params.To,
		params.Amount,
		params.Fee,
		utxoSet,
	)
	if err != nil {
		return err
	}

	if err := n.sendTx("", tx); err != nil {
		return err
	}

	*reply = tx.ID
	return nil
}

// Stop stops the running node.
func (n *Node) Stop(_ struct{}, reply *struct{}) error {
	n.interrupt <- os.Interrupt
	return nil
}

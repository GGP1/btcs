package node

import (
	"bytes"
	"fmt"
	"io"
	"net"

	"github.com/GGP1/btcs/block"
	"github.com/GGP1/btcs/encoding/gob"
	"github.com/GGP1/btcs/logger"
	"github.com/GGP1/btcs/tx"
	"github.com/GGP1/btcs/tx/utxo"
)

const (
	typeBlock = "block"
	typeTx    = "tx"
)

type handlerFunc func(request []byte) error

func (n *Node) messageHandlers() map[message]handlerFunc {
	return map[message]handlerFunc{
		msgAddr:      n.handleAddr,
		msgBlock:     n.handleBlock,
		msgInv:       n.handleInv,
		msgGetAddr:   n.handleGetAddr,
		msgGetBlocks: n.handleGetBlocks,
		msgGetData:   n.handleGetData,
		msgTx:        n.handleTx,
		msgVersion:   n.handleVersion,
		msgPing:      n.handlePing,
		msgPong:      n.handlePong,
	}
}

func (n *Node) request(address string, data []byte) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.Errorf("Peer %q is not available", address)
		n.peers.Remove(address)
		return nil
	}
	defer conn.Close()

	if _, err := io.Copy(conn, bytes.NewReader(data)); err != nil {
		return err
	}

	return nil
}

// handleAddr adds the requester address to the peer list and requests all peers for their blocks.
func (n *Node) handleAddr(request []byte) error {
	payload, err := getPayload[addr](request)
	if err != nil {
		return err
	}

	peersCount := n.peers.Add(payload.Addresses...)
	logger.Infof("There are %d known peers", peersCount)

	return n.peers.ForEach(func(addr string) error {
		return n.sendGetBlocks(addr)
	})
}

// sendAddr relays connection information for peers on the network.
func (n *Node) sendAddr(address string) error {
	nodes := addr{Addresses: append(n.peers.List(), n.hostAddress)}
	msg, err := newMessage(msgAddr, nodes)
	if err != nil {
		return err
	}

	return n.request(address, msg)
}

// handleBlock receives a block and adds it to the blockchain.
func (n *Node) handleBlock(request []byte) error {
	payload, err := getPayload[blockData](request)
	if err != nil {
		return err
	}

	block, err := gob.Decode[block.Block](payload.Block)
	if err != nil {
		return err
	}

	// Ignore invalid blocks
	if !block.IsValid() {
		return nil
	}

	if n.miner {
		// Notify the mining goroutine that we already got a new block so it restarts
		// the process
		n.newBlocks <- block
	}

	// Remove new block's transactions from the mempool
	for _, tx := range block.Transactions {
		n.txPool.Remove(tx.ID)
	}

	if err := n.blockchain.AddBlock(block); err != nil {
		return fmt.Errorf("adding block: %v", err)
	}

	utxoSet := &utxo.Set{Blockchain: n.blockchain}
	if err := utxoSet.Update(block); err != nil {
		return fmt.Errorf("updating utxo set: %v", err)
	}

	logger.Infof("Added block at height %d (%x)", block.Height, block.Hash)
	return nil
}

// sendBlock transmits a single serialized block.
//
// https://developer.bitcoin.org/reference/block_chain.html#serialized-blocks
func (n *Node) sendBlock(addr string, b block.Block) error {
	encodedBlock, err := gob.Encode(b)
	if err != nil {
		return err
	}

	blockData := blockData{
		AddrFrom: n.hostAddress,
		Block:    encodedBlock,
	}
	msg, err := newMessage(msgBlock, blockData)
	if err != nil {
		return err
	}

	return n.request(addr, msg)
}

// handleGetAddr sends the list of connected peers to the node requesting that information.
func (n *Node) handleGetAddr(request []byte) error {
	payload, err := getPayload[getblocks](request)
	if err != nil {
		return err
	}

	return n.sendAddr(payload.AddrFrom)
}

// sendGetAddr requests an "addr" message from the receiving node.
func (n *Node) sendGetAddr(address string) error {
	getAddr := getaddr{
		AddrFrom: n.hostAddress,
	}

	msg, err := newMessage(msgGetAddr, getAddr)
	if err != nil {
		return err
	}

	return n.request(address, msg)
}

// handleGetBlocks answers by sending all the block hashes the node has.
func (n *Node) handleGetBlocks(request []byte) error {
	payload, err := getPayload[getblocks](request)
	if err != nil {
		return err
	}

	hashes, err := n.blockchain.BlocksHashes()
	if err != nil {
		return err
	}

	return n.sendInv(payload.AddrFrom, typeBlock, hashes)
}

// sendGetBlocks requests block header hashes starting from
// a particular point in the blockchain.
func (n *Node) sendGetBlocks(address string) error {
	getBlocks := getblocks{
		AddrFrom: n.hostAddress,
	}
	msg, err := newMessage(msgGetBlocks, getBlocks)
	if err != nil {
		return err
	}

	return n.request(address, msg)
}

// handleGetData answers with the details of a block or transaction.
func (n *Node) handleGetData(request []byte) error {
	payload, err := getPayload[getdata](request)
	if err != nil {
		return err
	}

	switch payload.Type {
	case typeBlock:
		block, err := n.blockchain.Block([]byte(payload.ID))
		if err != nil {
			return err
		}

		if err := n.sendBlock(payload.AddrFrom, block); err != nil {
			return err
		}

	case typeTx:
		tx := n.txPool.Get(payload.ID)

		if err := n.sendTx(payload.AddrFrom, &tx); err != nil {
			return err
		}

	}

	return nil
}

// sendGetData requests one or more data objects from another node.
func (n *Node) sendGetData(address, kind string, id []byte) error {
	getData := getdata{
		AddrFrom: n.hostAddress,
		Type:     kind,
		ID:       id,
	}
	msg, err := newMessage(msgGetData, getData)
	if err != nil {
		return err
	}

	return n.request(address, msg)
}

// handleInv answers with the hashes of blocks or transactions the node has.
func (n *Node) handleInv(request []byte) error {
	payload, err := getPayload[inv](request)
	if err != nil {
		return err
	}

	logger.Infof("Received inventory with %d %s/s from %s",
		len(payload.Items),
		payload.Type,
		payload.AddrFrom)

	switch payload.Type {
	case typeBlock:
		for _, blockHash := range payload.Items {
			if err := n.sendGetData(payload.AddrFrom, typeBlock, blockHash); err != nil {
				return err
			}
		}

	case typeTx:
		for _, txID := range payload.Items {
			if !n.txPool.Contains(txID) {
				if err := n.sendGetData(payload.AddrFrom, typeTx, txID); err != nil {
					return err
				}
			}
		}

	}

	return nil
}

// sendInv transmits one or more inventories of objects known to the transmitting peer.
//
// The receiving peer can compare the inventories from an “inv” message against
// the inventories it has already seen, and then use a follow-up message
// to request unseen objects.
func (n *Node) sendInv(address, kind string, items [][]byte) error {
	inventory := inv{
		AddrFrom: n.hostAddress,
		Type:     kind,
		Items:    items,
	}
	msg, err := newMessage(msgInv, inventory)
	if err != nil {
		return err
	}

	return n.request(address, msg)
}

// handlePing answers with a "pong" message.
func (n *Node) handlePing(request []byte) error {
	payload, err := getPayload[ping](request)
	if err != nil {
		return err
	}

	return n.sendPong(payload.AddrFrom)
}

// sendPing helps confirm that the receiving peer is still connected.
//
// Bitcoin Core will, by default, disconnect from any clients which have not responded
// to a “ping” message within 20 minutes.
func (n *Node) sendPing(addr string) error {
	ping := ping{AddrFrom: n.hostAddress}
	msg, err := newMessage(msgPing, ping)
	if err != nil {
		return err
	}

	return n.request(addr, msg)
}

// handlePong logs when another peer sent a pong message.
func (n *Node) handlePong(request []byte) error {
	payload, err := getPayload[pong](request)
	if err != nil {
		return err
	}

	logger.Info(payload.AddrFrom, " says PONG")
	return nil
}

// sendPong replies to a “ping” message, proving to the pinging node
// that the ponging node is still alive.
func (n *Node) sendPong(addr string) error {
	pong := pong{AddrFrom: n.hostAddress}
	msg, err := newMessage(msgPong, pong)
	if err != nil {
		return err
	}

	return n.request(addr, msg)
}

// handleTx receives a transaction, adds it to the mempool and includes it in the next block.
func (n *Node) handleTx(request []byte) error {
	payload, err := getPayload[transaction](request)
	if err != nil {
		return err
	}

	txx, err := gob.Decode[tx.Tx](payload.Transaction)
	if err != nil {
		return err
	}

	if err := n.blockchain.VerifyTx(txx); err != nil {
		return err
	}
	n.txPool.Add(txx)

	logger.Debugf("Received a new transaction (%x) from %s", txx.ID, payload.AddrFrom)

	// Broadcast the transactions to other peers
	err = n.peers.ForEach(func(addr string) error {
		return n.sendInv(addr, typeTx, [][]byte{txx.ID})
	})
	if err != nil {
		return err
	}

	return nil
}

// sendTx transmits a single encoded transaction.
func (n *Node) sendTx(address string, tx *tx.Tx) error {
	encodedTx, err := gob.Encode(tx)
	if err != nil {
		return err
	}

	transaction := transaction{
		AddrFrom:    n.hostAddress,
		Transaction: encodedTx,
	}
	msg, err := newMessage(msgTx, transaction)
	if err != nil {
		return err
	}

	// Send the transaction to ourselves so we can include it in the next block
	if err := n.request(n.hostAddress, msg); err != nil {
		return err
	}

	if address == "" {
		return n.peers.ForEach(func(addr string) error {
			return n.request(addr, msg)
		})
	}

	return n.request(address, msg)
}

// handleVersion exchanges information with other peer to find the longer blockchain.
func (n *Node) handleVersion(request []byte) error {
	payload, err := getPayload[version](request)
	if err != nil {
		return err
	}

	bestHeight, err := n.blockchain.BestHeight()
	if err != nil {
		return err
	}
	n.peers.Add(payload.AddrFrom)

	peerBestHeight := payload.BestHeight
	if bestHeight == peerBestHeight {
		return nil
	}

	if bestHeight < peerBestHeight {
		return n.sendGetBlocks(payload.AddrFrom)
	}

	return n.sendVersion(payload.AddrFrom)
}

// sendVersion provides information about the transmitting node
// to the receiving node at the beginning of a connection.
func (n *Node) sendVersion(addr string) error {
	bestHeight, err := n.blockchain.BestHeight()
	if err != nil {
		return err
	}

	version := version{
		AddrFrom:   n.hostAddress,
		Version:    n.version,
		BestHeight: bestHeight,
	}
	msg, err := newMessage(msgVersion, version)
	if err != nil {
		return err
	}

	return n.request(addr, msg)
}

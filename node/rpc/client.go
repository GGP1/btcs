package rpc

import (
	"fmt"
	"net/rpc"

	"github.com/GGP1/btcs/block"
	"github.com/GGP1/btcs/node"
	"github.com/GGP1/btcs/tx"
)

// Client is the requester of the node's RPC server methods.
type Client struct {
	client *rpc.Client
}

// NewClient creates a client that is connected to the node's RPC server.
//
// Call Close to release the Client's associated resources when done.
func NewClient() (Client, error) {
	client, err := rpc.Dial("tcp", node.RPCAddress)
	if err != nil {
		return Client{}, fmt.Errorf("failed connecting to the RPC server, please check if the node is running: %v", err)
	}

	return Client{client}, nil
}

// AddNode adds a node to the peer list.
func (c *Client) AddNode(address string) (int, error) {
	var peersNum int
	if err := c.client.Call("Node.AddNode", address, &peersNum); err != nil {
		return 0, err
	}

	return peersNum, nil
}

// Close releases resources related to the rcp client.
func (c *Client) Close() error {
	return c.client.Close()
}

// DisconnectNode removes a node from the peer list.
func (c *Client) DisconnectNode(address string) (int, error) {
	var peersNum int
	if err := c.client.Call("Node.DisconnectNode", address, &peersNum); err != nil {
		return 0, err
	}

	return peersNum, nil
}

// GetBestHeight returns the node's blockchain best height.
func (c *Client) GetBestHeight() (int32, error) {
	var bestHeight int32
	if err := c.client.Call("Node.GetBestHeight", struct{}{}, &bestHeight); err != nil {
		return 0, err
	}

	return bestHeight, nil
}

// GetBlock returns a block given a hash.
func (c *Client) GetBlock(hash []byte) (block.Block, error) {
	var block block.Block
	if err := c.client.Call("Node.GetBlock", hash, &block); err != nil {
		return block, err
	}

	return block, nil
}

// GetLastBlock returns the last block (tip) of a chain.
func (c *Client) GetLastBlock() (block.Block, error) {
	var block block.Block
	if err := c.client.Call("Node.GetLastBlock", struct{}{}, &block); err != nil {
		return block, err
	}

	return block, nil
}

// GetPeerInfo returns data about each connected node.
func (c *Client) GetPeerInfo() ([]string, error) {
	var peerAddresses []string
	if err := c.client.Call("Node.GetPeerInfo", struct{}{}, &peerAddresses); err != nil {
		return nil, err
	}

	return peerAddresses, nil
}

// GetRawMempool returns the node's mempool transaction ids.
func (c *Client) GetRawMempool() ([]string, error) {
	var txIDs []string
	if err := c.client.Call("Node.GetRawMempool", struct{}{}, &txIDs); err != nil {
		return nil, err
	}

	return txIDs, nil
}

// GetTransaction returns a transaction with the id provided.
func (c *Client) GetTransaction(id []byte) (block.Block, tx.Tx, error) {
	var resp node.GetTransactionResponse
	if err := c.client.Call("Node.GetTransaction", id, &resp); err != nil {
		return block.Block{}, tx.Tx{}, err
	}

	return resp.Block, resp.Tx, nil
}

// GetAddressUTXOs returns the unspent outputs of an address.
func (c *Client) GetAddressUTXOs(address string) ([]tx.Output, error) {
	var utxos []tx.Output
	if err := c.client.Call("Node.GetAddressUTXOs", address, &utxos); err != nil {
		return nil, err
	}

	return utxos, nil
}

// GetAddressesUTXOs returns the UTXOs corresponding to a set of addresses.
func (c *Client) GetAddressesUTXOs(addresses []string) (map[string][]tx.Output, error) {
	var utxos map[string][]tx.Output
	if err := c.client.Call("Node.GetAddressesUTXOs", addresses, &utxos); err != nil {
		return nil, err
	}

	return utxos, nil
}

// ListBlocks returns all blocks from the chain.
func (c *Client) ListBlocks() ([]block.Block, error) {
	var blocks []block.Block
	if err := c.client.Call("Node.ListBlocks", struct{}{}, &blocks); err != nil {
		return blocks, err
	}

	return blocks, nil
}

// SendPing sends a signal request to another node.
func (c *Client) SendPing() error {
	var reply struct{}
	return c.client.Call("Node.SendPing", struct{}{}, &reply)
}

// SendTx sends sends a transaction to another node and returns the transaction id.
func (c *Client) SendTx(params node.SendTxParams) ([]byte, error) {
	var reply []byte
	if err := c.client.Call("Node.SendTx", params, &reply); err != nil {
		return nil, err
	}

	return reply, nil
}

// Stop stops the running node.
func (c *Client) Stop() error {
	var reply struct{}
	return c.client.Call("Node.Stop", struct{}{}, &reply)
}

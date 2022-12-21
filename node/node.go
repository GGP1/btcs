package node

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/GGP1/btcs/block"
	"github.com/GGP1/btcs/logger"
	"github.com/GGP1/btcs/mempool"
	"github.com/GGP1/btcs/mining"
	"github.com/GGP1/btcs/tx/utxo"
)

// Node represents a Bitcoin Node.
type Node struct {
	blockchain  *block.Chain
	txPool      *mempool.TxPool
	peers       *peers
	newBlocks   chan block.Block
	interrupt   chan os.Signal
	hostAddress string
	version     int
	miner       bool
}

// New creates a new node.
func New(hostAddr string, seedNodes []string, miner bool) (*Node, error) {
	blockchain, err := block.LoadChain()
	if err != nil {
		if err != block.ErrBlockchainNotFound {
			return nil, err
		}

		blockchain, err = block.NewChain()
		if err != nil {
			return nil, err
		}

		// Create utxo set index
		utxoSet := utxo.Set{Blockchain: blockchain}
		if err := utxoSet.Reindex(); err != nil {
			return nil, err
		}
	}

	return &Node{
		blockchain:  blockchain,
		txPool:      mempool.NewTxPool(),
		peers:       newPeers(hostAddr, seedNodes),
		interrupt:   make(chan os.Signal, 1),
		newBlocks:   make(chan block.Block, 1),
		hostAddress: hostAddr,
		miner:       miner,
		version:     1,
	}, nil
}

// Run starts the execution of the node.
func (n *Node) Run(accountName string) error {
	_, port, err := net.SplitHostPort(n.hostAddress)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", port))
	if err != nil {
		return err
	}

	rpcListener, err := n.RunRPCServer()
	if err != nil {
		return err
	}

	logger.Info("Starting node server at ", n.hostAddress)
	go n.listen(listener, n.messageHandlers())

	// Initiate the connection with the version message to caught up with the network.
	err = n.peers.ForEach(func(addr string) error {
		return n.sendVersion(addr)
	})
	if err != nil {
		return err
	}

	if n.miner {
		go func() {
			if err := n.startMining(accountName); err != nil {
				logger.Fatal("Mining: ", err)
			}
		}()
	}

	signal.Notify(n.interrupt, os.Interrupt, syscall.SIGTERM)
	<-n.interrupt

	return n.Close(listener, rpcListener)
}

// Close releases the resources associated to a node.
func (n *Node) Close(listeners ...net.Listener) error {
	for _, ln := range listeners {
		if err := ln.Close(); err != nil {
			return err
		}
	}

	close(n.newBlocks)
	close(n.interrupt)
	logger.Info("Server stopped")

	return n.blockchain.Close()
}

// listen accepts and handles TCP connections. It should be called inside a goroutine.
func (n *Node) listen(ln net.Listener, handlers map[message]handlerFunc) {
	for {
		select {
		case <-n.interrupt:
			return

		default:
			conn, err := ln.Accept()
			if err != nil {
				logger.Error("Connection: ", err)
				continue
			}

			go func() {
				if err := handleConn(conn, handlers); err != nil {
					_, port, _ := net.SplitHostPort(conn.LocalAddr().String())
					logger.Errorf("Connection on port %s: %v", port, err)
				}
				conn.Close()
			}()
		}
	}
}

func (n *Node) startMining(accountName string) error {
	miner, err := mining.NewCPUMiner(accountName, n.txPool, n.newBlocks)
	if err != nil {
		return err
	}

	for {
		prevBlock, err := n.blockchain.LastBlock()
		if err != nil {
			return err
		}

		newBlock, err := miner.Mine(&prevBlock)
		if err != nil {
			return err
		}

		// Miner is no-op or the computation was cancelled because the block has been
		// mined by another node, continue to the next one.
		if newBlock.Hash == nil {
			continue
		}

		// Broadcast the newly mined block
		err = n.peers.ForEach(func(addr string) error {
			return n.sendBlock(addr, newBlock)
		})
		if err != nil {
			return err
		}

		if err := n.blockchain.AddBlock(newBlock); err != nil {
			return err
		}

		utxoSet := &utxo.Set{Blockchain: n.blockchain}
		if err := utxoSet.Update(newBlock); err != nil {
			return err
		}
	}
}

func handleConn(conn io.ReadCloser, handlers map[message]handlerFunc) error {
	request, err := io.ReadAll(conn)
	if err != nil {
		return fmt.Errorf("reading conn: %w", err)
	}
	msg := bytesToMessage(request[:messageLength])

	handle, ok := handlers[msg]
	if !ok {
		return fmt.Errorf("unknown command: %s", msg)
	}

	if err := handle(request); err != nil {
		return fmt.Errorf("%s handler: %w", msg, err)
	}

	return nil
}

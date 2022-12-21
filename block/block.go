package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"time"

	"github.com/GGP1/btcs/encoding/gob"
	"github.com/GGP1/btcs/tx"
	"github.com/GGP1/btcs/tx/merkle"
)

const (
	// genesisCoinbaseData is the data the first Bitcoin block contains.
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
	genesisAddr         = "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
)

// Header represents a block header.
type Header struct {
	PrevBlockHash  []byte
	MerkleRootHash []byte
	Timestamp      int64
	Nonce          uint32
	Version        int32
	Bits           uint32
}

// Block represents a block in the blockchain.
type Block struct {
	*Header
	Hash         []byte
	Transactions []tx.Tx
	Height       int32
}

// NewBlock creates and returns a block without a header hash and nonce.
// It should be mined before being saved in the databse.
func NewBlock(prevBlock *Block, txs []tx.Tx) (*Block, error) {
	merkleRootHash, err := merkleRootHash(txs)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	newHeight := prevBlock.Height + 1
	blockIndex.addNode(newHeight, now)

	return &Block{
		Header: &Header{
			PrevBlockHash:  prevBlock.Hash,
			MerkleRootHash: merkleRootHash,
			Version:        1,
			Timestamp:      now,
			Bits:           CalculateNextDifficulty(*prevBlock),
		},
		Height:       newHeight,
		Transactions: txs,
	}, nil
}

// NewGenesis creates and returns the first block of the chain.
//
// It's called the "genesis", it's pre-mined and statically embedded in the client so
// every node starts with one known block.
func NewGenesis() (*Block, error) {
	coinbaseTx, err := tx.NewCoinbase(genesisAddr, genesisCoinbaseData, 0, 0)
	if err != nil {
		return nil, err
	}

	merkleRootHash, _ := hex.DecodeString("898325b2e3f11b70cc81b6f0fc97381e82294cecefc1e483e7826c09a1557714")
	hash, _ := hex.DecodeString("000000f72eda1d4d8a8418c992ef803f7e060290c1208abac7c7b1a77d27b3fc")
	return &Block{
		Header: &Header{
			Version:        1,
			PrevBlockHash:  []byte{},
			MerkleRootHash: merkleRootHash,
			Timestamp:      1670513773,
			Nonce:          374174,
			Bits:           baseDifficulty,
		},
		Hash:         hash,
		Height:       0,
		Transactions: []tx.Tx{*coinbaseTx},
	}, nil
}

// IsGenesis returns whether a block is the genesis one or not.
func (b Block) IsGenesis() bool {
	return len(b.PrevBlockHash) == 0 && b.Height == 0
}

// IsValid validates a block's proof-of-work.
func (b Block) IsValid() bool {
	target := CompactToBig(b.Bits)
	if target.Sign() <= 0 || target.Cmp(MaxTarget) > 0 {
		return false
	}

	data, err := b.PowData()
	if err != nil {
		return false
	}

	n, err := IntToBytes(int64(b.Nonce))
	if err != nil {
		return false
	}

	data = append(data, n...)
	hash := sha256.Sum256(data)
	hashInt := new(big.Int).SetBytes(hash[:])

	return hashInt.Cmp(target) <= 0
}

// PowData joins a block's fields so we can generate its hash.
//
// It does not include the nonce, which should be appended to the end of the data.
func (b Block) PowData() ([]byte, error) {
	timestamp, err := IntToBytes(b.Timestamp)
	if err != nil {
		return nil, err
	}
	bits, err := IntToBytes(int64(b.Bits))
	if err != nil {
		return nil, err
	}

	return bytes.Join(
		[][]byte{b.PrevBlockHash, b.MerkleRootHash, timestamp, bits},
		[]byte{},
	), nil
}

// IntToBytes converts an int64 into a byte array.
func IntToBytes(num int64) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, num); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// merkleRootHash returns a block's transactions merkle tree root hash.
func merkleRootHash(txs []tx.Tx) ([]byte, error) {
	encodedTxs := make([][]byte, 0, len(txs))

	for _, tx := range txs {
		encodedTx, err := gob.Encode(tx)
		if err != nil {
			return nil, err
		}
		encodedTxs = append(encodedTxs, encodedTx)
	}

	// Transactions in a block are represented using a merkle tree, and the root node hash
	// of the tree allows us to quickly check if a transaction belonged to this block.
	merkleTree := merkle.NewTree(encodedTxs)
	return merkleTree.Root.Hash, nil
}

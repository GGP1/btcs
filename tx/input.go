package tx

// Input represents a transaction input,
// each input references the output of another transaction.
type Input struct {
	Signature  []byte
	PubKey     []byte
	PrevOutput OutPoint
}

// OutPoint represents the previous output being spent.
type OutPoint struct {
	// Transaction ID of the previous output
	TxID []byte
	// Index of the referenced output in the previous transaction
	Index int
}

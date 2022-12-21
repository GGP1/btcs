package wallet

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"errors"

	"github.com/GGP1/btcs/encoding/base58"

	"github.com/btcsuite/btcd/btcec/v2"
)

// BIP32: https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki

var (
	version = []byte{0x00}

	// xpub
	public = []byte{4, 136, 178, 30}
	// xprv
	private = []byte{4, 136, 173, 228}

	zero = []byte{0}
	four = []byte{4}

	curve *btcec.KoblitzCurve = btcec.S256()

	masterKeySeed = []byte("Bitcoin seed")
)

const (
	// seed length can be between 128 and 512 bits; 256 bits is advised.
	//
	// For mnemonic seeds the range is between 128 and 256 (12-24 words).
	seedLength         int    = 256
	firstHardenedChild uint32 = 0x80000000
)

// Key defines the components of a hierarchical deterministic wallet master key.
type Key struct {
	// Version bytes (4 bytes)
	Vbytes []byte
	// Fingerprint of the parent's key (4 bytes)
	Fingerprint []byte
	// Child number (4 bytes)
	ChildNumber []byte
	// Chain code (32 bytes)
	Chaincode []byte
	// Public/Private key data (33 bytes)
	Key []byte
	// Depth (1 byte)
	Depth uint16
}

// NewMasterKey returns a new master key given a random seed.
func NewMasterKey(seed []byte) (*Key, error) {
	mac := hmac.New(sha512.New, masterKeySeed)
	_, _ = mac.Write(seed)
	intermediary := mac.Sum(nil)
	secret := intermediary[:len(intermediary)/2]

	return &Key{
		Vbytes:      private,
		Fingerprint: make([]byte, 4),
		ChildNumber: make([]byte, 4),
		Chaincode:   intermediary[len(intermediary)/2:],
		Key:         append(zero, secret...),
		Depth:       0,
	}, nil
}

// Address returns the wallet's address.
func (k *Key) Address() string {
	x, y := expand(k.Key)
	paddedKey := bytes.Join([][]byte{four, x.Bytes(), y.Bytes()}, []byte{})

	// version + pubKeyHash + checksum
	pubKeyHash := HashPubKey(paddedKey)
	versionedPayload := append(version, pubKeyHash...)
	sum := checksum(versionedPayload)
	fullPayload := append(versionedPayload, sum...)

	return string(base58.Encode(fullPayload))
}

// Child returns the ith child of a key.
func (k *Key) Child(i uint32) (*Key, error) {
	switch {
	case bytes.Compare(k.Vbytes, private) == 0:
		return k.newPrivateChild(i)

	case bytes.Compare(k.Vbytes, public) == 0:
		return k.newPublicChild(i)

	default:
		return nil, errors.New("invalid wallet version")
	}
}

// Public returns the public key version of the key.
func (k *Key) Public() *Key {
	if bytes.Compare(k.Vbytes, public) == 0 {
		return k
	}
	return &Key{public, k.Fingerprint, k.ChildNumber, k.Chaincode, privateToPublic(k.Key), k.Depth}
}

func (k *Key) newPrivateChild(i uint32) (*Key, error) {
	pub := privateToPublic(k.Key)
	intermediary, err := k.computeIntermediary(i)
	if err != nil {
		return nil, err
	}

	return &Key{
		Vbytes:      k.Vbytes,
		Fingerprint: HashPubKey(pub)[:4],
		ChildNumber: uint32ToByte(i),
		Chaincode:   intermediary[32:],
		Key:         addPrivateKeys(intermediary[:32], k.Key),
		Depth:       k.Depth + 1,
	}, nil
}

func (k *Key) newPublicChild(i uint32) (*Key, error) {
	if i >= firstHardenedChild {
		return &Key{}, errors.New("can't do private derivation on public key")
	}

	intermediary, err := k.computeIntermediary(i)
	if err != nil {
		return nil, err
	}
	pubI := privateToPublic(intermediary[:32])

	return &Key{
		Vbytes:      k.Vbytes,
		Fingerprint: HashPubKey(k.Key)[:4],
		ChildNumber: uint32ToByte(i),
		Chaincode:   intermediary[32:],
		Key:         addPublicKeys(pubI, k.Key),
		Depth:       k.Depth + 1,
	}, nil
}

func (k *Key) computeIntermediary(i uint32) ([]byte, error) {
	idx := uint32ToByte(i)
	mac := hmac.New(sha512.New, k.Chaincode)

	if bytes.Compare(k.Vbytes, private) == 0 && i < firstHardenedChild {
		pub := privateToPublic(k.Key)
		mac.Write(append(pub, idx...))
	} else {
		mac.Write(append(k.Key, idx...))
	}
	intermediary := mac.Sum(nil)

	iL := bigIntFromBytes(intermediary[:32])
	if iL.Cmp(curve.N) >= 0 || iL.Sign() == 0 {
		return nil, errors.New("invalid child")
	}

	return intermediary, nil
}

// Serialize returns the encoded version of Key.
func (k *Key) Serialize() []byte {
	buf := new(bytes.Buffer)
	buf.Write(k.Vbytes)
	buf.WriteByte(byte(k.Depth))
	buf.Write(k.Fingerprint)
	buf.Write(k.ChildNumber)
	buf.Write(k.Chaincode)
	buf.Write(k.Key)

	// Append the checksum of the key at the end of the byte slice
	serializedKey := append(buf.Bytes(), checksum(buf.Bytes())...)
	return serializedKey
}

// DeserializeKey takes a byte slice and creates a Key.
func DeserializeKey(data []byte) (*Key, error) {
	// 78 bytes from the key and 4 from the checksum
	if len(data) != 82 {
		return nil, errors.New("invalid serialized key length")
	}

	sum1 := checksum(data[0 : len(data)-4])
	sum2 := data[len(data)-4:]
	if bytes.Compare(sum1, sum2) != 0 {
		return nil, errors.New("invalid checksum")
	}

	key := &Key{
		Vbytes:      data[0:4],
		Depth:       uint16(data[4]),
		Fingerprint: data[5:9],
		ChildNumber: data[9:13],
		Chaincode:   data[13:45],
	}

	if data[45] == byte(0) {
		key.Key = data[46:78]
	} else {
		key.Key = data[45:78]
	}

	return key, nil
}

package wallet

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/big"
	"strconv"
	"strings"

	"github.com/GGP1/btcs/encoding/base58"

	"golang.org/x/crypto/ripemd160"
)

const addressChecksumLen = 4

var errInvalidAddress = errors.New("invalid address")

// FormatBalance converts the balance to the denomination specified and
// returns a string with the value and the unit.
//
// Initially, balance is represented in satoshis.
func FormatBalance(balance int, unit string) string {
	fBalance := float64(balance)

	switch strings.ToLower(unit) {
	case "sat":
		unit = "SAT"

	case "ubtc":
		fBalance = fBalance / 100
		unit = "µBTC"

	case "mbtc":
		fBalance = fBalance / 100000
		unit = "mBTC"

	default:
		fBalance = fBalance / 100000000
		unit = "BTC"
	}

	sBalance := strings.TrimRight(strings.TrimRight(strconv.FormatFloat(fBalance, 'f', 9, 64), "0"), ".")
	return sBalance + " " + unit
}

// HashPubKey hashes the public key.
//
// https://bitcoin.stackexchange.com/questions/9202/why-does-bitcoin-use-two-hash-functions-sha-256-and-ripemd-160-to-create-an-ad
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)
	hash := ripemd160.New()
	_, _ = hash.Write(publicSHA256[:])
	return hash.Sum(nil)
}

// ValidateAddress checks if address is valid.
func ValidateAddress(address string) error {
	if len(address) == 0 {
		return errInvalidAddress
	}

	pubKeyHash := base58.Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := []byte{pubKeyHash[0]}
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append(version, pubKeyHash...))

	if bytes.Compare(actualChecksum, targetChecksum) != 0 {
		return errInvalidAddress
	}

	return nil
}

// Appends to data the first (len(data) / 32)bits of the result of sha256(data)
func addChecksum(data []byte) []byte {
	// Get first byte of sha256
	hash := sha256.Sum256(data)
	firstChecksumByte := hash[0]

	// len() is in bytes so we divide by 4
	checksumBitLength := uint(len(data) / 4)

	// For each bit of check sum we want we shift the data one the left
	// and then set the (new) right most bit equal to checksum bit at that index
	// staring from the left
	dataBigInt := bigIntFromBytes(data)

	for i := uint(0); i < checksumBitLength; i++ {
		// Bitshift 1 left
		dataBigInt.Mul(dataBigInt, big.NewInt(2))

		// Set rightmost bit if leftmost checksum bit is set
		if firstChecksumByte&(1<<(7-i)) > 0 {
			dataBigInt.Or(dataBigInt, big.NewInt(1))
		}
	}

	return dataBigInt.Bytes()
}

func addPrivateKeys(k1, k2 []byte) []byte {
	i1 := bigIntFromBytes(k1)
	i2 := bigIntFromBytes(k2)
	i1.Add(i1, i2)
	i1.Mod(i1, curve.N)
	k := i1.Bytes()
	return append(zero, k...)
}

func addPublicKeys(k1, k2 []byte) []byte {
	x1, y1 := expand(k1)
	x2, y2 := expand(k2)
	return compress(curve.Add(x1, y1, x2, y2))
}

func bigIntFromBytes(b []byte) *big.Int {
	return big.NewInt(0).SetBytes(b)
}

// checksum generates a checksum for a public key.
//
// https://bitcoin.stackexchange.com/questions/110065/checksum-sha256sha256prefixdata-why-double-hashing?noredirect=1&lq=1
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

func compress(x, y *big.Int) []byte {
	two := big.NewInt(2)
	rem := two.Mod(y, two).Uint64()
	rem += 2
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(rem))
	rest := x.Bytes()
	pad := 32 - len(rest)
	if pad != 0 {
		zeroes := make([]byte, pad)
		rest = append(zeroes, rest...)
	}
	return append(b[1:], rest...)
}

// 2.3.4 of SEC1 - https://www.secg.org/sec1-v2.pdf
func expand(key []byte) (*big.Int, *big.Int) {
	x := bigIntFromBytes(key[1:])
	y := big.NewInt(0)

	ySquared := big.NewInt(0)
	ySquared.Exp(x, big.NewInt(3), nil)
	ySquared.Add(ySquared, curve.Params().B)

	y.ModSqrt(ySquared, curve.Params().P)

	yMod2 := big.NewInt(0)
	yMod2.Mod(y, big.NewInt(2))

	signY := uint64(key[0]) - 2
	if signY != yMod2.Uint64() {
		y.Sub(curve.Params().P, y)
	}

	return x, y
}

// padByteSlice returns a byte slice of the given size with contents of the
// given slice left padded and any empty spaces filled with 0's.
func padByteSlice(slice []byte, length int) []byte {
	offset := length - len(slice)
	if offset <= 0 {
		return slice
	}

	newSlice := make([]byte, length)
	copy(newSlice[offset:], slice)

	return newSlice
}

func privateToPublic(key []byte) []byte {
	return compress(curve.ScalarBaseMult(key))
}

func uint32ToByte(i uint32) []byte {
	a := make([]byte, 4)
	binary.BigEndian.PutUint32(a, i)
	return a
}

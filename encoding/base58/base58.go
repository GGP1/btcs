package base58

import (
	"bytes"
	"math/big"
)

const pubKeyHashVersion = 0x00

// All alphanumeric characters except for "0", "I", "O", and "l"
var chars = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Encode encodes a byte array to Base58.
func Encode(input []byte) []byte {
	var result []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(chars)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, chars[mod.Int64()])
	}

	// https://en.bitcoin.it/wiki/Base58Check_encoding#Version_bytes
	if input[0] == pubKeyHashVersion {
		result = append(result, chars[0])
	}

	reverseBytes(result)
	return result
}

// Decode decodes Base58-encoded data.
func Decode(input []byte) []byte {
	result := big.NewInt(0)

	for _, b := range input {
		charIndex := bytes.IndexByte(chars, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()

	if len(input) > 0 {
		if input[0] == chars[0] {
			decoded = append([]byte{pubKeyHashVersion}, decoded...)
		}
	}

	return decoded
}

// reverseBytes reverses a byte array.
func reverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

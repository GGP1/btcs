package base58

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase58(t *testing.T) {
	rawHash := "0019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"
	hash, err := hex.DecodeString(rawHash)
	assert.NoError(t, err)

	encoded := Encode(hash)
	assert.Equal(t, "14VYJtj3yEDffZem7N3PkK563wkLZZ8RjKzcfY", string(encoded))

	decoded := Decode([]byte("14VYJtj3yEDffZem7N3PkK563wkLZZ8RjKzcfY"))
	assert.Equal(t, strings.ToLower("0019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"), hex.EncodeToString(decoded))
}

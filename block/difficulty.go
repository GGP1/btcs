// Copyright (c) 2014-2016 The btcsuite developers

package block

import (
	"math/big"

	"github.com/GGP1/btcs/logger"
)

// MaxTarget is the highest proof of work value a block can have.
// It is the value 2^255 - 1. In the Bitcoin mainnet it's 2^224 - 1.
var MaxTarget = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(1))

const (
	// Difficulty of the genesis block.
	//
	// In the Bitcoin mainnet, it's 0x1d00ffff.
	baseDifficulty uint32 = 0x1e04ffff

	// targetTimePerBlock is the desired time (in seconds) to generate each block.
	//
	// In the Bitcoin mainnet, it's 10 minutes.
	targetTimePerBlock = 20

	// blocksRetargetPeriod is the number of blocks until we re-calculate the difficulty.
	//
	// In the Bitcoin mainnet, it's 2016.
	blocksRetargetPeriod = 16

	// targetTimespan is the desired amount of time that should elapse
	// before the block difficulty requirement is examined to determine how
	// it should be changed in order to maintain the desired block
	// generation rate.
	targetTimespan = targetTimePerBlock * blocksRetargetPeriod

	// retargetAdjustmentFactor is the adjustment factor used to limit the minimum and
	// maximum amount of adjustment that can occur between difficulty retargets.
	retargetAdjustmentFactor = 4
	minRetargetTimespan      = targetTimespan / retargetAdjustmentFactor
	maxRetargetTimespan      = targetTimespan * retargetAdjustmentFactor
)

// CalculateNextDifficulty adjusts the difficulty to find a block's hash
// every blocksRetargetPeriod blocks.
func CalculateNextDifficulty(prevBlock Block) uint32 {
	nextBlockHeight := prevBlock.Height + 1
	// Return the previous block's difficulty if this block
	// is not at a difficulty retarget period
	if nextBlockHeight%blocksRetargetPeriod != 0 {
		return prevBlock.Bits
	}

	// Get the timestamp of the block at the previous retarget (targetTimespan time worth of blocks)
	lastRetargetTs := blockIndex.nodeTimestamp(nextBlockHeight - blocksRetargetPeriod)
	actualTimespan := prevBlock.Timestamp - lastRetargetTs
	logger.Debugf("Difficulty adjustment. Target timespan %d seconds, actual timespan %d seconds",
		targetTimespan,
		actualTimespan)

	if actualTimespan < minRetargetTimespan {
		actualTimespan = minRetargetTimespan
	} else if actualTimespan > maxRetargetTimespan {
		actualTimespan = maxRetargetTimespan
	}

	// nextDifficulty = currentDifficulty * actualTimespan / targetTimespan
	oldTarget := CompactToBig(prevBlock.Bits)
	newTarget := new(big.Int).Mul(oldTarget, big.NewInt(actualTimespan))
	newTarget.Div(newTarget, big.NewInt(targetTimespan))

	if newTarget.Cmp(MaxTarget) > 0 {
		newTarget.Set(MaxTarget)
	}

	newTargetBits := BigToCompact(newTarget)
	logger.Debugf("Old target: %08x", prevBlock.Bits)
	logger.Debugf("New target: %08x", newTargetBits)

	return newTargetBits
}

// CompactToBig converts a compact representation of a whole number N to an
// unsigned 32-bit number. The representation is similar to IEEE754 floating
// point numbers.
//
// Like IEEE754 floating point, there are three basic components: the sign,
// the exponent, and the mantissa. They are broken out as follows:
//
//   - the most significant 8 bits represent the unsigned base 256 exponent
//
//   - bit 23 (the 24th bit) represents the sign bit
//
//   - the least significant 23 bits represent the mantissa
//
//     -------------------------------------------------
//     |   Exponent     |    Sign    |    Mantissa     |
//     -------------------------------------------------
//     | 8 bits [31-24] | 1 bit [23] | 23 bits [22-00] |
//     -------------------------------------------------
//
// The formula to calculate N is:
//
//	N = (-1^sign) * mantissa * 256^(exponent-3)
//
// This compact form is only used in bitcoin to encode unsigned 256-bit numbers
// which represent difficulty targets, thus there really is not a need for a
// sign bit, but it is implemented here to stay consistent with bitcoind.
func CompactToBig(compact uint32) *big.Int {
	// Extract the mantissa, sign bit, and exponent.
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes to represent the full 256-bit number.  So,
	// treat the exponent as the number of bytes and shift the mantissa
	// right or left accordingly.  This is equivalent to:
	// N = mantissa * 256^(exponent-3)
	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}

	// Make it negative if the sign bit is set.
	if isNegative {
		bn = bn.Neg(bn)
	}

	return bn
}

// BigToCompact converts a whole number N to a compact representation using
// an unsigned 32-bit number. The compact representation only provides 23 bits
// of precision, so values larger than (2^23 - 1) only encode the most
// significant digits of the number. See CompactToBig for details.
func BigToCompact(n *big.Int) uint32 {
	// No need to do any work if it's zero.
	if n.Sign() == 0 {
		return 0
	}

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes.  So, shift the number right or left
	// accordingly.  This is equivalent to:
	// mantissa = mantissa / 256^(exponent-3)
	var mantissa uint32
	exponent := uint(len(n.Bytes()))
	if exponent <= 3 {
		mantissa = uint32(n.Bits()[0])
		mantissa <<= 8 * (3 - exponent)
	} else {
		// Use a copy to avoid modifying the caller's original number.
		tn := new(big.Int).Set(n)
		mantissa = uint32(tn.Rsh(tn, 8*(exponent-3)).Bits()[0])
	}

	// When the mantissa already has the sign bit set, the number is too
	// large to fit into the available 23-bits, so divide the number by 256
	// and increment the exponent accordingly.
	if mantissa&0x00800000 != 0 {
		mantissa >>= 8
		exponent++
	}

	// Pack the exponent, sign bit, and mantissa into an unsigned 32-bit
	// int and return it.
	compact := uint32(exponent<<24) | mantissa
	if n.Sign() < 0 {
		compact |= 0x00800000
	}
	return compact
}

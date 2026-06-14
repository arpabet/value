/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package value

import "fmt"

/*
Decoding limits guard the parser against hostile or corrupt input. The defaults
are generous enough for normal use but finite, so a small crafted message cannot
exhaust memory or the stack. Set any limit to 0 to disable that check.
*/

var (
	// MaxParseDepth bounds the nesting depth of decoded values to prevent
	// stack exhaustion from deeply nested input.
	MaxParseDepth = 1000

	// MaxParseCollectionLen bounds the element count of a single decoded list
	// or map, so a tiny header claiming a huge length cannot force an oversized
	// allocation.
	MaxParseCollectionLen = 1 << 24 // ~16M elements

	// MaxParseByteLen bounds the length of a single decoded string, binary, or
	// extension value.
	MaxParseByteLen = 1 << 30 // 1 GiB

	// MaxDecimalExponent bounds the absolute base-10 exponent of a decoded
	// decimal. A huge exponent is cheap to encode but can make later operations
	// (converting to an integer, rescaling for Add) materialize an enormous
	// number, so reject it at decode time.
	MaxDecimalExponent = 1 << 20 // ~1e6
)

func checkDepth(depth int) error {
	if MaxParseDepth > 0 && depth > MaxParseDepth {
		return fmt.Errorf("value: max parse depth %d exceeded", MaxParseDepth)
	}
	return nil
}

func checkCollectionLen(cnt int) error {
	if cnt < 0 {
		return fmt.Errorf("value: negative collection length %d", cnt)
	}
	if MaxParseCollectionLen > 0 && cnt > MaxParseCollectionLen {
		return fmt.Errorf("value: collection length %d exceeds limit %d", cnt, MaxParseCollectionLen)
	}
	return nil
}

func checkByteLen(n int) error {
	if n < 0 {
		return fmt.Errorf("value: negative byte length %d", n)
	}
	if MaxParseByteLen > 0 && n > MaxParseByteLen {
		return fmt.Errorf("value: byte length %d exceeds limit %d", n, MaxParseByteLen)
	}
	return nil
}

// initialSliceCap returns a bounded starting capacity for a collection whose
// header claims cnt elements. It never pre-allocates more than a small cap, so
// a hostile length cannot amplify into a large allocation before any element
// has actually been read.
func initialSliceCap(cnt int) int {
	const max = 4096
	if cnt < max {
		return cnt
	}
	return max
}

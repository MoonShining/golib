package rand

import _ "unsafe"

func FastRandn(n uint32) uint32 {
	// This is similar to fastrand() % n, but faster.
	// See https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
	return uint32(uint64(FastRand()) * uint64(n) >> 32)
}

//go:linkname FastRand runtime.fastrand
func FastRand() uint32

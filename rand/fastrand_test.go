package rand

import "testing"

func TestRand(t *testing.T) {
	n := FastRandn(100)
	t.Log(n)
}

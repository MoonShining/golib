package counter

import (
	"sync"
)

type bucket struct {
	sync.Mutex
	// bucket start time(milliseconds)
	start int64
	// total count
	count int64
}

func (b *bucket) Incr(nuint int64, n int64) int64 {
	b.Lock()
	defer b.Unlock()
	if b.start == nuint {
		b.count++
		return b.count
	}

	b.start = nuint
	b.count = 1
	return b.count
}

func NewCounter(unitInMs int64, bucketSize int) *Counter {
	return &Counter{
		unit:    unitInMs,
		buckets: make([]bucket, bucketSize),
	}
}

type Counter struct {
	// time unit in each bucket(milliseconds)
	unit    int64
	buckets []bucket
}

func (c *Counter) Incr(timeInMs int64, n int64) int64 {
	nunit := timeInMs / c.unit
	idx := nunit % int64(len(c.buckets))
	return c.buckets[idx].Incr(nunit, n)
}

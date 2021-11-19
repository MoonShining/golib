package counter

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	counter := NewCounter(1000, 3)

	counter.Incr(time.Now().UnixNano()/(1000000), 1)
	time.Sleep(100 * time.Millisecond)
	counter.Incr(time.Now().UnixNano()/(1000000), 1)
	time.Sleep(100 * time.Millisecond)
	n := counter.Incr(time.Now().UnixNano()/(1000000), 1)
	assert.Equal(t, int64(3), n)

	time.Sleep(1100 * time.Millisecond)
	counter.Incr(time.Now().UnixNano()/(1000000), 1)
	n = counter.Incr(time.Now().UnixNano()/(1000000), 1)
	assert.Equal(t, int64(2), n)
}

func TestConcurrentCounter(t *testing.T) {
	counter := NewCounter(1000, 3)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		counter.Incr(time.Now().UnixNano()/(1000000), 1)
		time.Sleep(100 * time.Millisecond)
		counter.Incr(time.Now().UnixNano()/(1000000), 1)
		time.Sleep(100 * time.Millisecond)
		counter.Incr(time.Now().UnixNano()/(1000000), 1)
	}()
	go func() {
		defer wg.Done()
		counter.Incr(time.Now().UnixNano()/(1000000), 1)
		time.Sleep(100 * time.Millisecond)
		counter.Incr(time.Now().UnixNano()/(1000000), 1)
		time.Sleep(100 * time.Millisecond)
		counter.Incr(time.Now().UnixNano()/(1000000), 1)
	}()
	wg.Wait()
	n := counter.Incr(time.Now().UnixNano()/(1000000), 1)
	assert.Equal(t, int64(7), n)

	time.Sleep(1100 * time.Millisecond)
	counter.Incr(time.Now().UnixNano()/(1000000), 1)
	n = counter.Incr(time.Now().UnixNano()/(1000000), 1)
	assert.Equal(t, int64(2), n)
}

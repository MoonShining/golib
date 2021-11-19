package timingwheel

import (
	"errors"
	"net"
	"sync"
	"time"
)

var MaxTimeoutExceed = errors.New("max timeout exceed")

type TimingWheel struct {
	sync.Mutex

	interval   time.Duration
	maxTimeout time.Duration

	ticker *time.Ticker
	stop   chan struct{}

	cs  []chan struct{}
	pos int
}

func NewTimingWheel(interval time.Duration, buckets int) *TimingWheel {
	w := new(TimingWheel)
	w.interval = interval
	w.stop = make(chan struct{})
	w.maxTimeout = interval * (time.Duration(buckets))
	w.cs = make([]chan struct{}, buckets)
	for i := range w.cs {
		w.cs[i] = make(chan struct{})
	}
	w.ticker = time.NewTicker(interval)
	go w.run()
	return w
}

func (w *TimingWheel) Stop() {
	close(w.stop)
}

func (w *TimingWheel) After(timeout time.Duration) (<-chan struct{}, error) {
	if timeout >= w.maxTimeout {
		return nil, MaxTimeoutExceed
	}

	index := int(timeout / w.interval)

	w.Lock()
	index = (w.pos + index) % len(w.cs)
	b := w.cs[index]
	w.Unlock()

	return b, nil
}

func (w *TimingWheel) run() {
	for {
		select {
		case <-w.ticker.C:
			w.onTicker()
		case <-w.stop:
			w.ticker.Stop()
			return
		}
	}
}

func (w *TimingWheel) onTicker() {
	w.Lock()
	lastC := w.cs[w.pos]
	w.cs[w.pos] = make(chan struct{})
	w.pos = (w.pos + 1) % len(w.cs)
	w.Unlock()

	close(lastC)
}

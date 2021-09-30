package doublebuffer

import (
	"context"
	"sync"
	"time"
)

type buf interface {
	overflow() bool
	next() int64
}

func NewDoubleBuffer() *DoubleBuffer {
	d := &DoubleBuffer{bufs: make(chan buf, 1)}
	go d.fetchBufLoop()
	<-d.inited
	return d
}

type DoubleBuffer struct {
	sync.Mutex
	initOnce sync.Once
	inited   chan struct{}
	loadFunc func() (buf, error)
	cur      buf
	bufs     chan buf
}

func (d *DoubleBuffer) Next() (int64, error) {
	d.Lock()
	defer d.Unlock()

	if d.cur == nil {
		buf, err := d.getBuf()
		if err != nil {
			return 0, err
		}
		d.cur = buf
	}

	if d.cur.overflow() {
		buf, err := d.getBuf()
		if err != nil {
			return 0, err
		}
		d.cur = buf
	}

	return d.cur.next(), nil
}

func (d *DoubleBuffer) getBuf() (buf, error) {
	select {
	case buf := <-d.bufs:
		return buf, nil
	case <-time.After(time.Second):
		return nil, context.DeadlineExceeded
	}
}

func (d *DoubleBuffer) fetchBufLoop() {
	for {
		buf, err := d.loadFunc()
		if err != nil {
			d.bufs <- buf
			d.initOnce.Do(func() {
				close(d.inited)
			})
		}
	}
}

package linkedbuffer

import (
	"io"
	"strings"
	"sync"
)

var _ io.ReadWriter = (*LinkedBuffer)(nil)

func New(nodeCap int) *LinkedBuffer {
	return &LinkedBuffer{cache: &nodeCache{capacity: nodeCap}}
}

// LinkedBuffer is linked list of []byte
type LinkedBuffer struct {
	len int
	cur *node

	cache *nodeCache
}

func (l *LinkedBuffer) Write(p []byte) (int, error) {
	if l.cur == nil {
		l.cur = l.cache.acquire()
	}

	cur := l.cur
	lenp := len(p)
	for {
		written := cur.write(p)
		if written == len(p) {
			break
		}

		p = p[written:]
		cur.next = l.cache.acquire()
		cur = cur.next
	}
	l.len += lenp
	return lenp, nil
}

func (l *LinkedBuffer) Read(p []byte) (int, error) {
	readed, lenp := 0, len(p)

	for {
		if l.cur == nil {
			return readed, nil
		}
		n := l.cur.read(p)
		readed += n
		if readed == lenp {
			break
		}
		if !l.cur.exhausted() {
			break
		}

		tmp := l.cur
		l.cur = l.cur.next
		tmp.next = nil
		l.cache.release(tmp)

		p = p[n:]
	}
	l.len -= readed
	return readed, nil
}

func (l *LinkedBuffer) Len() int {
	return l.len
}

func (l *LinkedBuffer) String() string {
	size := 0
	for cur := l.cur; cur != nil; cur = cur.next {
		size += cur.wIdx - cur.rIdx
	}

	var sb strings.Builder
	sb.Grow(size)
	for cur := l.cur; cur != nil; cur = cur.next {
		sb.Write(cur.buf[cur.rIdx:cur.wIdx])
	}
	return sb.String()
}

type node struct {
	wIdx int
	rIdx int
	buf  []byte
	next *node
}

// write data to buf
func (n *node) write(p []byte) int {
	i := 0
	for ; i < len(p); i++ {
		if n.wIdx >= len(n.buf) {
			break
		}
		n.buf[n.wIdx] = p[i]
		n.wIdx++
	}
	return i
}

// read data from buf
func (n *node) read(p []byte) int {
	i := 0
	for ; i < len(p); i++ {
		if n.rIdx >= n.wIdx {
			break
		}
		p[i] = n.buf[n.rIdx]
		n.rIdx++
	}
	return i
}

// can't write data anymore
func (n *node) exhausted() bool {
	return n.rIdx >= len(n.buf)
}

func (n *node) reset() {
	n.wIdx, n.rIdx = 0, 0
}

type nodeCache struct {
	sync.Mutex
	capacity int
	free     []*node
}

func (nc *nodeCache) release(n *node) {
	nc.Lock()
	defer nc.Unlock()
	if len(nc.free) > 4 {
		return
	}
	n.reset()
	nc.free = append(nc.free, n)
}

func (nc *nodeCache) acquire() *node {
	nc.Lock()
	if len(nc.free) > 0 {
		ret := nc.free[len(nc.free)-1]
		nc.free[len(nc.free)-1] = nil
		nc.free = nc.free[:len(nc.free)-1]
		nc.Unlock()
		return ret
	} else {
		nc.Unlock()
		return &node{buf: make([]byte, nc.capacity)}
	}
}

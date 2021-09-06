package trylock

import "sync/atomic"

type Lock struct {
	n int32
}

func (l *Lock) TryLock() bool {
	return atomic.CompareAndSwapInt32(&l.n, 0, 1)
}

func (l *Lock) UnLock() bool {
	return atomic.CompareAndSwapInt32(&l.n, 1, 0)
}

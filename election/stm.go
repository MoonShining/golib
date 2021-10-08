package election

import (
	"context"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
	"time"
)

var defaultTimeout = time.Second

type LeaderTask interface {
	Start(ctx context.Context, term int64)
}

type StateMachine struct {
	leaderKey  string
	leaderVal  string
	leaderTask LeaderTask

	client    *clientv3.Client
	leaseUtil *lease

	leaseId  clientv3.LeaseID
	revision int64
}

func (s *StateMachine) Run(ctx context.Context) {
	for {
		ok, err := s.election(ctx)
		if err != nil {
			if err == context.Canceled {
				return
			}
			continue
		}

		if ok {
			s.leader(ctx)
		} else {
			s.follower(ctx)
		}

		if err := ctx.Err(); err == context.Canceled {
			return
		}
	}
}

func (s *StateMachine) election(ctx context.Context) (bool, error) {
	leaseID, err := s.leaseUtil.grant(ctx)
	if err != nil {
		return false, err
	}
	kvc := clientv3.NewKV(s.client)

	txnCtx, txnCancel := context.WithTimeout(ctx, defaultTimeout)
	resp, err := kvc.
		Txn(txnCtx).
		If(clientv3.Compare(clientv3.CreateRevision(s.leaderKey), "=", 0)).
		Then(clientv3.OpPut(s.leaderKey, s.leaderVal, clientv3.WithLease(leaseID))).
		Commit()
	txnCancel()
	if err != nil {
		return false, err
	}

	s.revision = resp.Header.Revision
	if resp.Succeeded {
		s.leaseId = leaseID
	}
	return resp.Succeeded, nil
}

func (s *StateMachine) leader(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		defer cancel()
		s.leaderTask.Start(ctx, s.revision)
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()
		defer cancel()
		s.leaseUtil.keepaliveLease(ctx, s.leaseId)

	}()
	wg.Wait()
}

func (s *StateMachine) follower(ctx context.Context) {
	kvc := clientv3.NewKV(s.client)
	for {
		func() {
			watcher := clientv3.NewWatcher(s.client)
			defer watcher.Close()
			rch := watcher.Watch(ctx, s.leaderKey, clientv3.WithRev(s.revision))
			for wresp := range rch {
				if wresp.Err() != nil {
					break
				}
				for _, ev := range wresp.Events {
					if ev.Type == mvccpb.DELETE {
						// no leader
						return
					}
				}
			}

			for {
				getCtx, getCancel := context.WithTimeout(ctx, defaultTimeout)
				resp, err := kvc.Get(getCtx, s.leaderKey)
				getCancel()
				if err != nil {
					if err == context.Canceled {
						return
					}
					continue
				}
				if len(resp.Kvs) == 0 {
					// no leader
					return
				}
				s.revision = resp.Header.Revision
				break
			}
		}()
	}
}

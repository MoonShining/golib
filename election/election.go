package election

import (
	"context"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// OnLeader should quit when leaseFail chan have data
type LeaderTask interface {
	Do(leaseFail chan error, term int64) chan error
}

type Leadership interface {
	// Campaign calls LeaderTask only when itself become leader
	Campaign(ctx context.Context, key string, val string, ttl int64, onLeader LeaderTask)
}

var (
	defaultTimeout = time.Second
)

func NewLeadership(client *clientv3.Client) Leadership {
	return &leadership{client: client, lease: newLease(client)}
}

type leadership struct {
	client *clientv3.Client
	lease  *lease
}

func (l *leadership) Campaign(ctx context.Context, leaderKey string, leaderData string, ttl int64, leaderTask LeaderTask) {
	kvc := clientv3.NewKV(l.client)

	for {
		leaseID, err := l.lease.grant(ctx, ttl)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		txnCtx, txnCancel := context.WithTimeout(ctx, defaultTimeout)
		resp, err := kvc.
			Txn(txnCtx).
			If(clientv3.Compare(clientv3.CreateRevision(leaderKey), "=", 0)).
			Then(clientv3.OpPut(leaderKey, leaderData, clientv3.WithLease(leaseID))).
			Commit()
		txnCancel()
		if err != nil {
			l.lease.revoke(ctx, leaseID)
			time.Sleep(time.Second)
			continue
		}

		if resp.Succeeded {
			keepaliveCtx, keepaliveCancel := context.WithCancel(ctx)
			keepAliveErr := l.lease.keepaliveLease(keepaliveCtx, leaseID)
			<-leaderTask.Do(keepAliveErr, resp.Header.Revision)
			keepaliveCancel()
			l.lease.revoke(ctx, leaseID)
		} else {
			l.lease.revoke(ctx, leaseID)
			l.watch(ctx, leaderKey, resp.Header.Revision)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (l *leadership) watch(ctx context.Context, key string, rev int64) {
	kvc := clientv3.NewKV(l.client)
	watcher := clientv3.NewWatcher(l.client)
	defer watcher.Close()

NOLEADER:
	for {
		rch := watcher.Watch(ctx, key, clientv3.WithRev(rev))
		for wresp := range rch {
			if wresp.Err() != nil {
				break
			}
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.DELETE {
					break NOLEADER
				}
			}
		}

		for {
			getCtx, getCancel := context.WithTimeout(ctx, defaultTimeout)
			resp, err := kvc.Get(getCtx, key)
			getCancel()
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			if len(resp.Kvs) == 0 {
				break NOLEADER
			}
			rev = resp.Header.Revision
			break
		}
	}
}

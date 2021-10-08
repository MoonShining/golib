package election

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var defaultLeaseTimeout = time.Second

func newLease(client *clientv3.Client, ttl int64) *lease {
	return &lease{ttl: ttl, lease: clientv3.NewLease(client)}
}

type lease struct {
	ttl   int64
	lease clientv3.Lease
}

// grant lease with ttl seconds
func (l *lease) grant(ctx context.Context) (clientv3.LeaseID, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultLeaseTimeout)
	defer cancel()
	resp, err := l.lease.Grant(ctx, l.ttl)
	if err != nil {
		return 0, err
	}
	return resp.ID, nil
}

func (l *lease) revoke(ctx context.Context, id clientv3.LeaseID) error {
	ctx, cancel := context.WithTimeout(ctx, defaultLeaseTimeout)
	defer cancel()
	_, err := l.lease.Revoke(ctx, id)
	return err
}

func (l *lease) keepaliveLease(ctx context.Context, id clientv3.LeaseID) {
	keepaliveResp, err := l.lease.KeepAlive(ctx, id)
	if err != nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			ctx, cancel := context.WithTimeout(context.TODO(), defaultTimeout)
			l.lease.Revoke(ctx, id)
			cancel()
			return
		case _, ok := <-keepaliveResp:
			if !ok {
				return
			}
		}
	}
}

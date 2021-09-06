package election

import (
	"context"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var defaultLeaseTimeout = time.Second

func newLease(client *clientv3.Client) *lease {
	return &lease{lease: clientv3.NewLease(client)}
}

type lease struct {
	lease clientv3.Lease
}

// grant lease with ttl seconds
func (l *lease) grant(ctx context.Context, ttl int64) (clientv3.LeaseID, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultLeaseTimeout)
	defer cancel()
	resp, err := l.lease.Grant(ctx, ttl)
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

func (l *lease) keepaliveLease(ctx context.Context, id clientv3.LeaseID) chan error {
	leaseFail := make(chan error, 1)
	go func() {
		keepaliveResp, err := l.lease.KeepAlive(ctx, id)
		if err != nil {
			leaseFail <- err
			return
		}
		for range keepaliveResp {
		}
		if err := ctx.Err(); err != nil {
			leaseFail <- err
		} else {
			leaseFail <- errors.New("lease chan closed")
		}

	}()
	return leaseFail
}

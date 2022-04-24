package lock

import (
	"context"
	"fmt"
	"github.com/lincx-911/lincxlock/config"
)

// LockRPC
type LockRPC struct {
}

// LockRPCArg
type LockRPCArgs struct {
	Locktype string
	Timeout  int
	JobName  string
	Hosts    []string
}

// LockReply
type LockReply struct {
	LockID int64
}

// LockRPCArg
type UnlockRPCArgs struct {
	LockID int64
}

// LockReply
type UnlockReply struct {
}

func (lr LockRPC) NewLock(ctx context.Context, arg LockRPCArgs, reply *LockReply) error {
	conf, err := config.NewLockConf(arg.Locktype, arg.Timeout, arg.Hosts)
	if err != nil {
		return err
	}
	lock, err := NewLincxLock(arg.JobName, conf)
	if err != nil {
		return err
	}
	err = lock.Glock.Lock()
	if err != nil {
		return err
	}
	reply.LockID = lock.LockID
	return nil
}

func (lr LockRPC) Unlock(ctx context.Context, arg UnlockRPCArgs, reply *UnlockReply) error {
	 
	ll, ok :=LockMap.Get(arg.LockID)
	if !ok {
		return fmt.Errorf("lock_id %d is not valid", arg.LockID)
	}
	err := ll.Glock.Unlock()
	if err != nil {
		return err
	}
	return nil
}

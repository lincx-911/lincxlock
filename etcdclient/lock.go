package etcdclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lincx-911/lincxlock/config"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// etcd client put/get demo
// use etcd/clientv3

// EtcdLock 分布式锁
type Etcdlock struct {
	Client   *clientv3.Client // 客户端
	LockName  string           // 任务名
	canle context.CancelFunc
	Lease    clientv3.Lease
	LeaseId  clientv3.LeaseID //租约id
	IsLocked bool             //是否上锁
	Timeout int 

}

var ErrLock = errors.New("acquit lock fail")

// 初始化分布式锁
func NewEtcdlock(JobName string,conf *config.LockConfig)(*Etcdlock,error){
	el := &Etcdlock{LockName: JobName,Timeout: conf.Timeout}
	err := el.Init(conf)
	if err!=nil{
		return nil,err
	}
	return el,nil
}

func (el *Etcdlock)Init(conf *config.LockConfig) error{
	var err error
	el.Client, err = clientv3.New(clientv3.Config{
		Endpoints:   conf.Hosts,
		DialTimeout: time.Duration(conf.DailTime) * time.Second,
	})
	if err!=nil{
		return err
	}
	return err
}

// Lock上锁
func (el *Etcdlock) LockOld(jobName string) error {

	return nil
}

func (el *Etcdlock) Lock()error{
	var err error
	var ctx context.Context
	ctx,el.canle = context.WithCancel(context.TODO())
	el.Lease = clientv3.NewLease(el.Client)
	lresp,err := el.Lease.Grant(context.TODO(),int64(el.Timeout))
	if err!=nil{
		return err
	}
	el.LeaseId = lresp.ID
	_,err = el.Lease.KeepAlive(ctx,el.LeaseId)
	if err!=nil{
		return err
	}
	tresp,err := el.Client.Txn(context.TODO()).If(clientv3.Compare(clientv3.CreateRevision(el.LockName),"=",0)).
	Then(clientv3.OpPut(el.LockName,"",clientv3.WithLease(el.LeaseId))).
	Else().Commit()
	if err!=nil{
		return err
	}
	if tresp.Succeeded{
		return nil
	}
	err = func () error {
		getOpts := append(clientv3.WithLastCreate(), clientv3.WithMaxCreateRev(lresp.Revision-1))
		for {
			resp, err := el.Client.Get(ctx, el.LockName, getOpts...)
			if err != nil {
				return  err
			}
			if len(resp.Kvs) == 0 {
				return  nil
			}
			lastKey := string(resp.Kvs[0].Key)
			if err = waitPre(ctx, el.Client, lastKey, resp.Header.Revision); err != nil {
				return  err
			}
		}
	}()
	if err!=nil{
		el.Unlock()
		return err
	}
	kresp,err:=el.Client.Get(ctx, fmt.Sprintf("%s/%x",el.LockName,el.LeaseId))
	if err!=nil{
		el.Unlock()
		return err
	}
	if len(kresp.Kvs)==0{
		return errors.New("lock is expired")
	}
	return nil
}
	


func waitPre(ctx context.Context, client *clientv3.Client, key string, rev int64) error {
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wr clientv3.WatchResponse
	wch := client.Watch(cctx, key, clientv3.WithRev(rev))
	for wr = range wch {
		for _, ev := range wr.Events {
			if ev.Type ==1 {
				return nil
			}
		}
	}
	if err := wr.Err(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return errors.New("lost watcher")
}


func (el *Etcdlock)Unlock()error{
	el.canle()
	el.Lease.Revoke(context.TODO(),el.LeaseId)
	return nil
}
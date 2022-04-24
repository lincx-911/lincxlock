package etcdclient

import (
	"context"
	"errors"
	"time"

	"github.com/lincx-911/lincxlock/config"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var (
	Client *clientv3.Client
)

// EtcdLock 分布式锁
type EtcdLock struct {
	Ctx      context.Context
	JobName  string // 任务名
	Session  *concurrency.Session
	Mutex    *concurrency.Mutex // Mutex
	IsLocked bool               //是否上锁
}

// InitClient 初始化客户端
func (el *EtcdLock) Init(conf *config.LockConfig) error {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Hosts,
		DialTimeout: time.Duration(conf.DailTime) * time.Second,
		Password:    conf.Password,
		Username:    conf.User,
	})

	if err != nil {
		return err
	}
	//el.Ctx, _ = context.WithTimeout(context.Background(), time.Duration(conf.Timeout*10)*time.Second)
	el.Ctx = context.Background()
	el.Session, err = concurrency.NewSession(client, concurrency.WithTTL(config.LockConf.Timeout))
	if err != nil {
		return err
	}
	el.Mutex = concurrency.NewMutex(el.Session, el.JobName)
	return nil
}

func NewEtcdLock(JobName string, conf *config.LockConfig) (*EtcdLock, error) {
	el := &EtcdLock{
		JobName: JobName,
	}
	err := el.Init(conf)
	if err != nil {
		return nil, err
	}
	return el, nil
}

// Lock etcd分布式锁加锁
func (el *EtcdLock) Lock() error {
	var err error

	err = el.Mutex.Lock(el.Ctx)
	if err != nil {
		el.Session.Close()
		return err
	}

	el.IsLocked = true
	return nil
}

// Unlock etcd分布式锁解锁
func (el *EtcdLock) Unlock() error {
	if !el.IsLocked {
		return errors.New("had not locked!")
	}
	defer el.Session.Close()
	err := el.Mutex.Unlock(el.Ctx)
	if err != nil {
		return err
	}
	return nil
}

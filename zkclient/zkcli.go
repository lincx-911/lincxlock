package zkclient

import (
	"errors"
	"time"

	"github.com/lincx-911/lincxlock/config"
	"github.com/samuel/go-zookeeper/zk"
)

// ZKLock zookeeper分布式锁
type ZkLock struct {
	Mutex    *zk.Lock
	JobName  string
	IsLocked bool //是否上锁
}

// 初始化分布式锁
func (zl *ZkLock) Init(conf *config.LockConfig) error {
	hosts := conf.Hosts
	// 连接zk
	conn, _, err := zk.Connect(hosts, time.Second*time.Duration(conf.Timeout))
	if err != nil {
		return err
	}
	zl.Mutex = zk.NewLock(conn, zl.JobName, zk.WorldACL(zk.PermAll))
	return nil
}

func NewZkLock(JobName string, conf *config.LockConfig) (*ZkLock, error) {
	zl := &ZkLock{
		JobName: JobName,
	}
	err := zl.Init(conf)
	if err != nil {
		return nil, err
	}
	return zl, nil
}

// Lock 加锁
func (zl *ZkLock) Lock() error {
	err := zl.Mutex.Lock()
	if err != nil {
		zl.IsLocked = false
		return err
	}
	zl.IsLocked = true
	return nil
}

// Unlock 解锁
func (zl *ZkLock) Unlock() error {
	if !zl.IsLocked {
		return errors.New("had not locked")
	}
	err := zl.Mutex.Unlock()
	if err != nil {
		return err
	}
	return nil
}

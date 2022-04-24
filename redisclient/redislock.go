package redisclient

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/lincx-911/lincxlock/config"

	"time"
)

const (
	SET_IF_NOT_EXIST     = "NX"            // 不存在则执行
	SET_WITH_EXPIRE_TIME = "EX"            // 过期时间(秒)  PX 毫秒
	DefaultTimout        = 5 * time.Second // 默认时间为5s
)

// Redislock redis distribute lock
type Redislock struct {
	Mutex    *redsync.Mutex //封装锁资源加锁解锁
	LockName string         // 用来作为redis 的键使用的，每个锁应该有唯一的名称
}

// NewRedisLock return struct Redislock
func NewRedisLock(lockName string, conf *config.LockConfig) (*Redislock, error) {
	rl := &Redislock{
		LockName: lockName,
	}
	locksync := NewSync(conf.Hosts, conf.Password, conf.DailTime)
	rl.Mutex = NewMutex(locksync, lockName)
	return rl, nil
}

// Lock 加锁
func (l *Redislock) Lock() error {
	return l.Mutex.Lock()
}

// Unlock 解锁
func (l *Redislock) Unlock() error {
	_, err := l.Mutex.Unlock()
	return err
}

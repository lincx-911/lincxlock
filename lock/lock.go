package lock

import (
	"fmt"
	"log"

	"github.com/lincx-911/lincxlock/config"
	ec "github.com/lincx-911/lincxlock/etcdclient"
	rc "github.com/lincx-911/lincxlock/redisclient"
	
	"github.com/lincx-911/lincxlock/util/snowflake"
	zc "github.com/lincx-911/lincxlock/zkclient"
)

const lockMapDelay = 1

// LincxLock 封装锁对象 暴露接口
type LincxLock struct {
	Glock GlobalLock
	LockID int64
	ExpiredTime int64
}

// GlobalLock 暴露加锁解锁接口
type GlobalLock interface {
	Lock() error
	Unlock() error
}
// LockList 锁对象
var LockMap *ExpiredMap

func init(){
	LockMap = NewExpiredMap(lockMapDelay)
}

// InitLock 初始化分布式锁--针对框架，读取配置文件获取锁的配置信息
func (ll *LincxLock) InitLock(fileAddr config.FileAddrType, fileType config.FileType, filePath, fileName string) error {

	err := config.LoadLocalFile(fileAddr, fileType, filePath, fileName)
	if err != nil {
		log.Printf("viper error:%v", err)
		return err
	}

	w,err:=snowflake.NewWorker(1)
	if err!=nil{
		return err
	}
	ll.LockID = w.Next()
	return nil
}

// NewLincxLock 新建锁对象
func NewLincxLock(jobName string, conf *config.LockConfig) (ll *LincxLock, err error) {
	w,err:=snowflake.NewWorker(1)
	if err!=nil{
		return nil,err
	}
	ll = new(LincxLock)
	ll.LockID = w.Next()
	var lock GlobalLock
	// 判断锁的类型
	if conf.Locktype == config.ZKLock {

		lock, err = zc.NewZkLock(jobName, conf)
		if err != nil {
			return
		}
	} else if conf.Locktype == config.EtcdLock {
		lock, err = ec.NewEtcdLock(jobName, conf)
		if err != nil {
			return
		}
	} else if conf.Locktype == config.RedisLock {

		lock, err = rc.NewRedisLock(jobName, conf)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		err = fmt.Errorf("had not support this lock type")
		return
	}
	
	
	ll.Glock = lock
	LockMap.Set(ll.LockID,ll,int64(conf.Timeout))
	return
}

// Lock lock加锁 zklock jsoname:"/xxx/xxx"--针对框架
func (ll *LincxLock) Lock() (err error) {
	err=ll.Glock.Lock()
	if err!=nil{
		LockMap.Delete(ll.LockID)
	}
	return
}

// Unlock 解锁
func (ll *LincxLock) Unlock() error {
	LockMap.Delete(ll.LockID)
	return ll.Glock.Unlock()
}


package config

import (
	"errors"
)

const (
	RedisLock     LockType = "redislock"
	ZKLock        LockType = "zklock"
	EtcdLock      LockType = "etcdlock"
	DefaultTimout          = 5// 默认时间为5s
)

// LockType 锁的类型
type LockType string

// LockConfig 分布式锁配置
type LockConfig struct {
	Locktype LockType `json:"lock_type"` //锁的类型
	Timeout  int      `json:"timeout"`   //锁的超时时间 单位秒
	Hosts    []string `json:"hosts"`     // 服务器地址
	DailTime int      `json:"dailtime"`  //连接超时时间
	Password string   `json:"password"`  //密码
	User     string   `json:"user"`      //用户名
}

//
type lockConfigOption func(lc *LockConfig)

// LockConf 全局锁配置
var LockConf *LockConfig

// TypeList 目前支持锁的类型
var TypeList []LockType

// WithDailTime returns a connection option dail server
func WithDailTime(t int)lockConfigOption{
	return func(lc *LockConfig) {
		if t<=0{
			t = DefaultTimout
		}
		lc.DailTime = t
	}
}
// WithDailTime returns a connection option dail server
func WithPassword(password string)lockConfigOption{
	return func(lc *LockConfig) {
		lc.Password = password
	}
}
// WithDailTime returns a connection option dail server
func WithAuthUser(auth string)lockConfigOption{
	return func(lc *LockConfig) {
		lc.User = auth
	}
}

// NewLockConf 新建LockConf
func NewLockConf(locktype string,locktimeout int,hosts []string,options ...lockConfigOption)(*LockConfig,error){
	if len(hosts)==0{
		return nil,errors.New("hosts is nil")
	}
	if !checkType(LockType(locktype)){
		return nil,errors.New("this lock type not supported")
	}
	if locktimeout<0{
		return nil,errors.New("timeout must >= 0")
	}else if locktimeout==0{
		locktimeout = DefaultTimout
	}
	lconf :=  &LockConfig{
		Locktype: LockType(locktype),
		Timeout: locktimeout,
		Hosts: hosts,
	}
	// Set provided options.
	for _, option := range options {
		option(lconf)
	} 
	return lconf,nil
}


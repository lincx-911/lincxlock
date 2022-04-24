package rpcserver

import (
	"log"
	"time"

	"github.com/lincx-911/lincxlock/lock"

	"github.com/lincx-911/lincxrpc/registry"
	"github.com/lincx-911/lincxrpc/registry/kvregistry"
	"github.com/lincx-911/lincxrpc/registry/memory"
	"github.com/lincx-911/lincxrpc/server"
)

var (
	DefaultRPCServerOption server.Option = server.DefaultOption
)

// StartServer 启动rpc server
func NewServer(appkey string, tags map[string]string, registry registry.Registry, serverOpt *server.Option) server.RPCServer {
	serverOpt.RegisterOption.AppKey = "lincxlock-app"
	serverOpt.Registry = registry
	serverOpt.Tags = tags
	sv := server.NewRPCServer(*serverOpt)
	err := sv.Register(lock.LockRPC{})
	if err != nil {
		log.Printf("rpc server err: %v", err.Error())
		return nil
	}
	return sv
}

// NewZkRegistry zk注册中心
// appkey 唯一标识app服务 cfg zk数据源配置 servicePath 注册中心储存路径 updateInterval 定时任务启动间隔 hots zk服务器列表
func NewZkRegistry(appkey string, servicePath string, updateInterval time.Duration, hosts []string) registry.Registry {
	return kvregistry.NewKVRegistry(kvregistry.ZK, hosts, appkey, nil, servicePath, updateInterval)
}

// NewMemoryRegistry 本地内存注册中心

func NewMemoryRegistry()registry.Registry{
	return memory.NewInMemoryRegistry()
}
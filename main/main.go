package main

import (
	//"lincxlock/httpserver"
	"context"
	"github.com/lincx-911/lincxlock/config"
	"github.com/lincx-911/lincxlock/lock"
	"github.com/lincx-911/lincxlock/rpcserver"
	"log"
	"sync"
	"time"

	"github.com/lincx-911/lincxrpc/client"
	"github.com/lincx-911/lincxrpc/codec"
	"github.com/lincx-911/lincxrpc/registry"
	"github.com/lincx-911/lincxrpc/server"
	//"github.com/lincx-911/lincxrpc/registry/libkv"
	//"github.com/lincx-911/lincxrpc/registry/memory"
)

//var registry =libkv.NewKVRegistry(libkv.ZK,[]string{"172.23.50.111:2181"},"lincxlock-app",nil, "",10*time.Second)
var (
	zkserver = "172.27.78.70:2181"
	redisserver = "172.27.78.70:6379"
	etcdserver = "172.27.78.70:2379"
	sv server.RPCServer
	re1 registry.Registry
)

func main() {
	//httpserver.StartHttp(8888)
	//httpserver.StartHttps(8881)
	
	// StartServer()
	
	// time.Sleep(10*time.Second)
	// var wg sync.WaitGroup
	// for i := 0; i < 5; i++ {
	// 	wg.Add(1)
	// 	go func(n int){
	// 		defer wg.Done()
	// 		RpcCall(n)
	// 	}(i)
		
	// }
	// wg.Wait()
	// sv.Close()
	//res1=rpcserver.NewZkRegistry("lincxlock-app","/lincxlock/service/rpc",10*time.Second,[]string{zkserver})
	CallLockLocal()
}

func CallLockLocal() {
	err:=config.LoadLocalFile(config.LocalFile, config.YamlFile, "./", "config.yaml")
	if err != nil {
		log.Printf("initlock error:%v", err)
		return
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {

			defer wg.Done()
			
			ll ,err:= lock.NewLincxLock("/1locklincx",config.LockConf)
			
			if err != nil {
				log.Printf("initlock error:%v", err)
				return
			}
			err = ll.Lock()
			if err != nil {
				log.Printf("number %v gorotine lock error:%v", n, err)
				return
			}
			log.Printf("number %v gorotine get lock\n", n)
			time.Sleep(1 * time.Second)
			log.Printf("number %v gorotine lease lock\n", n)
			ll.Unlock()
		}(i)
	}
	wg.Wait()

}


func CallLockLock(n int){
	err:=config.LoadLocalFile(config.LocalFile, config.YamlFile, "../", "config.yaml")
	if err != nil {
		log.Printf("initlock error:%v", err)
		return
	}
	ll ,err:= lock.NewLincxLock("/lincxlock/lock1",config.LockConf)
			
	if err != nil {
		log.Printf("initlock error:%v", err)
		return
	}
	err = ll.Lock()
	if err != nil {
		log.Printf("number %v lock error:%v", n, err)
		return
	}
	log.Printf("number %v get lock\n", n)
	time.Sleep(1 * time.Second)
	log.Printf("number %v gorotine lease lock\n", n)
	ll.Unlock()
}

func StartServer(){
	sv = rpcserver.NewServer("lincxlock-app",nil,re1,&rpcserver.DefaultRPCServerOption)
	go func ()  {
		defer func ()  {
			if err:=recover();err!=nil{
				log.Printf("panic err %v",err)
			}
		}()
		sv.Serve("tcp",":8881",nil)
	}()
}

func StopServer(){
	sv.Close()
}
func RpcCall(n int){
	log.Printf("provider len %d\n",len(re1.GetServiceList()))
	op := &client.DefaultSGOption
	op.AppKey = "lincxlock-app"
	op.SerializeType = codec.MessagePackType
	op.RequestTimeout = time.Second*10
	op.DialTimeout = time.Second*10
	op.FailMode = client.FailRetry
	op.Retries = 3
	// op.Selector = selector.NewAppointedSelector()
	op.Heartbeat = true
	op.HeartbeatInterval = time.Second * 10
	op.HeartbeatDegradeThreshold = 10
	op.Registry = re1
	c := client.NewSGClient(*op)
	lockargs := lock.LockRPCArgs{
		Locktype: "zklock",
		Timeout: 5,
		JobName: "/distributed-lock/lock",
		Hosts: []string{zkserver},
	}
	reply := &lock.LockReply{}
	ctx := context.Background()
	err := c.Call(ctx, "LockRPC.NewLock", lockargs,reply )
	if err != nil {
		log.Println("err!!!" + err.Error())
		return
	} 
	time.Sleep(1*time.Second)
	log.Printf("goroutine %d get lock\n",n)

	unlockargs := lock.UnlockRPCArgs{
		LockID: reply.LockID,
	}
	ulreply := &lock.UnlockReply{}
	err = c.Call(ctx, "LockRPC.Unlock", unlockargs,ulreply )
	if err!=nil{
		log.Println("err!!"+err.Error())
		return
	}
	log.Printf("goroutine %d unlock\n",n)

}


func CallLock(locktype,lockey string,hosts []string){
	conf,err:=config.NewLockConf(locktype,5,hosts)
	if err!=nil{
		
		return
	}
	ll ,err:= lock.NewLincxLock(lockey,conf)
	
	if err != nil {
		
		return
	}
	err = ll.Lock()
	if err != nil {
		
		return
	}
	
	
	ll.Unlock()
	
}
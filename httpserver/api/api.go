package api

import (
	"fmt"
	"github.com/lincx-911/lincxlock/config"
	"github.com/lincx-911/lincxlock/lock"

	"github.com/gin-gonic/gin"
)

// LockParam 加锁参数
type LockParam struct{
	config.LockConfig
	JobName string `json:"jobname"` // 锁的路径
}

// LockResponse 加锁响应
type LockResponse struct{
	LockID int64 `json:"lock_id"` // 锁的id
}

// UnLockParam 解锁参数
type UnLockParam struct{
	LockID int64 `json:"lock_id"` // 锁对象的id
}

// Lock 
func Lock(ctx *gin.Context) {
	lockconf:= LockParam{}
	_=ctx.ShouldBind(&lockconf)
	
	conf,err:=config.NewLockConf(string(lockconf.Locktype),lockconf.Timeout,lockconf.Hosts)
	if err!=nil{
		HandleParamsError(ctx,err)
		return 
	}
	lock,err:=lock.NewLincxLock(lockconf.JobName,conf)
	if err!=nil{
		HandleOperationError(ctx,err)
		return 
	}
	err=lock.Glock.Lock()
	if err!=nil{
		HandleOperationError(ctx,err)
		return 
	}
	res:=&LockResponse{
		lock.LockID,
	}
	HandleOkReturn(ctx,res)
}

// Unlock
func Unlock(ctx *gin.Context){
	unlock := &UnLockParam{}
	_ = ctx.ShouldBind(unlock)
	ll, ok :=lock.LockMap.Get(unlock.LockID)
	if !ok{
		HandleOperationError(ctx,fmt.Errorf("lock_id %d is not valid",unlock.LockID))
		return 
	}
	err := ll.Glock.Unlock()
	if err!=nil{
		HandleServerError(ctx,err)
		return 
	}
	HandleOkReturn(ctx,err)
}
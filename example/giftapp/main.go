package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
	"github.com/lincx-911/lincxlock/config"
	"github.com/lincx-911/lincxlock/lock"
	"github.com/gin-gonic/gin"
)

const(
	OK             = iota
	ParamsError    // 传参错误
	ServerError    // 服务错误
)

var (
	timeout               = 3
	Gift_Lock_Prefix      = "gift_lock_"
	User_Gift_Lock_Prefix = "user_gift_lcok"
	USER_GIFTED_KEY       = "user_gift"
	LockType = "redislock"
	LockHosts = []string{"172.27.47.185:6371"}
)


type RewardParam struct{
	UserID string `json:"user_id"`
	GiftID string `json:"gift_id"`
}

// GetRewardHandle 领奖controller层
func GetRewardHandle(ctx *gin.Context) {
	user_id := ctx.Query("user_id")
	if len(user_id)==0{
		HandleAPIReturn(ctx,ParamsError,"param user_id is nil",nil)
		return 
	}
	gift_id := ctx.Query("gift_id")
	if len(gift_id)==0{
		HandleAPIReturn(ctx,ParamsError,"param gift_id is nil",nil)
		return 
	}
	res,err := RewardService(user_id,gift_id)
	if err!=nil{
		HandleAPIReturn(ctx,ServerError,err.Error(),nil)
		return
	}
	HandleAPIReturn(ctx,OK,"success",res)
}

//HandleAPIReturn 返回请求
func HandleAPIReturn(ctx *gin.Context, code int, msg string, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})
}


// RewardService 领奖service层
func RewardService(userid, giftid string) (string, error) {
	conf,err:=config.NewLockConf(LockType,5,LockHosts,config.WithPassword("1234"))
	if err!=nil{
		return "",err
	}
	uMutex ,err:= lock.NewLincxLock(User_Gift_Lock_Prefix+userid,conf)
	
	if err != nil {
		
		return "",err
	}
	// 领奖逻辑
//	log.Printf("%s lockname:%s,value:%s",userid,uMutex.Name(),uMutex.Value())
	err = uMutex.Lock()
	if err!=nil{
		return "",err
	}
	defer uMutex.Unlock()
	endTime := time.Now().Add(time.Second * time.Duration(timeout))
	for time.Now().Before(endTime) {
		if SISMembers(Pool.Get(), USER_GIFTED_KEY, userid) {
			return "", errors.New("已领奖")
		}
		// 加锁，领奖
		lock_key := Gift_Lock_Prefix + giftid
		gMutex,err := lock.NewLincxLock(lock_key,conf)
		if err!=nil{
			continue
		}
		err = gMutex.Lock()
		if err!=nil{
			continue
		}
		defer gMutex.Unlock()
		//礼包数量
		numstr, err := Get(Pool.Get(), giftid)
		if err != nil {
			return "", err
		}
		num, _ := strconv.Atoi(numstr)
		if num <= 0 {
			return "", errors.New("礼包已领完")
		}
		// 礼包数量减一
		err = Decr(Pool.Get(), giftid)
		if err != nil {
			return "", err
		}
		// 设置用户领奖状态
		err = SAdd(Pool.Get(), USER_GIFTED_KEY, userid)
		if err != nil {
			return "", err
		}
		return giftid, nil
	}
	return "", errors.New("请求失败，请重试")
}

// InitRouter 初始化路由
func InitRouter() *gin.Engine{
	router := gin.Default()
	rLock := router.Group("/lincx")
	{
		rLock.GET("/reward",GetRewardHandle)
		
	}
	return router
}

func main() {
	username := "user"
	gift_id := "wzry11220-3332211-5565222"
	ok := true
	if !ok {

	} else {
		var wg sync.WaitGroup
		for i := 0; i < 120; i++ {
			wg.Add(1)
			go func(user string) {
				defer wg.Done()
				_, err := RewardService(user, gift_id)
				if err != nil {
					log.Printf("%s reword fail:%v", user, err)
					return
				}
				log.Printf("%s reword successfully", user)

			}(username+strconv.Itoa(i))

		}
		wg.Wait()
	}
	
	
}

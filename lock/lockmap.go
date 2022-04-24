package lock


import (
	"sync"
	"time"
)

type val struct {
	data        *LincxLock
	expiredTime int64
}

const delChannelCap = 100

type ExpiredMap struct {
	m       map[int64]*val
	timeMap map[int64][]int64
	lck     *sync.Mutex
	stop    chan struct{}
	delay time.Duration
}

type delMsg struct {
	keys []int64
	t    int64
}

func NewExpiredMap(delay int) *ExpiredMap {
	e := ExpiredMap{
		m:       make(map[int64]*val),
		lck:     new(sync.Mutex),
		timeMap: make(map[int64][]int64),
		stop:    make(chan struct{}),
		delay: time.Duration(delay),
	}
	go e.run()
	return &e
}

func (em *ExpiredMap)run(){
	now := time.Now().Unix()
	ticker := time.NewTicker(em.delay*time.Second)
	defer ticker.Stop()
	delCh := make(chan *delMsg, delChannelCap)
	go func ()  {
		for v:= range delCh{
			em.multiDelete(v.keys,v.t)
		}
	}()
	for{
		select{
		case <-ticker.C:
			now++
			em.lck.Lock()
			if keys, found := em.timeMap[now]; found {
				delCh <- &delMsg{keys: keys, t: now}
			}
			em.lck.Unlock()
		case <-em.stop:
			close(delCh)
			return
		}
	}
}

func (em *ExpiredMap) Set(key int64, value *LincxLock, expireSeconds int64) {
	if expireSeconds <= 0 {
		return
	}
	em.lck.Lock()
	defer em.lck.Unlock()
	expiredTime := time.Now().Unix() + expireSeconds
	em.m[key] = &val{
		data:        value,
		expiredTime: expiredTime,
	}
	em.timeMap[expiredTime] = append(em.timeMap[expiredTime], key) //过期时间作为key，放在map中
}


func (e *ExpiredMap) Get(key int64) (value *LincxLock,found bool) {
	e.lck.Lock()
	defer e.lck.Unlock()
	if found = e.checkDeleteKey(key); !found {
		return
	}
	value = e.m[key].data
	return
}

func (e *ExpiredMap) Delete(key int64) {
	e.lck.Lock()
	delete(e.m, key)
	e.lck.Unlock()
}


func (e *ExpiredMap) multiDelete(keys []int64, t int64) {
	e.lck.Lock()
	defer e.lck.Unlock()
	delete(e.timeMap, t)
	for _, key := range keys {
		delete(e.m, key)
	}
}


func (e *ExpiredMap) Remove(key int64) {
	e.Delete(key)
}

func (e *ExpiredMap) TTL(key int64) int64 {
	e.lck.Lock()
	defer e.lck.Unlock()
	if !e.checkDeleteKey(key) {
		return -1
	}
	return e.m[key].expiredTime - time.Now().Unix()
}

func (e *ExpiredMap) Clear() {
	e.lck.Lock()
	defer e.lck.Unlock()
	e.m = make(map[int64]*val)
	e.timeMap = make(map[int64][]int64)
}
 
func (e *ExpiredMap) Close() { // todo 关闭后在使用怎么处理
	e.lck.Lock()
	defer e.lck.Unlock()
	e.stop <- struct{}{}
	//e.m = nil
	//e.timeMap = nil
}
 
func (e *ExpiredMap) Stop() {
	e.Close()
}

func (e *ExpiredMap) checkDeleteKey(key int64) bool {
	if val, found := e.m[key]; found {
		if val.expiredTime <= time.Now().Unix() {
			delete(e.m, key)
			//delete(e.timeMap, val.expiredTime)
			return false
		}
		return true
	}
	return false
}
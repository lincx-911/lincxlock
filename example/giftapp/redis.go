package main

import (
	"log"
	"time"

	
	redigolib "github.com/gomodule/redigo/redis"
)

// var slock = `
// if redigolib.call("EXISTS", KEYS[1]) == 1 then
// 	return 0
// 	end
// 	return redigolib.call("SET", KEYS[1], ARGV[1], "EX", ARGV[2], "NX")
// `
var ScriptUnlock = `
	if redigolib.call("GET", KEYS[1]) == ARGV[1] then
	return redigolib.call("DEL", KEYS[1])
	else
	return 0
	end
`
const(
	SUCCESS_REPLY = "OK"
)

var (
	Conn redigolib.Conn
	Pool *redigolib.Pool
)

func init() {
	//var err error
	// Conn, err = redigolib.Dial("tcp", "172.27.78.70:6371", redigolib.DialDatabase(0),)
	// 	//red.DialPassword(""),
	// 	// redigolib.DialPassword("1234"))
	// if err != nil {
	// 	log.Println("redigolib.Dial err=", err)
	// }
	Pool = GetPool("172.27.47.185:6371","1234")
}



func GetPool(server ,password string)*redigolib.Pool{
	if Pool!=nil{
		return Pool
	}
	return &redigolib.Pool{
		MaxIdle:     5,//空闲数
        IdleTimeout: 240 * time.Second,
        MaxActive:   10,//最大数
		Wait:        true,
        Dial: func() (redigolib.Conn, error) {
            c, err :=redigolib.Dial("tcp", server)
            if err != nil {
                return nil, err
            }
            if password != "" {
                if _, err := c.Do("AUTH", password); err != nil {
                    c.Close()
                    return nil, err
                }
            }
            return c, err
        },
		TestOnBorrow: func(c redigolib.Conn, t time.Time) error {
            _, err := c.Do("PING")
            return err
        },
	}
}

// Get redigolib get
func Get(conn redigolib.Conn, key string) (string, error) {
	defer conn.Close()
	val, err := redigolib.String(conn.Do("GET", key))
	if err != nil {
		log.Printf("redigolib get error: %s\n", err.Error())
		return "", err
	}
	return val, err
}

// Set redigolib set
func Set(conn redigolib.Conn, key string,value interface{}) (bool, error) {
	defer conn.Close()
	ok, err := conn.Do("SET", key, value)
	if err != nil {
		log.Printf("redigolib set error: val:%s %s\n", ok,err.Error())
		return false, err
	}
	return ok=="OK", nil
}
// SetNX redigolib setnx
func SetNX(conn redigolib.Conn, key, value string,timeout int)(bool,error){
	defer conn.Close()
	reply, err := redigolib.String(conn.Do("SET", key, value, "EX", timeout, "NX"))
	if reply==SUCCESS_REPLY{
		return true,nil
	}
	if err==redigolib.ErrNil{
		return false,nil
	}
	return false,err
}

// HSet redigolib hset
func HSet(conn redigolib.Conn, key, field string, data interface{}) error {
	defer conn.Close()
	_, err := conn.Do("HSET", key, field, data)
	if err != nil {
		log.Printf("redigolib hSet error: %s\n", err.Error())
	}
	return err
}

// HGet redigolib hget
func HGet(conn redigolib.Conn, key, field string) (interface{}, error) {
	defer conn.Close()
	data, err := conn.Do("HGET", key, field)
	if err != nil {
		log.Printf("redigolib hSet error: %s\n", err.Error())
		return nil, err
	}
	return data, nil
}

/**
redigolib INCR 将 key 所储存的值加上增量 1
*/
func Incr(conn redigolib.Conn, key string) error {
	defer conn.Close()
	_, err := conn.Do("INCR", key)
	if err != nil {
		log.Printf("redigolib incrby error: %s\n", err.Error())
		return err
	}
	return nil
}

/**
redigolib INCRBY 将 key 所储存的值加上增量 n
*/
func IncrBy(conn redigolib.Conn, key string, n int) error {
	defer conn.Close()
	_, err := conn.Do("INCRBY", key, n)
	if err != nil {
		log.Printf("redigolib incrby error: %s\n", err.Error())
		return err
	}
	return nil
}

/**
redigolib DECR 将 key 中储存的数字值减一。
*/
func Decr(conn redigolib.Conn, key string) error {
	defer conn.Close()
	_, err := conn.Do("DECR", key)
	if err != nil {
		log.Printf("redigolib decr error: %s\n", err.Error())
		return err
	}
	return nil
}

/**
redigolib SADD 将一个或多个 member 元素加入到集合 key 当中，已经存在于集合的 member 元素将被忽略。
*/
func SAdd(conn redigolib.Conn,key, v string) error {
	defer conn.Close()
	_, err := conn.Do("SADD", key, v)
	if err != nil {
	   log.Printf("SADD error: %s", err.Error())
	   return err
	}
	return nil
 }
 
 /**
 redigolib SMEMBERS 返回集合 key 中的所有成员。
 return map
 */
 func SMembers(conn redigolib.Conn,key string) (interface{}, error) {
	defer conn.Close()
	data, err := redigolib.Strings(conn.Do("SMEMBERS", key))
	if err != nil {
	   log.Printf("json nil: %v", err)
	   return nil, err
	}
	return data, nil
 }
 
 /**
 redigolib SISMEMBER 判断 member 元素是否集合 key 的成员。
 return bool
 */
 func SISMembers(conn redigolib.Conn,key, v string) bool {
	defer conn.Close()
	b, err := redigolib.Bool(conn.Do("SISMEMBER", key, v))
	if err != nil {
	   log.Printf("SISMEMBER error: %s", err.Error())
	   return false
	}
	return b
 }

  /**
 redigolib Script 执行脚本
 return interface{},error
 */
 func Script(conn redigolib.Conn,scriptStr string,args ...interface{})(interface{},error){
	defer conn.Close()
	lua := redigolib.NewScript(1, scriptStr)
	lua.Load(conn)
	redigolibargs := redigolib.Args.Add(args)
	reply, err := lua.Do(conn, redigolibargs...)
	if err!=nil{
		log.Printf("SISMEMBER error: %s", err.Error())
		if err==redigolib.ErrNil{
			return nil,nil
		}
		return nil,err
	}

	return reply,err
 }

 func RedisInt(reply interface{}, err error)(int,error){
	 
	num,err:=redigolib.Int(reply,err)
	if err==redigolib.ErrNil{
		return 0,nil
	}
	return num,err
 }
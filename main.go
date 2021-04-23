package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gitstliu/go-redis-cluster"
	"strings"
	"time"
)

var (
	GVA_REDIS  *redis.Cluster
)

//redis操作接口参数结构体
type goRedisParam struct {
	Data  []string `json:"data"`
}

//redis操作接口结果结构体
type goRedisReturn struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

//初始化redis集群
func initRedis() {
	redisConnArr :=  []string{"127.0.0.1:6379","127.0.0.1:6380","127.0.0.1:6381"}
	cluster, err := redis.NewCluster(
		&redis.Options{
			StartNodes:   redisConnArr,
			ConnTimeout:  500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
			WriteTimeout: 500 * time.Millisecond,
			KeepAlive:    16,
			AliveTime:    60 * time.Second,
		})
	GVA_REDIS = cluster
	if err != nil {
		fmt.Print("redis集群初始化失败", err.Error())
	}
}

//操作redis集群
func GoRedis(c *gin.Context) {
	c.Header("content-type", "application/json;charset=utf-8")
	//接收参数并校验
	var res goRedisParam
	c.ShouldBindJSON(&res)
	data_len := len(res.Data)
	if data_len == 0 {
		c.Writer.WriteString("{\"status\":\"-1\",\"message\":\"cmd empty\"}")
		return
	}

	//初始化redis cluster
	initRedis()

	var rep interface{}
	var redis_err interface{}
	var return_data goRedisReturn
	return_data.Status = "0"
	return_data.Message = "OK"
	cmd := strings.ToLower(res.Data[0])

	//批量处理
	if cmd == "batch" {
		if data_len < 2 {
			c.Writer.WriteString("{\"status\":\"-1\",\"message\":\"fail,params error\"}")
			return
		}

		batch := GVA_REDIS.NewBatch()
		for i := 0; i < data_len; i++ {
			if i > 0 {
				//解析json
				var tempArr []string
				var data []byte = []byte(res.Data[i])
				json.Unmarshal(data, &tempArr)

				arr_len := len(tempArr)
				if arr_len == 0 {
					fmt.Print("RunBatch error: %s", res.Data[i])
					c.Writer.WriteString("{\"status\":\"-1\",\"message\":\"fail,no json\"}")
					return
				}
				//传参
				var params []interface{}
				for k := 0; k < arr_len; k++ {
					if k > 0 {
						params = append(params, tempArr[k])
					}
				}
				batch.Put(tempArr[0], params...)
			}
		}
		_, err := GVA_REDIS.RunBatch(batch)
		if err != nil {
			fmt.Print("RunBatch error: %s", err.Error())
			c.Writer.WriteString("{\"status\":\"-1\",\"message\":\"fail,RunBatch error\"}")
			return
		}
		c.Writer.WriteString("{\"status\":\"0\",\"message\":\"OK\",\"data\":\"\"}")
		return
	}

	value_type := typeof(cmd)

	var params []interface{}
	for k := 0; k < data_len; k++ {
		if k > 0 {
			params = append(params, res.Data[k])
		}
	}
	if value_type == "int" {
		rep, redis_err = redis.Int(GVA_REDIS.Do(res.Data[0], params...))
	} else if value_type == "map" {
		rep, redis_err = redis.StringMap(GVA_REDIS.Do(res.Data[0], params...))
	} else if value_type == "array" {
		rep, redis_err = redis.Strings(GVA_REDIS.Do(res.Data[0], params...))
	} else {
		rep, redis_err = redis.String(GVA_REDIS.Do(res.Data[0], params...))
	}

	if redis_err != nil {
		fmt.Print("指令执行失败", redis_err)
		c.Writer.WriteString("{\"status\":\"-1\",\"message\":\"fail,no result\"}")
		return
	}
	return_data.Data = rep
	json_str, err := json.Marshal(return_data)
	if err != nil {
		fmt.Print("操作结果转json异常", err)
		c.Writer.WriteString("{\"status\":\"-1\",\"message\":\"fail,json error\"}")
		return
	}
	c.Writer.WriteString(string(json_str))
	return
}

type server interface {
	ListenAndServe() error
}

func main() {
	var Router = gin.Default()
	Router.POST("goRedis", GoRedis)

	s := &http.Server{
		Addr:           ":8282",
		Handler:        Router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Printf("program is running\n")
	fmt.Print(s.ListenAndServe())
}

//传入变量，返回类型typeof
func typeof(cmd string) string {
	value_type := "string"
	if cmd == "del" || cmd == "llen" || cmd == "hlen" || cmd == "hset" || cmd == "hexists" || cmd == "lpush" || cmd == "rpush" || cmd == "sadd" || cmd == "scard" || cmd == "srem" || cmd == "zadd" || cmd == "zcard" || cmd == "zrem" || cmd == "ttl" || cmd == "expire" || cmd == "lrem" || cmd == "hdel" {
		value_type = "int"
	} else if cmd == "hgetall" {
		value_type = "map"
	} else if cmd == "hkeys" || cmd == "hmget" || cmd == "lrange" || cmd == "smembers" {
		value_type = "array"
	}
	return value_type
}
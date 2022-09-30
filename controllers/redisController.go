package controllers

import (
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/gomodule/redigo/redis"
)

type RedisController struct {
	beego.Controller
}

//test for redis function
func (c *RedisController) Get() {

	redisaddr, err := beego.AppConfig.String("redisaddr")

	conn, err := redis.Dial("tcp", redisaddr)
	if err != nil {
		logs.Info("failed to connect to redis server,err:%s", err.Error())
	}

	conn.Send("mset", "k1", "1", "k2", "v2")
	conn.Send("mget", "k1", "k2")
	conn.Flush()
	reply, err := conn.Receive()
	if err != nil {
		logs.Info("failed to write to redis,err:%s", err.Error())
	}
	logs.Info(reply)

	//通过事务的方式进行get,set
	conn.Send("MULTI")
	conn.Send("mset", "k3", "v3", "k4", "v4")
	conn.Send("mget", "k1", "k2", "k3", "k4")
	reply, err = conn.Do("EXEC")
	if err != nil {
		logs.Info("failed to write to redis,err:%s", err.Error())
	}
	logs.Info(reply)

	//通过适配读取响应的值
	reply, _ = redis.String(conn.Do("get", "k1"))
	logs.Info("String get k1:", reply)

	reply, _ = redis.Int(conn.Do("get", "k1"))
	logs.Info("int get k1:", reply)

	reply, _ = redis.Strings(conn.Do("mget", "k1", "k2", "k3"))
	logs.Info("mget k1 k2 k3:", reply)

	//scan过程中连续读取，如果出现错误，就不往下读
	var k1 int
	var k2 int
	var k3 string
	reply1, _ := redis.Values(conn.Do("mget", "k1", "k2", "k3"))
	logs.Info("before scan values:", reply1)
	redis.Scan(reply1, &k1, &k2, &k3)
	logs.Info("scan values:", k1, k2, k3)

	c.Ctx.WriteString("redis响应成功")
}

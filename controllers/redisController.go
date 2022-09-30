package controllers

import (
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/gomodule/redigo/redis"
)

type RedisController struct {
	beego.Controller
}

func (c *RedisController) Get() {

	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		logs.Info("failed to connect to redis server,err:%s", err.Error())
	}

	conn.Send("mset", "k1", "v1", "k2", "v2")
	conn.Send("mget", "k1", "k2")
	conn.Flush()
	reply, err := conn.Receive()
	if err != nil {
		logs.Info("failed to write to redis,err:%s", err.Error())
	}
	logs.Info(reply)

	conn.Send("MULTI")
	conn.Send("mset", "k3", "v3", "k4", "v4")
	conn.Send("mget", "k1", "k2", "k3", "k4")
	reply, err = conn.Do("EXEC")
	if err != nil {
		logs.Info("failed to write to redis,err:%s", err.Error())
	}
	logs.Info(reply)

	c.Ctx.WriteString("redis响应成功")
}

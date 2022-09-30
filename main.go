package main

import (
	_ "web/models"
	_ "web/routers"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	beego.BConfig.Listen.EnableAdmin = true
	beego.AddFuncMap("getPrevIndex", getPrevIndex)
	beego.AddFuncMap("getNextIndex", getNextIndex)

	beego.Run()
}

func getPrevIndex(pageNum int) int {
	return pageNum - 1
}

func getNextIndex(pageNum int) int {
	return pageNum + 1
}

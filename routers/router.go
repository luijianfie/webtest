package routers

import (
	"web/controllers"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

func init() {

	beego.InsertFilter("/index", beego.BeforeExec, filterFunc)
	beego.Router("/", &controllers.MainController{})
	beego.Router("/register", &controllers.RegisterController{})
	beego.Router("/login", &controllers.LoginController{})
	beego.Router("/addArticle", &controllers.MainController{}, "get:ShowAdd;post:HandleAddArticle")
	beego.Router("/index", &controllers.MainController{}, "get:ShowIndex")
	beego.Router("/content", &controllers.MainController{}, "get:ShowContent")
	beego.Router("/update", &controllers.MainController{}, "get:ShowUpdate;post:HandleUpdateArticle")
	beego.Router("/delete", &controllers.MainController{}, "get:HandleDeleteArticle")

	//增加文章类型页
	beego.Router("/addType", &controllers.MainController{}, "get:ShowAddType;post:HandleAddType")
	beego.Router("/deleteType", &controllers.MainController{}, "get:HandleDeleteType")

	beego.Router("/logout", &controllers.LoginController{}, "get:HandleLogout")

	//redis router
	beego.Router("/redis", &controllers.RedisController{})
}

func filterFunc(ctx *context.Context) {
	userid := ctx.Input.Session("userid")
	if userid == nil {
		ctx.Redirect(302, "/login")
	}
}

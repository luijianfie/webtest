package controllers

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"path"
	"time"
	"web/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/common/log"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	// c.Data["data"] = "这是一个get方法"
	// c.Data["Email"] = "astaxie@gmail.com"

	c.Redirect("/index", 302)
	// c.TplName = "index.html"
	// c.Redirect("/login", 302)
	// u := models.User{Name: "terry1"}
	// nu := models.User{Name: "tommy1", Pwd: "2222"}

	// err := updateUser(&u, &nu)

	// if err != nil {
	// 	logs.Info("err:%v", err)
	// 	// fmt.Printf("err:%v\n", err)
	// }
	// logs.Info("user:%v", u)
}

func (c *MainController) Post() {

}

func (c *MainController) ShowAdd() {

	uid := c.GetSession("userid")
	if uid == nil {
		c.Redirect("/login", 302)
		return
	}

	c.TplName = "add.html"
	c.Layout = "layout.html"
	var types []*models.ArticleType

	//从redis读取缓存

	redisaddr, _ := beego.AppConfig.String("redisaddr")
	redisConn, connerr := redis.Dial("tcp", redisaddr)

	if connerr != nil {
		logs.Info("failed to connect to redis,addr:%s", redisaddr)
	} else {
		defer redisConn.Close()
		var typebuf []byte
		typebuf, err := redis.Bytes(redisConn.Do("get", "types"))
		if err == nil {
			dec := gob.NewDecoder(bytes.NewBuffer(typebuf))
			dec.Decode(&types)
		}
	}

	if types == nil {
		o := orm.NewOrm()
		_, err := o.QueryTable("ArticleType").All(&types)
		if err != nil {
			logs.Info("failed to get article type,err:%s", err.Error())
			logs.Info(err.Error())
			return
		}
		if connerr == nil {
			var typebuf bytes.Buffer
			enc := gob.NewEncoder(&typebuf)
			enc.Encode(&types)
			redisConn.Do("set", "types", typebuf)
		}

	}

	c.Data["types"] = types
}

func (c *MainController) ShowIndex() {

	uid := c.GetSession("userid")
	if uid == nil {
		c.Redirect("/login", 302)
		return
	}

	c.TplName = "index.html"

	typeId, err := c.GetInt("select")
	if err != nil {
		typeId = -1
	}

	isFirstPage := false
	isLastPage := false
	//获取当前的pageid
	pageIndex, err := c.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}

	pageSize := 5

	var articles []models.Article
	var types []models.ArticleType

	o := orm.NewOrm()
	_, err = o.QueryTable("ArticleType").All(&types)
	if err != nil {
		logs.Info(err.Error())
		return
	}

	table := o.QueryTable("Article")
	if err != nil {
		logs.Info(err.Error())
		return
	}
	var totalRecs int64

	if typeId == -1 {
		totalRecs, err = table.Count()
	} else {
		totalRecs, err = table.RelatedSel("Type").Filter("Type__Id", typeId).Count()
	}

	if err != nil {
		logs.Info(err.Error())
		return
	}

	totalPages := (totalRecs + int64(pageSize) - 1) / int64(pageSize)

	if typeId == -1 {
		table.Limit(pageSize, (pageIndex-1)*pageSize).RelatedSel("Type").All(&articles)
	} else {
		table.RelatedSel("Type").Filter("Type__Id", typeId).Limit(pageSize, (pageIndex-1)*pageSize).All(&articles)
	}
	if pageIndex == 1 {
		isFirstPage = true
	}
	if pageIndex == int(totalPages) {
		isLastPage = true
	}

	c.Data["articles"] = articles
	c.Data["curPage"] = pageIndex
	c.Data["totalRecs"] = totalRecs
	c.Data["totalPages"] = totalPages
	c.Data["isFirstPage"] = isFirstPage
	c.Data["isLastPage"] = isLastPage
	c.Data["types"] = types
	c.Data["typeId"] = typeId
	logs.Info("typeId:%d", typeId)

	c.Layout = "layout.html"
}

func (c *MainController) HandleAddArticle() {

	uid := c.GetSession("userid")
	if uid == nil {
		c.Redirect("/login", 302)
		return
	}

	name := c.GetString("articleName")
	content := c.GetString("content")
	typeId, err := c.GetInt("select")

	c.TplName = "add.html"
	if len(name) == 0 || len(content) == 0 {
		c.Data["errmsg"] = "invalid title or content."
		return
	}

	if err != nil {
		logs.Info("failed to get article type.")
		c.Data["errmsg"] = "failed to get article type."

		return
	}

	f, h, err := c.GetFile("uploadname")

	//上传文件失败打印相关信息
	if err != nil && f != nil {
		logs.Info(err.Error())
		c.Data["errmsg"] = "failed to upload file."
		f.Close()
		return
	}

	a := models.Article{
		Name:    name,
		Content: content,
	}

	//保存文件的，并且填充结构体信息
	if f != nil {
		defer f.Close()

		ext := path.Ext(h.Filename)
		if ext != ".jpg" && ext != ".png" {
			logs.Info("wrong file type:%s", ext)
			c.Data["errmsg"] = "wrong file type."
			return
		}
		if h.Size > 100000 {
			logs.Info("size too big:%d", h.Size)
			c.Data["errmsg"] = "too big."
			return
		}

		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

		c.SaveToFile("uploadname", "./static/img/"+filename)
		a.Img = "/static/img/" + filename

	}

	o := orm.NewOrm()

	//查询文章类型信息
	atype := models.ArticleType{Id: typeId}
	err = o.Read(&atype)
	if err != nil {
		logs.Info(err.Error())
		c.Data["errmsg"] = "failed to get article type."
		return
	}
	a.Type = &atype

	//这部分应该可以通过session的形式保存的
	user := models.User{Id: uid.(int)}
	err = o.Read(&user)
	if err != nil {
		logs.Info(err.Error())
		c.Data["errmsg"] = "failed to get user."
		return
	}
	a.User = &user

	_, err = o.Insert(&a)
	if err != nil {
		logs.Info("%v,%s,user:%v", a, err.Error(), user)
		c.Data["errmsg"] = "failed to create article."
		return
	}
	c.Redirect("/", 302)

}

func (c *MainController) ShowContent() {

	uid := c.GetSession("userid")
	if uid == nil {
		c.Redirect("/login", 302)
		return
	}

	c.Layout = "layout.html"
	c.TplName = "content.html"
	id, err := c.GetInt("id")
	if err != nil {
		logs.Info("error:%s", err.Error())
		return
	}

	article := models.Article{Id: id}
	o := orm.NewOrm()
	err = o.QueryTable("Article").Filter("Id", id).RelatedSel("Type").One(&article)
	if err != nil {
		logs.Info("read article error:%s", err.Error())
		return
	}

	c.Data["article"] = article
	article.Count = article.Count + 1
	o.Update(&article, "Count")
}

func (c *MainController) ShowUpdate() {

	uid := c.GetSession("userid")
	if uid == nil {
		c.Redirect("/login", 302)
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		logs.Info("failed to get id:%s", err.Error())
		return
	}

	ar := models.Article{Id: id}
	o := orm.NewOrm()
	o.Read(&ar)
	c.Data["article"] = ar
	c.TplName = "update.html"
	c.Layout = "layout.html"

}
func (c *MainController) HandleUpdateArticle() {
	uid := c.GetSession("userid")
	if uid == nil {
		c.Redirect("/login", 302)
		return
	}

	c.TplName = "index.html"
	c.Layout = "layout.html"

	id, err := c.GetInt("id")

	if err != nil {
		logs.Info("failed to get id:%s", err.Error())
		return
	}

	name := c.GetString("articleName")
	content := c.GetString("content")
	if len(name) == 0 || len(content) == 0 {
		c.Data["errmsg"] = "invalid title or content."
		return
	}

	a := models.Article{
		Id: id,
	}

	o := orm.NewOrm()
	err = o.Read(&a)
	if err != nil {
		log.Info("failed to read article id:%d.", id)
	}
	a.Name = name
	a.Content = content

	f, h, err := c.GetFile("uploadname")

	//上传文件失败打印相关信息
	if err != nil && f != nil {
		logs.Info(err.Error())
		c.Data["errmsg"] = "failed to upload file."
		f.Close()
		return
	}

	//保存文件的，并且填充结构体信息
	if f != nil {
		defer f.Close()

		ext := path.Ext(h.Filename)
		if ext != ".jpg" && ext != ".png" {
			logs.Info("wrong file type:%s", ext)
			c.Data["errmsg"] = "wrong file type."
			return
		}
		if h.Size > 500000 {
			logs.Info("size too big:%d", h.Size)
			c.Data["errmsg"] = "too big."
			return
		}

		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

		c.SaveToFile("uploadname", "./static/img/"+filename)
		a.Img = "/static/img/" + filename

	}

	_, err = o.Update(&a)
	if err != nil {
		logs.Info(err.Error())
		c.Data["errmsg"] = "failed to create article."
		return
	}

	c.Redirect("/index", 302)

}
func (c *MainController) HandleDeleteArticle() {

	uid := c.GetSession("userid")
	if uid == nil {
		c.Redirect("/login", 302)
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		logs.Info("failed to get id :%s", err.Error())
	} else {
		ar := models.Article{Id: id}
		o := orm.NewOrm()
		_, err = o.Delete(&ar)

		if err != nil {
			logs.Info("failed to delete article %d, error:%s.", id, err.Error())
		}
	}
	c.Redirect("/index", 302)
}

func (c *MainController) ShowAddType() {

	uid := c.GetSession("userid")
	if uid == nil {
		c.Redirect("/login", 302)
		return
	}

	c.TplName = "addType.html"
	c.Layout = "layout.html"

	o := orm.NewOrm()

	var articleTypes []models.ArticleType
	_, err := o.QueryTable("ArticleType").All(&articleTypes)
	if err != nil {
		logs.Info("failed to read article type data,err:%s", err.Error())
	}
	c.Data["articleTypes"] = articleTypes

}
func (c *MainController) HandleAddType() {

	uid := c.GetSession("userid")
	if uid == nil {
		c.Redirect("/login", 302)
		return
	}

	typeName := c.GetString("typeName")

	if len(typeName) == 0 {
		logs.Info("type name is empty.")
	} else {

		o := orm.NewOrm()
		newType := models.ArticleType{Name: typeName}
		_, err := o.Insert(&newType)
		if err != nil {
			logs.Info("fail to insert new article type,error:%s", err.Error())
		}
	}

	c.Redirect("/addType", 302)

}

func (c *MainController) HandleDeleteType() {

	uid := c.GetSession("userid")
	if uid == nil {
		c.Redirect("/login", 302)
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		logs.Info("failed to get id.")
	} else {
		o := orm.NewOrm()
		atype := models.ArticleType{Id: id}
		_, err = o.Delete(&atype)
		if err != nil {
			logs.Info("failed to delete article type, it's id is :%d.", id)

		}
	}
	c.Redirect("/addType", 302)

}

type LoginController struct {
	beego.Controller
}

func (c *LoginController) Get() {
	c.TplName = "login.html"
}
func (c *LoginController) Post() {

	name := c.GetString("userName")
	pwd := c.GetString("pwd")
	logs.Info("name:%s,pwd:%s", name, pwd)

	if len(name) == 0 || len(pwd) == 0 {
		c.Data["info"] = "invalid user name or password"
		c.TplName = "login.html"
		return
	}

	u := models.User{Name: name}
	err := queryUser(&u)
	if err != nil {
		c.Ctx.WriteString(err.Error())
		return
	}

	if u.Pwd != pwd {
		c.Data["info"] = "wrong password"
		c.TplName = "login.html"
		return
	}

	c.SetSession("userid", u.Id)
	c.Redirect("/index", 302)

}

func (c *LoginController) HandleLogout() {
	c.DelSession("userid")
	c.Redirect("/login", 302)
}

type RegisterController struct {
	beego.Controller
}

func (c *RegisterController) Get() {

	c.TplName = "register.html"
	// c.Redirect("/register", 302)
}

func (c *RegisterController) Post() {
	name := c.GetString("userName")
	pwd := c.GetString("pwd")
	c.TplName = "register.html"

	if len(name) == 0 || len(pwd) == 0 {
		c.Redirect("/register", 302)
		return
	}

	u := models.User{Name: name, Pwd: pwd}
	err := insertUser(&u)
	if err != nil {
		// c.Redirect("/register", 302)
		c.Data["data"] = err.Error()
		return
	}

	c.Redirect("/index", 302)
}

//	err := insertUser("terry", "test")
func insertUser(u *models.User) error {

	o := orm.NewOrm()
	_, err := o.Insert(u)

	if err != nil {
		return err
	}
	return nil
}

// u := models.User{Name: "terry"}
// err := queryUser(&u)
// query user by id or name
func queryUser(u *models.User) error {

	var err error
	o := orm.NewOrm()

	if &u.Id != nil {
		err = o.Read(u)
	}

	if &u.Name != nil {
		err = o.Read(u, "Name")
	}
	return err
}

func updateUser(ou *models.User, nu *models.User) error {

	err := queryUser(ou)
	if err != nil {
		return err
	}
	if &nu.Name != nil {
		ou.Name = nu.Name
	}
	if &nu.Pwd != nil {
		ou.Pwd = nu.Pwd
	}

	o := orm.NewOrm()
	_, err = o.Update(ou)

	return err
}

func deleteUser(u *models.User) error {
	o := orm.NewOrm()
	_, err := o.Delete(u)
	return err
}

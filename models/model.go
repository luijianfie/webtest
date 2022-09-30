package models

import (
	"time"

	"github.com/beego/beego/v2/adapter/orm"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Id       int
	Name     string
	Pwd      string
	Articles []*Article `orm:"reverse(many)"`
}
type Article struct {
	Id      int          `orm:"pk;auto"`
	Name    string       `orm:"size(100)"`
	Time    time.Time    `orm:"auto_now_add;type(datetime)"`
	Count   int          `orm:"default(0)"`
	Content string       `orm:"size(4000)"`
	Img     string       `orm:"size(100)"`
	Type    *ArticleType `orm:"rel(fk)"`
	User    *User        `orm:"rel(fk)"`
}

type ArticleType struct {
	Id       int
	Name     string     `orm:"size(20)"`
	Articles []*Article `orm:"reverse(many)"`
}

func init() {
	orm.RegisterDataBase("default", "mysql", "root:test@tcp(127.0.0.1:3306)/beego?charset=utf8")
	orm.RegisterModel(new(User), new(Article), new(ArticleType))
	orm.RunSyncdb("default", false, true)
}

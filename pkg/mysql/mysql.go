package mysql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Link struct {
	User     string
	Password string
	Ip       string
	Port     string
	Db       string
}

func Start(mysqllink Link) *gorm.DB {
	user := mysqllink.User
	password := mysqllink.Password
	ip := mysqllink.Ip
	port := mysqllink.Port
	db := mysqllink.Db
	dsn := user + ":" + password + "@tcp(" + ip + ":" + port + ")/" + db + "?charset=utf8mb4&parseTime=True&loc=Local"
	mysqlconfig := &gorm.Config{NamingStrategy: schema.NamingStrategy{
		SingularTable: true,
	}}
	dbcli, err := gorm.Open(mysql.Open(dsn), mysqlconfig)

	if err != nil {
		panic("无法连接到数据库：" + err.Error() + dsn)
	}
	return dbcli
}

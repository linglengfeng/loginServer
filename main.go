package main

import (
	"fmt"
	"loginServer/config"
	"loginServer/request"
	"loginServer/shell"
	"loginServer/src/db"
	"loginServer/src/log"
	"loginServer/src/mailer"
	"os"
	"time"
)

func main() {
	log.Start()
	fmt.Println("log start successed...")
	db.Start()
	fmt.Println("db start successed...")
	mailer.Start()
	fmt.Println("sendgrid start successed...")

	if len(os.Args) > 1 && os.Args[1] == "shell" {
		go request.Start() // Gin 在后台运行
		time.Sleep(1 * time.Second)
		node := ""
		host := ""
		if config.Config != nil {
			node = config.Config.GetString("server_name")
			host = config.Config.GetString("gin.ip")
		}
		shell.Start(node, host) // Shell 在前台运行
	} else {
		request.Start() // 正常模式，只运行 Gin
	}
}

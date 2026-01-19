package shell

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Start() {
	fmt.Println("=== adminServer Debug Shell ===")
	fmt.Println("Commands: user, db, config, exit")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("debug> ")
		if !scanner.Scan() {
			break
		}

		cmd := strings.TrimSpace(scanner.Text())
		switch cmd {
		case "user":
			// 调用你的用户相关函数
			testUserFunctions()
		case "db":
			// 测试数据库连接
			fmt.Println("db")
		case "config":
			// 显示配置
			fmt.Println("config")
		case "exit":
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}

func testUserFunctions() {
	// 在这里调用你的业务函数
	fmt.Println("Testing user functions...")
	// user.Create()
	// user.FindByID(1)
}

package shell

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chzyer/readline"
)

// showBreakMenuAndHandle 模拟 Erlang 风格的 BREAK 菜单：
// - a：退出 shell（abort）
// - c：继续（continue）
// 返回 true 表示需要退出 shell
func showBreakMenuAndHandle(rl *readline.Instance) bool {
	fmt.Println("BREAK: (a)bort (c)ontinue")

	for {
		rl.SetPrompt("1> ")
		line, err := rl.Readline()
		if err != nil {
			// 再次 Ctrl+C 或 EOF：直接退出 shell
			return true
		}
		s := strings.TrimSpace(line)
		if s == "" {
			continue
		}
		switch strings.ToLower(s) {
		case "c", "continue":
			return false
		case "a", "abort", "q", "quit":
			return true
		default:
			fmt.Println("请输入 a/abort 退出，或输入 c/continue 继续。")
		}
	}
}

// handleMetaCommand 处理类似 gore 的冒号命令
// 返回 true 表示已处理（不再走普通代码执行）
func handleMetaCommand(line string) bool {
	switch {
	case line == ":help" || line == ":h":
		fmt.Print(`可用命令:
  :h, :help              查看帮助
  :q, :quit              退出
  :import <pkg> [path]   导入包（示例: :import fmt | :import json encoding/json）
  :func <decl>           定义函数（示例: :func add(a,b int) int { return a+b }）
  :history [n]           查看历史输入（默认 50 条）
  :println(<expr...>)    打印表达式（示例: :println("Hello")）
  :pwd                   输出当前模块根目录
  :cd <path>             切换工作目录并重新加载模块/解释器
  :timeout <ms>          设置 go run 调用超时（毫秒，示例: :timeout 5000）

提示:
  - 启动时会自动构造 GOPATH 映射，以支持 go modules 项目本地包 import
  - 若本地包 import 被第三方依赖卡住：可用 pkg.CONST（源码解析）或 pkg.Func(...)（go run 执行）
`)
		return true

	case line == ":history" || strings.HasPrefix(line, ":history "):
		n := 50
		if strings.HasPrefix(line, ":history ") {
			v := strings.TrimSpace(strings.TrimPrefix(line, ":history "))
			if v != "" {
				if nn, err := strconv.Atoi(v); err != nil || nn <= 0 {
					fmt.Println("用法: :history [n]  例如 :history 100")
					return true
				} else {
					n = nn
				}
			}
		}

		lines := getHistorySnapshot()
		if len(lines) == 0 {
			fmt.Println("(暂无历史记录)")
			return true
		}
		start := 0
		if n > 0 && len(lines) > n {
			start = len(lines) - n
		}
		for i := start; i < len(lines); i++ {
			// 统一显示 1-based 序号
			fmt.Printf("%d  %s\n", i+1, lines[i])
		}
		return true

	case line == ":pwd":
		if modRoot == "" {
			fmt.Println("(modRoot 未初始化)")
			return true
		}
		fmt.Println(modRoot)
		return true

	case strings.HasPrefix(line, ":cd "):
		target := strings.TrimSpace(strings.TrimPrefix(line, ":cd "))
		if target == "" {
			fmt.Println("用法: :cd <path>")
			return true
		}
		if err := os.Chdir(target); err != nil {
			fmt.Printf("切换目录失败: %v\n", err)
			return true
		}
		initInterpreter()
		fmt.Printf("已切换到: %s\n", target)
		return true

	case strings.HasPrefix(line, ":timeout "):
		v := strings.TrimSpace(strings.TrimPrefix(line, ":timeout "))
		ms, err := strconv.Atoi(v)
		if err != nil || ms <= 0 {
			fmt.Println("用法: :timeout <ms>  例如 :timeout 5000")
			return true
		}
		goRunTimeout = time.Duration(ms) * time.Millisecond
		fmt.Printf("go run 超时已设置为: %v\n", goRunTimeout)
		return true

	case strings.HasPrefix(line, ":import "):
		if interpreter == nil {
			fmt.Println("错误: 解释器未初始化")
			return true
		}
		args := strings.Fields(strings.TrimSpace(strings.TrimPrefix(line, ":import ")))
		if len(args) == 0 {
			fmt.Println("用法: :import <pkg> [path]  例如 :import fmt 或 :import json encoding/json")
			return true
		}

		var stmt string
		if len(args) == 1 {
			pkg := args[0]
			if !strings.Contains(pkg, "/") && isLocalTopPackageDir(pkg) && modName != "" {
				stmt = fmt.Sprintf(`import %s "%s/%s"`, pkg, modName, pkg)
				markImported(pkg)
			} else {
				stmt = fmt.Sprintf(`import "%s"`, pkg)
			}
		} else {
			alias := args[0]
			path := args[1]
			stmt = fmt.Sprintf(`import %s "%s"`, alias, path)
			markImported(alias)
		}

		if _, err := interpreter.Eval(stmt); err != nil {
			fmt.Printf("导入失败: %v\n", err)
			fmt.Println("提示: 若本地包 import 被第三方依赖卡住，可：")
			fmt.Println("  - 读取常量/字面量：pkg.NAME   (例如 request.STATUS)")
			fmt.Println("  - 调用函数：pkg.Func(...)     (例如 request.Stop())")
		} else {
			fmt.Println("导入成功")
		}
		return true

	case strings.HasPrefix(line, ":func "):
		decl := strings.TrimSpace(strings.TrimPrefix(line, ":func "))
		if decl == "" {
			fmt.Println("用法: :func <decl>  例如 :func add(a,b int) int { return a+b }")
			return true
		}
		if !strings.HasPrefix(strings.TrimSpace(decl), "func ") {
			decl = "func " + decl
		}
		// :func 定义函数，不记录结果，cmdNo 传 0
		executeCode(decl, 0)
		return true

	case strings.HasPrefix(line, ":println"):
		rest := strings.TrimSpace(strings.TrimPrefix(line, ":println"))
		if rest == "" {
			fmt.Println()
			return true
		}
		if strings.HasPrefix(rest, "(") {
			executeCode("fmt.Println"+rest, 0)
			return true
		}
		executeCode("fmt.Println("+rest+")", 0)
		return true
	}

	return false
}

package shell

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
)

// Start 启动交互式 Shell，类似 Erlang shell
func Start(node, host string) {
	fmt.Println("=== Go Interactive Shell ===")
	fmt.Println(":(h)elp 查看用法；:(q)uit 退出。")
	fmt.Println()

	// 初始化解释器
	initInterpreter()

	var input strings.Builder
	var inMultiline bool
	cmdNo := 1

	// prompt 规则：
	// - node 和 host 都为空：显示 1> / 2>...
	// - 只要有一个不为空：显示 (node@host)1> / (node@host)2>...；缺省项用默认值补齐
	showNodeHost := strings.TrimSpace(node) != "" || strings.TrimSpace(host) != ""
	if showNodeHost {
		if strings.TrimSpace(node) == "" {
			node = "node"
		}
		if strings.TrimSpace(host) == "" {
			host = "127.0.0.1"
		}
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt: buildPrompt(showNodeHost, node, host, cmdNo, inMultiline),
		// 不落盘：history 只保存在内存中（readline 自带 session 内历史）
		HistoryFile:     "",
		InterruptPrompt: "^C",
		EOFPrompt:       "quit",
	})
	if err != nil {
		fmt.Printf("错误: 初始化 readline 失败: %v\n", err)
		return
	}
	defer rl.Close()

	for {
		rl.SetPrompt(buildPrompt(showNodeHost, node, host, cmdNo, inMultiline))

		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				// Ctrl+C：进入 Erlang 风格 BREAK 菜单
				input.Reset()
				inMultiline = false
				fmt.Println()
				if quit := showBreakMenuAndHandle(rl); quit {
					fmt.Println("Quit!")
					return
				}
				continue
			}
			// EOF（Ctrl+D）或其他错误：退出
			break
		}
		trimmedLine := strings.TrimSpace(line)
		// 记录原始输入到内存 history（包含冒号命令）
		if trimmedLine != "" {
			addHistoryLine(line)
		}

		// 处理退出命令
		if trimmedLine == ":quit" || trimmedLine == ":q" {
			fmt.Println("Quit!")
			return
		}

		// 冒号命令
		if strings.HasPrefix(trimmedLine, ":") {
			if handled := handleMetaCommand(trimmedLine); handled {
				cmdNo++
				continue
			}
		}

		// 多行输入（以反斜杠结尾表示继续）
		if strings.HasSuffix(trimmedLine, "\\") {
			input.WriteString(strings.TrimSuffix(trimmedLine, "\\"))
			input.WriteString("\n")
			inMultiline = true
			continue
		}

		// 添加当前行
		if inMultiline {
			input.WriteString(trimmedLine)
			inMultiline = false
		} else {
			input.WriteString(trimmedLine)
		}

		code := input.String()
		input.Reset()

		if code == "" {
			continue
		}

		// 先处理 :v(n) 引用，例如 request.TestAdd(:v(1), 4)
		expanded, ok := expandValueRefs(code)
		if !ok {
			// 存在无效的 :v(n) 引用，本次命令不再执行，但行号依然递增
			cmdNo++
			continue
		}

		executeCode(expanded, cmdNo)
		cmdNo++
	}
}

func buildPrompt(showNodeHost bool, node, host string, cmdNo int, inMultiline bool) string {
	if inMultiline {
		return "... "
	}
	if !showNodeHost {
		return fmt.Sprintf("%d> ", cmdNo)
	}
	return fmt.Sprintf("(%s@%s)%d> ", node, host, cmdNo)
}

func TestAdd(intList ...int) int {
	sum := 0
	for _, v := range intList {
		sum += v
	}
	return sum
}

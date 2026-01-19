package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// tryRunLocalCallWithGo 尝试用 `go run` 执行本地包的函数调用（例如 request.Stop()）。
// 返回值：
// - string：标准输出中（已去除尾部换行）的内容
// - bool：若为 true 表示已处理（无论成功或失败，错误信息已输出）；false 表示该表达式不适用于 go run 处理。
func tryRunLocalCallWithGo(expr string) (string, bool) {
	m := localCallRe.FindStringSubmatch(strings.TrimSpace(expr))
	if len(m) != 4 {
		return "", false
	}
	alias := m[1]
	funcName := m[2]
	args := strings.TrimSpace(m[3])

	if !isLocalTopPackageDir(alias) || modRoot == "" || modName == "" {
		return "", false
	}

	importPath := fmt.Sprintf("%s/%s", modName, alias)
	prog := buildGoRunProgram(importPath, alias, funcName, args)
	out, runErr := runGoSnippet(prog)
	if runErr != nil {
		fmt.Printf("错误: %v\n", runErr)
		if strings.TrimSpace(out) != "" {
			fmt.Println(strings.TrimRight(out, "\r\n"))
		}
		return strings.TrimRight(out, "\r\n"), true
	}
	if strings.TrimSpace(out) != "" {
		fmt.Println(strings.TrimRight(out, "\r\n"))
	}
	return strings.TrimRight(out, "\r\n"), true
}

func buildGoRunProgram(importPath, alias, funcName, args string) string {
	call := fmt.Sprintf("%s.%s(%s)", alias, funcName, args)
	return fmt.Sprintf(`package main

import (
	"fmt"
	%[2]s "%[1]s"
)

func main() {
	// 先按“有返回值”生成，若无返回值会编译失败，外层会自动 fallback
	var v any = %[3]s
	if v != nil {
		fmt.Print(v)
	}
}
`, importPath, alias, call)
}

func runGoSnippet(program string) (string, error) {
	if modRoot == "" {
		return "", fmt.Errorf("未定位到 go.mod 根目录")
	}
	goRunMu.Lock()
	defer goRunMu.Unlock()
	if goRunDir == "" {
		td, err := os.MkdirTemp("", "loginServer-shell-go-*")
		if err != nil {
			return "", err
		}
		goRunDir = td
		goRunFile = filepath.Join(goRunDir, "main.go")
	}
	mainFile := goRunFile
	if err := os.WriteFile(mainFile, []byte(program), 0o644); err != nil {
		return "", err
	}

	out, err := runGo(mainFile)
	if err == nil {
		return out, nil
	}

	// fallback：无返回值版本
	prog2 := regexp.MustCompile(`(?s)var v any = .*?\n\s*if v != nil \{\n\s*fmt\.Print\(v\)\n\s*\}\n`).ReplaceAllString(program, "")
	callRe := regexp.MustCompile(`var v any = (.*)\n`)
	m := callRe.FindStringSubmatch(program)
	call := ""
	if len(m) == 2 {
		call = strings.TrimSpace(m[1])
	}
	if call != "" {
		prog2 = strings.Replace(prog2, "func main() {", "func main() {\n\t"+call, 1)
	}
	if err := os.WriteFile(mainFile, []byte(prog2), 0o644); err != nil {
		return "", err
	}
	out2, err2 := runGo(mainFile)
	if err2 != nil {
		return out2, err
	}
	return out2, nil
}

func runGo(mainFile string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), goRunTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", mainFile)
	cmd.Dir = modRoot
	env := os.Environ()
	env = append(env, "GO111MODULE=on")
	cmd.Env = env
	b, err := cmd.CombinedOutput()
	out := string(b)
	if ctx.Err() == context.DeadlineExceeded {
		return out, fmt.Errorf("go run 超时: %v", goRunTimeout)
	}
	if err != nil {
		return out, err
	}
	return out, nil
}


# Go Shell（交互式 REPL）

这是一个可嵌入到任意 Go 工程中的交互式 Shell（REPL），用于在本地快速执行 Go 表达式、调用工程内包函数、排查问题。

Interactive Go shell inspired by the Erlang shell.

底层主要使用 [yaegi](https://github.com/traefik/yaegi) 解释执行；当遇到本地包依赖复杂导致 yaegi import 失败时，会自动回退到 `go run` 执行本地包函数调用。

## 作为包集成（宿主程序启动）

在你的宿主程序中按需启动 Shell（示例）：

```go
package main

import "yourmodule/shell"

func main() {
	// node/host 仅用于显示提示符，可传空字符串
	shell.Start("node", "127.0.0.1")
}
```

启动后提示符示例：

```text
(node@127.0.0.1)1>
```

## 基本用法

- **表达式**：直接输入表达式，会打印结果，并记录到该行的结果缓存（用于 `:v(n)`）。

```text
1> 1 + 2
3
```

- **语句**：如 `var/const/type/func`、赋值、`fmt.Print*` 等，会执行但不记录结果。

```text
2> x := 10
3> x
10
```

- **多行输入**：行尾加 `\` 续行：

```text
4> func add(a, b int) int { \
... return a + b }
```

## `:v(n)` 引用上一行结果

支持在后续输入中通过 `:v(n)` 引用第 `n` 行表达式的结果（以文本形式替换）。

示例：

```text
(node@127.0.0.1)1> request.TestAdd(1,2,3)
6
(node@127.0.0.1)2> request.TestAdd(:v(1),4)
10
```

说明：
- `:v(1)` 会在执行前被替换成第一行保存的结果 `6`
- 因此 `request.TestAdd(:v(1),4)` 等价于 `request.TestAdd(6,4)`
- 若引用的行不存在结果，会提示错误并跳过执行该行

## 冒号命令（Meta Commands）

在提示符下可使用以下命令：

- **`:h` / `:help`**：查看帮助
- **`:q` / `:quit`**：退出
- **`:import <pkg> [path]`**：导入包  
  - 示例：`:import fmt`  
  - 示例：`:import json encoding/json`
- **`:func <decl>`**：定义函数  
  - 示例：`:func add(a,b int) int { return a+b }`
- **`:println(<expr...>)`**：打印表达式（等价于 `fmt.Println(...)`）
- **`:pwd`**：输出当前 `go.mod` 根目录
- **`:cd <path>`**：切换工作目录并重新加载模块/解释器
- **`:timeout <ms>`**：设置 `go run` 回退执行的超时（毫秒），例如 `:timeout 5000`
- **`:history [n]`**：查看历史输入（默认最近 50 条）

## 本地包调用规则（yaegi vs `go run`）

当你输入类似下面这种“本地包函数调用”：

```text
request.TestAdd(1,2,3)
```

执行流程大致是：

1. 先尝试 yaegi 直接 `Eval` 该表达式（更快、无需编译）。
2. 若失败：
   - 若是 `pkg.NAME` 这种简单字面量，会尝试从源码解析 `const/var`（仅基础字面量）。
   - 若是 `pkg.Func(...)` 且 `pkg` 是工程根目录下的本地包，会回退用 `go run` 生成临时程序执行，并打印输出。

> 这也是为什么你会看到某些本地包调用能跑起来，即使 yaegi import 被第三方依赖卡住。

## History（仅内存，不落盘）

Shell 的历史记录**只保存在内存中**：

- 退出 shell 后历史会自动清空
- `:history` 读取的是内存中的历史列表

## Windows 注意事项（GOPATH 映射）

为了让 yaegi 更容易导入 go modules 下的本地包，启动时会在临时目录构造一个 GOPATH 映射（Windows 下使用 `mklink /J` 建立目录联接）。如果你看到映射相关警告，通常意味着：

- 当前目录向上未找到 `go.mod`，或
- 权限不足导致 `mklink` 失败

可通过 `:pwd` / `:cd` 切换到正确目录后重新初始化。


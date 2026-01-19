package shell

import "sync"

// 每一行（命令序号）对应的执行结果文本，仅记录“表达式”结果，用于 :v(n) 引用。
// 下标使用 1-based 行号。
var (
	lineResultsMu sync.RWMutex
	lineResults   = map[int]string{}
)

func setLineResult(line int, value string) {
	if line <= 0 {
		return
	}
	lineResultsMu.Lock()
	lineResults[line] = value
	lineResultsMu.Unlock()
}

func getLineResult(line int) (string, bool) {
	if line <= 0 {
		return "", false
	}
	lineResultsMu.RLock()
	defer lineResultsMu.RUnlock()
	v, ok := lineResults[line]
	return v, ok
}


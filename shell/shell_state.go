package shell

import (
	"sync"
	"time"

	"github.com/traefik/yaegi/interp"
)

// 运行时状态（同包内共享）
var (
	interpreter *interp.Interpreter
	modRoot     string
	modName     string
	goPathRoot  string

	importMu      sync.Mutex
	importedAlias = map[string]bool{}

	goRunTimeout = 10 * time.Second

	goRunMu   sync.Mutex
	goRunDir  string
	goRunFile string

	// history 仅保存在内存中（不落盘）
	historyMu    sync.RWMutex
	historyLines []string
)

func addHistoryLine(s string) {
	historyMu.Lock()
	historyLines = append(historyLines, s)
	historyMu.Unlock()
}

func getHistorySnapshot() []string {
	historyMu.RLock()
	defer historyMu.RUnlock()
	out := make([]string, len(historyLines))
	copy(out, historyLines)
	return out
}


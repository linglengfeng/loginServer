package shell

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	hooksMu sync.RWMutex
	hooks   = map[string]reflect.Value{}
)

// RegisterFunc 注册一个可在 REPL 中通过 `shell.<name>(...)` 调用的函数。
//
// 例如：
//   shell.RegisterFunc("stopHttp", request.Stop)
//   在 REPL 中：shell.stopHttp()
func RegisterFunc(name string, fn any) {
	if name == "" {
		panic("shell.RegisterFunc: empty name")
	}
	v := reflect.ValueOf(fn)
	if !v.IsValid() || v.Kind() != reflect.Func {
		panic(fmt.Sprintf("shell.RegisterFunc: %s is not a function", name))
	}

	hooksMu.Lock()
	hooks[name] = v
	hooksMu.Unlock()
}

// ClearFuncs 清空所有已注册的函数（可用于测试或重载）。
func ClearFuncs() {
	hooksMu.Lock()
	hooks = map[string]reflect.Value{}
	hooksMu.Unlock()
}

// RegisterFuncs 批量注册函数。
func RegisterFuncs(m map[string]any) {
	for k, v := range m {
		RegisterFunc(k, v)
	}
}

// getHookSymbols 给 yaegi 注入符号用（内部使用）。
func getHookSymbols() map[string]reflect.Value {
	hooksMu.RLock()
	defer hooksMu.RUnlock()

	out := make(map[string]reflect.Value, len(hooks))
	for k, v := range hooks {
		out[k] = v
	}
	return out
}


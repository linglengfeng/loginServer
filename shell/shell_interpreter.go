package shell

import (
	"fmt"
	"reflect"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// initInterpreter 初始化 Yaegi 解释器
func initInterpreter() {
	var err error
	modRoot, modName, err = findModuleRootAndName()
	if err != nil {
		fmt.Printf("警告: 未能定位 go.mod（本地包 import 可能不可用）: %v\n", err)
	}
	goPathRoot, err = ensureGoPathMapping(modRoot, modName)
	if err != nil {
		fmt.Printf("警告: GOPATH 映射初始化失败（本地包 import 可能不可用）: %v\n", err)
	}

	i := interp.New(interp.Options{GoPath: goPathRoot})
	i.Use(stdlib.Symbols)

	prelude := `
package main

import (
	"fmt"
	"reflect"
)

// 全局函数定义（通过解释器注册）
`
	if _, err := i.Eval(prelude); err != nil {
		fmt.Printf("警告: 预定义代码执行失败: %v\n", err)
	}

	i.Use(map[string]map[string]reflect.Value{
		"shell": getHookSymbols(),
	})

	interpreter = i
}


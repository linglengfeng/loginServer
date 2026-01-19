package shell

import (
	"fmt"
	"reflect"
	"strings"
)

// executeCode 执行代码并输出结果。
// cmdNo 为当前命令行号（1-based）；当 cmdNo<=0 时，不记录结果（用于内部调用）。
func executeCode(code string, cmdNo int) {
	if interpreter == nil {
		fmt.Println("错误: 解释器未初始化")
		return
	}

	trimmedCode := strings.TrimSpace(code)
	if trimmedCode == "" {
		return
	}

	autoImportForCode(trimmedCode)

	// 判断是否为语句（包含分号、赋值、函数定义等）
	isStatement := strings.Contains(code, ";") ||
		strings.Contains(code, ":=") ||
		strings.Contains(code, "=") ||
		strings.HasPrefix(trimmedCode, "func ") ||
		strings.HasPrefix(trimmedCode, "var ") ||
		strings.HasPrefix(trimmedCode, "const ") ||
		strings.HasPrefix(trimmedCode, "type ") ||
		strings.HasPrefix(trimmedCode, "package ") ||
		strings.Contains(code, "fmt.Print") ||
		strings.Contains(code, "println(") ||
		strings.Contains(code, "Println(")

	if isStatement {
		if _, err := interpreter.Eval(code); err != nil {
			fmt.Printf("错误: %v\n", err)
		}
		return
	}

	// 表达式：优先尝试直接通过 yaegi 计算并打印/记录结果
	v, err := interpreter.Eval(code)
	if err == nil {
		if v.IsValid() && v.Kind() != reflect.Invalid {
			val := v.Interface()
			fmt.Println(val)
			if cmdNo > 0 {
				setLineResult(cmdNo, fmt.Sprint(val))
			}
		}
		return
	}

	// 若解释失败，尝试解析本地包字面量（const/var）
	if vv, ok := tryResolveLocalLiteral(trimmedCode); ok {
		fmt.Println(vv)
		if cmdNo > 0 {
			setLineResult(cmdNo, vv)
		}
		return
	}

	// 尝试通过 go run 调用本地包函数（例如 request.TestAdd(1,2,3)）
	if out, ok := tryRunLocalCallWithGo(trimmedCode); ok {
		if cmdNo > 0 && strings.TrimSpace(out) != "" {
			setLineResult(cmdNo, strings.TrimSpace(out))
		}
		return
	}

	// 最后再尝试一次直接执行（可能是语句或无返回值的调用等）
	if _, err2 := interpreter.Eval(code); err2 != nil {
		fmt.Printf("错误: %v\n", err)
	}
}


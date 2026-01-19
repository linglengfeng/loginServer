package shell

import (
	"fmt"
	"regexp"
	"strconv"
)

// :v(n) 引用模式，例如 :v(1)、:v(2)
var valueRefRe = regexp.MustCompile(`:v\((\d+)\)`)

// expandValueRefs 将代码中的 :v(n) 替换为对应行的结果文本。
// 若存在无效引用，会打印错误并返回 ok=false。
func expandValueRefs(code string) (out string, ok bool) {
	var errMsg string
	out = valueRefRe.ReplaceAllStringFunc(code, func(m string) string {
		sub := valueRefRe.FindStringSubmatch(m)
		if len(sub) != 2 {
			return m
		}
		n, _ := strconv.Atoi(sub[1])
		if n <= 0 {
			errMsg = fmt.Sprintf("错误: :v(%d) 行号必须为正整数", n)
			return m
		}
		val, found := getLineResult(n)
		if !found {
			errMsg = fmt.Sprintf("错误: 未找到第 %d 行的结果 (:v(%d))", n, n)
			return m
		}
		return val
	})
	if errMsg != "" {
		fmt.Println(errMsg)
		return "", false
	}
	return out, true
}


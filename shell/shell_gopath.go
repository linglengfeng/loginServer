package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func findModuleRootAndName() (root string, module string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	cur := wd
	for {
		gomod := filepath.Join(cur, "go.mod")
		b, rerr := os.ReadFile(gomod)
		if rerr == nil {
			lines := strings.Split(string(b), "\n")
			for _, ln := range lines {
				ln = strings.TrimSpace(ln)
				if strings.HasPrefix(ln, "module ") {
					return cur, strings.TrimSpace(strings.TrimPrefix(ln, "module ")), nil
				}
			}
			return cur, "", fmt.Errorf("go.mod 未找到 module 行")
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", "", fmt.Errorf("未找到 go.mod（从 %s 向上）", wd)
		}
		cur = parent
	}
}

func ensureGoPathMapping(moduleRoot, moduleName string) (string, error) {
	if moduleRoot == "" || moduleName == "" {
		return "", fmt.Errorf("moduleRoot 或 moduleName 为空")
	}
	tmp := filepath.Join(os.TempDir(), "loginServer-gopath-cache")
	srcDir := filepath.Join(tmp, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		return "", err
	}
	dst := filepath.Join(srcDir, filepath.FromSlash(moduleName))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		if _, err := os.Stat(dst); err == nil {
			return tmp, nil
		}
		cmd := exec.Command("cmd", "/C", "mklink", "/J", dst, moduleRoot)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("mklink /J 失败: %v, 输出: %s", err, strings.TrimSpace(string(out)))
		}
	} else {
		if _, err := os.Lstat(dst); err == nil {
			return tmp, nil
		}
		if err := os.Symlink(moduleRoot, dst); err != nil {
			return "", fmt.Errorf("symlink 失败: %v", err)
		}
	}
	return tmp, nil
}


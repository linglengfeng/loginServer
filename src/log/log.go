package log

import (
	"fmt"
	"loginServer/config"
	"loginServer/pkg/logger"
)

const (
	// DefaultCaller 默认调用栈深度
	// 调用链：业务代码 -> src/log -> pkg/logger -> runtime.Caller
	// runtime.Caller(3) 通常能定位到业务调用点（例如 main.go / handler.go）
	DefaultCaller = 3
)

// Start 初始化日志系统，从配置文件中读取日志配置并启动
func Start() error {
	// 允许通过配置覆盖
	ioWay := config.Config.GetInt("log.ioway")
	if ioWay == 0 {
		ioWay = logger.IoWayFileCtl
	}
	caller := config.Config.GetInt("log.caller")
	if caller == 0 {
		caller = DefaultCaller
	}

	logcfg := logger.LogCfg{
		Loglv:      config.Config.GetString("log.level"),
		Remain_day: config.Config.GetInt("log.remain_day"),
		Path:       config.Config.GetString("log.path"),
		Ioway:      ioWay,
		Caller:     caller,
		Showfile:   config.Config.GetInt("log.showfile"),
		Showfunc:   config.Config.GetInt("log.showfunc"),
	}
	if err := logger.Start(logcfg); err != nil {
		return fmt.Errorf("logger start failed: %w", err)
	}
	return nil
}

// Debug 输出调试级别日志
func Debug(format string, a ...any) {
	logger.Debug(format, a...)
}

// Info 输出信息级别日志
func Info(format string, a ...any) {
	logger.Info(format, a...)
}

// Warn 输出警告级别日志
func Warn(format string, a ...any) {
	logger.Warn(format, a...)
}

// Error 输出错误级别日志
func Error(format string, a ...any) {
	logger.Error(format, a...)
}

package log

import (
	"loginServer/config"
	"loginServer/pkg/logger"
)

const (
	// DefaultCaller 默认调用栈深度
	DefaultCaller = 3
)

// Start 初始化日志系统，从配置文件中读取日志配置并启动
func Start() {
	logcfg := logger.LogCfg{
		Loglv:      config.Config.GetString("log.level"),
		Remain_day: config.Config.GetInt("log.remain_day"),
		Path:       config.Config.GetString("log.path"),
		Ioway:      logger.IoWayFileCtl,
		Caller:     DefaultCaller,
		Showfile:   config.Config.GetInt("log.showfile"),
		Showfunc:   config.Config.GetInt("log.showfunc"),
	}
	logger.Start(logcfg)
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

package logger

import (
	"fmt"

	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"golang.org/x/exp/slog"
)

const (
	IoWayFile    = 1
	IoWayCtl     = 2
	IoWayFileCtl = 3
	NoCaller     = -1

	prefixDebug = "log_debug_"
	prefixInfo  = "log_info_"
	prefixWarn  = "log_warn_"
	prefixError = "log_error_"
	levelDebug  = "debug"
	levelInfo   = "info"
	levelWarn   = "warn"
	levelError  = "error"
)

type Logger struct {
	level    string
	self_lv  string
	path     string
	filename string
	time     time.Time
	fd       *os.File
	logger   *slog.Logger
}

var (
	DEBUG         *Logger
	INFO          *Logger
	WARN          *Logger
	ERROR         *Logger
	fileLock      sync.RWMutex
	levelFileList = [4]string{prefixDebug, prefixInfo, prefixWarn, prefixError}
	LogCfgValue   LogCfg
	cronJob       *cron.Cron
)

type formatLogInfo struct {
	msg      string
	file     string
	funcName string
}

type LogCfg struct {
	Loglv      string
	Remain_day int
	Path       string
	Ioway      int
	Caller     int
	Showfile   int
	Showfunc   int
}

// Start 初始化日志系统
func Start(logcfg LogCfg) error {
	LogCfgValue = logcfg
	if err := createLogFile(); err != nil {
		return err
	}
	logLevel := LogCfgValue.Loglv
	logFilePath := LogCfgValue.Path
	var err error
	if DEBUG, err = NewLogger(logLevel, levelDebug, logFilePath); err != nil {
		return err
	}
	if INFO, err = NewLogger(logLevel, levelInfo, logFilePath); err != nil {
		return err
	}
	if WARN, err = NewLogger(logLevel, levelWarn, logFilePath); err != nil {
		return err
	}
	if ERROR, err = NewLogger(logLevel, levelError, logFilePath); err != nil {
		return err
	}
	startLogJob()
	return nil
}

// NewLogger 创建新的日志记录器
func NewLogger(level string, selfLv string, path string) (*Logger, error) {
	now := time.Now()
	postFix := now.Format("20060102")
	prefix := getPrefixByLevel(selfLv)
	filename := logfilename(LogCfgValue.Path, prefix, postFix)

	l := &Logger{time: now, level: level, path: path, filename: filename, self_lv: selfLv}
	f, err := os.OpenFile(l.filename, os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0755)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	iow := getWriterByIoWay(LogCfgValue.Ioway, f)
	logger := createSlogLogger(level, iow)
	l.logger = logger
	l.fd = f
	return l, nil
}

// getPrefixByLevel 根据日志级别获取文件前缀
func getPrefixByLevel(level string) string {
	switch level {
	case levelDebug:
		return prefixDebug
	case levelInfo:
		return prefixInfo
	case levelWarn:
		return prefixWarn
	case levelError:
		return prefixError
	default:
		return prefixDebug
	}
}

// getWriterByIoWay 根据IO方式获取Writer
func getWriterByIoWay(ioWay int, file *os.File) io.Writer {
	switch ioWay {
	case IoWayFile:
		return file
	case IoWayCtl:
		return os.Stdout
	case IoWayFileCtl:
		return io.MultiWriter(os.Stdout, file)
	default:
		return io.MultiWriter(os.Stdout, file)
	}
}

// createSlogLogger 根据日志级别创建slog.Logger
func createSlogLogger(level string, writer io.Writer) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case levelDebug:
		logLevel = slog.LevelDebug
	case levelInfo:
		logLevel = slog.LevelInfo
	case levelWarn:
		logLevel = slog.LevelWarn
	case levelError:
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{Level: logLevel}))
}

// GetLogger 根据日志级别获取日志记录器
func GetLogger(logLv string) *Logger {
	fileLock.Lock()
	defer fileLock.Unlock()

	var loggerInfo *Logger
	switch logLv {
	case levelDebug:
		loggerInfo = DEBUG
	case levelInfo:
		loggerInfo = INFO
	case levelWarn:
		loggerInfo = WARN
	case levelError:
		loggerInfo = ERROR
	default:
		loggerInfo = DEBUG
	}

	// 日切轮转：创建新 logger，并关闭旧文件句柄，避免 fd 泄露
	if loggerInfo != nil && !isOneDay(loggerInfo.time) {
		old := loggerInfo
		newLogger, err := NewLogger(old.level, old.self_lv, old.path)
		if err == nil {
			loggerInfo = newLogger
		} else {
			// 轮转失败时继续用旧 logger，避免影响业务打印日志
			loggerInfo = old
		}
		switch logLv {
		case levelDebug:
			DEBUG = loggerInfo
		case levelInfo:
			INFO = loggerInfo
		case levelWarn:
			WARN = loggerInfo
		case levelError:
			ERROR = loggerInfo
		default:
			DEBUG = loggerInfo
		}
		if old.fd != nil {
			_ = old.fd.Sync()
			_ = old.fd.Close()
		}
	}
	if loggerInfo == nil {
		// 未初始化时兜底：输出到 stdout
		return &Logger{logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))}
	}
	return loggerInfo
}

// Debug 输出调试级别日志
func Debug(format string, a ...any) {
	formatLogInfo := Format(format, a...)
	loggerInfo := GetLogger(levelDebug)
	loggerInfo.logger.Debug(formatLogInfo.msg, slog.String("file", formatLogInfo.file), slog.String("func", formatLogInfo.funcName))
}

// Info 输出信息级别日志
func Info(format string, a ...any) {
	formatLogInfo := Format(format, a...)
	loggerInfo := GetLogger(levelInfo)
	loggerInfo.logger.Info(formatLogInfo.msg, slog.String("file", formatLogInfo.file), slog.String("func", formatLogInfo.funcName))
}

// Warn 输出警告级别日志
func Warn(format string, a ...any) {
	formatLogInfo := Format(format, a...)
	loggerInfo := GetLogger(levelWarn)
	loggerInfo.logger.Warn(formatLogInfo.msg, slog.String("file", formatLogInfo.file), slog.String("func", formatLogInfo.funcName))
}

// Error 输出错误级别日志
func Error(format string, a ...any) {
	formatLogInfo := Format(format, a...)
	loggerInfo := GetLogger(levelError)
	loggerInfo.logger.Error(formatLogInfo.msg, slog.String("file", formatLogInfo.file), slog.String("func", formatLogInfo.funcName))
}

// Format 格式化日志信息
func Format(format string, a ...any) formatLogInfo {
	msgStr := fmt.Sprintf(format, a...)
	codeStr := "-"
	funcName := "-"
	caller := LogCfgValue.Caller
	if caller != NoCaller && caller > 0 {
		pc, codePath, codeLine, ok := runtime.Caller(caller)
		if ok {
			_, fileName := filepath.Split(codePath)
			if LogCfgValue.Showfile != 0 {
				codeStr = fmt.Sprintf("%s:%d", fileName, codeLine)
			}
			if LogCfgValue.Showfunc != 0 {
				funcName = runtime.FuncForPC(pc).Name()
			}
		}
	}
	return formatLogInfo{msg: msgStr, file: codeStr, funcName: funcName}
}

// startLogJob 启动定时任务：创建新日志文件、删除过期日志
func startLogJob() {
	fileLock.Lock()
	defer fileLock.Unlock()
	if cronJob != nil {
		return
	}
	c := cron.New(cron.WithSeconds())
	spec := "@daily"
	c.AddFunc(spec, func() {
		Warn("执行log定时任务。。。")
		now := time.Now()
		_ = createLogFile()

		// 删除 n 天前的日志（文件句柄关闭由 GetLogger 轮转时处理）
		for _, filePrefix := range levelFileList {
			deleteDay := LogCfgValue.Remain_day
			removeLogFile := logfilename(LogCfgValue.Path, filePrefix,
				now.Add(time.Duration(-deleteDay)*24*time.Hour).Format("20060102"))
			if err := os.Remove(removeLogFile); err != nil && !os.IsNotExist(err) {
				Error("删除日志文件失败: %s, error: %v", removeLogFile, err)
			}
		}
	})
	c.Start()
	cronJob = c
}

// IsExist 检查文件或目录是否存在
func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil
}

func isDir(fileAddr string) bool {
	s, err := os.Stat(fileAddr)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// createLogFile 创建日志文件
func createLogFile() error {
	fileLock.Lock()
	defer fileLock.Unlock()
	now := time.Now()
	postFix := now.Format("20060102")
	logFilePath := LogCfgValue.Path
	if !isDir(logFilePath) {
		if err := os.MkdirAll(logFilePath, 0755); err != nil {
			return fmt.Errorf("创建日志目录失败: %w", err)
		}
	}
	for _, filePrefix := range levelFileList {
		logFile := logfilename(logFilePath, filePrefix, postFix)
		f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0755)
		if err != nil {
			return fmt.Errorf("创建日志文件失败: %s, error: %w", logFile, err)
		}
		f.Close()
	}
	return nil
}

// logfilename 生成日志文件名
func logfilename(logFilePath string, filePrefix string, postFix string) string {
	return filepath.Join(logFilePath, filePrefix+postFix+".log")
}

// isOneDay 检查两个时间是否在同一天
func isOneDay(oldTime time.Time) bool {
	now := time.Now()
	return now.Year() == oldTime.Year() &&
		now.Month() == oldTime.Month() &&
		now.Day() == oldTime.Day()
}

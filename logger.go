package mysqldb

import (
	"fmt"
	"io"
	"log"
)

const (
	LOG_PREFIX = "[mysqldb]"
	LOG_FORMAT = log.LstdFlags
	LOG_LEVEL  = LOG_DEBUG
)

const (
	LOG_DEBUG int = iota
	LOG_INFO
	LOG_WARN
	LOG_ERROR
)

type LogEntry struct {
	level    int
	LogInfo  *log.Logger
	LogWarn  *log.Logger
	LogDebug *log.Logger
	LogError *log.Logger
	prefix   string
	format   int
	output   io.Writer
}

func InitLogger(o io.Writer) *LogEntry {
	logger := &LogEntry{
		output: o,
		level:  LOG_LEVEL,
		prefix: LOG_PREFIX,
		format: LOG_FORMAT,
	}
	logger.Init()
	return logger
}

func (l *LogEntry) Init() {
	l.LogDebug = log.New(l.output, fmt.Sprintf("%s %5s: ", l.prefix, "DEBUG"), l.format)
	l.LogInfo = log.New(l.output, fmt.Sprintf("%s %5s: ", l.prefix, "INFO"), l.format)
	l.LogWarn = log.New(l.output, fmt.Sprintf("%s %5s: ", l.prefix, "WARN"), l.format)
	l.LogError = log.New(l.output, fmt.Sprintf("%s %5s: ", l.prefix, "ERROR"), l.format)
	return
}

func (l *LogEntry) Prefix() string {
	return l.prefix
}

func (l *LogEntry) SetPrefix(prefix string) {
	l.prefix = prefix
	l.Init()
	return
}

func (l *LogEntry) Flag() int {
	return l.format
}

func (l *LogEntry) SetFlag(flag int) {
	l.format = flag
	l.Init()
	return
}

func (l *LogEntry) Level() int {
	return l.level
}

func (l *LogEntry) SetLevel(level int) {
	l.level = level
	return
}

func (l *LogEntry) Debug(v ...interface{}) {
	if l.level <= LOG_DEBUG {
		l.LogDebug.Output(2, fmt.Sprint(v...))
	}
	return
}

func (l *LogEntry) Debugf(f string, v ...interface{}) {
	if l.level <= LOG_DEBUG {
		l.LogDebug.Output(2, fmt.Sprintf(f, v...))
	}
	return
}

func (l *LogEntry) Info(v ...interface{}) {
	if l.level <= LOG_INFO {
		l.LogInfo.Output(2, fmt.Sprint(v...))
	}
	return
}

func (l *LogEntry) Infof(f string, v ...interface{}) {
	if l.level <= LOG_INFO {
		l.LogInfo.Output(2, fmt.Sprintf(f, v...))
	}
	return
}

func (l *LogEntry) Warn(v ...interface{}) {
	if l.level <= LOG_WARN {
		l.LogWarn.Output(2, fmt.Sprint(v...))
	}
	return
}

func (l *LogEntry) Warnf(f string, v ...interface{}) {
	if l.level <= LOG_WARN {
		l.LogWarn.Output(2, fmt.Sprintf(f, v...))
	}
	return
}

func (l *LogEntry) Error(v ...interface{}) {
	if l.level <= LOG_ERROR {
		l.LogError.Output(2, fmt.Sprint(v...))
	}
	return
}

func (l *LogEntry) Errorf(f string, v ...interface{}) {
	if l.level <= LOG_ERROR {
		l.LogError.Output(2, fmt.Sprintf(f, v...))
	}
	return
}

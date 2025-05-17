package logger

import (
	"os"
	"sync"
)

type Logger struct {
	config    *LoggerConfig
	formatter Formatter
	output    *os.File
	mutex     sync.Mutex
}

type Formatter interface {
	Format(level, msg string, fields map[string]interface{}) (string, error)
}

func NewLogger() *Logger {
	cfg := DefaultConfig()
	var formatter Formatter = &JSONFormatter{}
	if cfg.Format == "text" {
		formatter = &TextFormatter{}
	}

	return &Logger{
		config:    cfg,
		formatter: formatter,
		output:    os.Stdout,
	}
}

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	if !l.config.ShouldLog("info") {
		return
	}
	l.log("INFO", msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	if !l.config.ShouldLog("error") {
		return
	}
	l.log("ERROR", msg, keysAndValues...)
}

func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	if !l.config.ShouldLog("debug") {
		return
	}
	l.log("DEBUG", msg, keysAndValues...)
}

func (l *Logger) log(level, msg string, keysAndValues ...interface{}) {
	fields := make(map[string]interface{})
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		key, ok := keysAndValues[i].(string)
		if ok {
			fields[key] = keysAndValues[i+1]
		}
	}

	formatted, err := l.formatter.Format(level, msg, fields)
	if err != nil {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.output.Write([]byte(formatted))
}
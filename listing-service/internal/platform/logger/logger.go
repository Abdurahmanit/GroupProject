package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Logger struct {
	config    *LoggerConfig
	formatter Formatter
	output    io.Writer
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

	// Открываем файл для записи логов
	logFile, err := os.OpenFile("internal/platform/logger/logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Ошибка при открытии файла лога:", err)
		logFile = os.Stdout // fallback
	}

	// MultiWriter пишет и в файл, и в консоль
	multiOutput := io.MultiWriter(os.Stdout, logFile)

	return &Logger{
		config:    cfg,
		formatter: formatter,
		output:    multiOutput,
	}
}


func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	if !l.config.ShouldLog("info") {
		return
	}
	l.log("INFO", msg, keysAndValues...)
}

func (l *Logger) Warn(msg string, keysAndValues ...interface{}) { // <--- ДОБАВЬ ЭТОТ МЕТОД
	if !l.config.ShouldLog("warn") { // <--- Добавь "warn" в уровни, если нужно
		return
	}
	l.log("WARN", msg, keysAndValues...)
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

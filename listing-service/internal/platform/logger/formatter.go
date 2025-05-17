package logger

import (
	"encoding/json"
	"fmt"
	"time"
)

type LogEntry struct {
	Level      string    `json:"level"`
	Timestamp  time.Time `json:"timestamp"`
	Message    string    `json:"message"`
	Fields     map[string]interface{} `json:"fields"`
}

type JSONFormatter struct{}

func (f *JSONFormatter) Format(level, msg string, fields map[string]interface{}) (string, error) {
	entry := LogEntry{
		Level:     level,
		Timestamp: time.Now(),
		Message:   msg,
		Fields:    fields,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

type TextFormatter struct{}

func (f *TextFormatter) Format(level, msg string, fields map[string]interface{}) (string, error) {
	timestamp := time.Now().Format(time.RFC3339)
	fieldStr := ""
	for k, v := range fields {
		fieldStr += fmt.Sprintf(" %s=%v", k, v)
	}
	return fmt.Sprintf("[%s] %s %s%s\n", timestamp, level, msg, fieldStr), nil
}
package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Init()
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	DPanic(args ...interface{})
	DPanicf(template string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	With(args ...interface{}) Logger
}

type zapLogger struct {
	level      string
	encoding   string
	timeFormat string
	sugar      *zap.SugaredLogger
}

type ZapLoggerConfig struct {
	Level      string
	Encoding   string
	TimeFormat string
}

func NewZapLogger(cfg ZapLoggerConfig) (Logger, error) {
	l := &zapLogger{
		level:      cfg.Level,
		encoding:   cfg.Encoding,
		timeFormat: cfg.TimeFormat,
	}
	l.Init()
	return l, nil
}

func (l *zapLogger) Init() {
	logLevel, err := zapcore.ParseLevel(strings.ToLower(l.level))
	if err != nil {
		logLevel = zapcore.InfoLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	if l.timeFormat != "" {
		encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(l.timeFormat)
	} else {
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	var encoder zapcore.Encoder
	if l.encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stderr),
		logLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	l.sugar = logger.Sugar()
}

func (l *zapLogger) Debug(args ...interface{}) {
	l.sugar.Debug(args...)
}

func (l *zapLogger) Debugf(template string, args ...interface{}) {
	l.sugar.Debugf(template, args...)
}

func (l *zapLogger) Info(args ...interface{}) {
	l.sugar.Info(args...)
}

func (l *zapLogger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

func (l *zapLogger) Warn(args ...interface{}) {
	l.sugar.Warn(args...)
}

func (l *zapLogger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

func (l *zapLogger) Error(args ...interface{}) {
	l.sugar.Error(args...)
}

func (l *zapLogger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

func (l *zapLogger) DPanic(args ...interface{}) {
	l.sugar.DPanic(args...)
}

func (l *zapLogger) DPanicf(template string, args ...interface{}) {
	l.sugar.DPanicf(template, args...)
}

func (l *zapLogger) Fatal(args ...interface{}) {
	l.sugar.Fatal(args...)
}

func (l *zapLogger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

func (l *zapLogger) With(args ...interface{}) Logger {
	newSugar := l.sugar.With(args...)
	return &zapLogger{
		level:      l.level,
		encoding:   l.encoding,
		timeFormat: l.timeFormat,
		sugar:      newSugar,
	}
}

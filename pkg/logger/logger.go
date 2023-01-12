// Package logger it is a simple encapsulation of the go.uber.org/zap package
package logger

import (
	"io"
	"log"
	"os"
	"runtime"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger log writer
var Logger *zap.Logger

// SugarLogger simple logger
var SugarLogger *zap.SugaredLogger

var defaultLogger *zap.Logger
var defaultWriter io.Writer

// InitLogger Initialize logger
func InitLogger(cfg *Config) (err error) {
	encoder := getEncoder()
	syncWriter := getLogWriter(cfg.FileName, cfg.MaxAge, cfg.MaxSize, cfg.MaxBackups, cfg.Comperss)

	level := new(zapcore.Level)
	err = level.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		log.Panic(err)
		return
	}

	core := zapcore.NewCore(encoder, syncWriter, level)
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))
	SugarLogger = Logger.Sugar()
	defaultLogger = Logger.WithOptions(zap.AddCallerSkip(1))

	return
}

func getEncoder() zapcore.Encoder {
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.TimeKey = "time"
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encodeConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encodeConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encodeConfig)
}

func getLogWriter(filename string, maxAge, maxSize, maxBackups int, compress bool) zapcore.WriteSyncer {
	umberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxAge:     maxAge,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		Compress:   compress,
	}
	defaultWriter = umberJackLogger
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(umberJackLogger), zapcore.AddSync(os.Stdout))
	// return zapcore.NewMultiWriteSyncer(zapcore.AddSync(umberJackLogger))
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Debug(msg string, fields ...zap.Field) {
	defaultLogger.Debug(msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(msg string, fields ...zap.Field) {
	defaultLogger.Info(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(msg string, fields ...zap.Field) {
	defaultLogger.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(msg string, fields ...zap.Field) {
	defaultLogger.Error(msg, fields...)
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func With(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}

var Logf *os.File

func RewriteStderrFile(filename string) error {
	var err error
	if runtime.GOOS == "windows" {
		return nil
	}

	Logf, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if err = syscall.Dup2(int(Logf.Fd()), int(os.Stderr.Fd())); err != nil {
		return err
	}

	runtime.SetFinalizer(Logf, func(fd *os.File) {
		fd.Close()
	})

	return nil
}

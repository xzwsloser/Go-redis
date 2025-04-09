package logger

import (
	"github.com/xzwsloser/Go-redis/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	logTmFmt = "2006-01-2 15:04:05"
)

var (
	innerLogger *zap.SugaredLogger
	logger      *zap.Logger
	writeFre    int64
	mu          sync.Mutex
)

func init() {
	core := initCore()
	logger = zap.New(core, zap.AddCaller())
	innerLogger = logger.Sugar()
}

func initCore() zapcore.Core {
	// 1. set the writeSyncer
	var writeSyncOpt []zapcore.WriteSyncer
	if config.GetLogConfig().File == "on" {
		fileName := config.GetLogConfig().Filename
		filePath := "../../" + fileName
		file, _ := os.OpenFile(filePath,
			os.O_CREATE|os.O_RDWR|os.O_APPEND,
			0644)
		writeSyncOpt = append(writeSyncOpt, zapcore.AddSync(file))
	}

	if config.GetLogConfig().Stdout == "on" {
		writeSyncOpt = append(writeSyncOpt, zapcore.AddSync(os.Stdout))
	}

	syncWriter := zapcore.NewMultiWriteSyncer(writeSyncOpt...)

	// 2. set the encoder
	// 2.1 set the time format
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + t.Format(logTmFmt) + "]")
	}

	// 2.2 set the level format out
	customLevelEncoder := func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + level.CapitalString() + "]")
	}

	// 2.3 set the line number
	customCallEncoder := func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + caller.TrimmedPath() + "]")
	}

	// 2.4 set the encoder config
	encoderConf := zapcore.EncoderConfig{
		CallerKey:      "caller_line",
		LevelKey:       "level_name",
		MessageKey:     "msg",
		TimeKey:        "ts",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     customTimeEncoder,
		EncodeLevel:    customLevelEncoder,
		EncodeCaller:   customCallEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	if config.GetLogConfig().Color == "on" {
		encoderConf.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	level, err := zapcore.ParseLevel(strings.ToLower(config.GetLogConfig().Level))
	if err != nil {
		panic(err)
	}

	return zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConf),
		syncWriter,
		zap.NewAtomicLevelAt(level))
}

func Debug(template string, args ...any) {
	innerLogger.Debugf(template, args...)
	atomic.AddInt64(&writeFre, 1)
}

func Info(template string, args ...any) {
	innerLogger.Infof(template, args...)
	atomic.AddInt64(&writeFre, 1)
	checkFlush()
}

func Warn(template string, args ...any) {
	innerLogger.Warnf(template, args...)
	atomic.AddInt64(&writeFre, 1)
	checkFlush()
}

func Error(template string, args ...any) {
	innerLogger.Errorf(template, args...)
	atomic.AddInt64(&writeFre, 1)
	checkFlush()
}

func Fatal(template string, args ...any) {
	defer checkFlush()
	innerLogger.Fatalf(template, args...)
	atomic.AddInt64(&writeFre, 1)
}

func checkFlush() {
	mu.Lock()
	defer mu.Unlock()
	if writeFre >= 5 {
		_ = innerLogger.Sync()
		writeFre = 0
	}
}

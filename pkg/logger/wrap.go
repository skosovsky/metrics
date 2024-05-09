package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

const (
	LevelTrace = slog.Level(-8)
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
	LevelFatal = slog.Level(12)
)

const (
	LevelKey = slog.LevelKey
)

type (
	Logger         = slog.Logger
	Attr           = slog.Attr
	Level          = slog.Level
	Handler        = slog.Handler
	Value          = slog.Value
	HandlerOptions = slog.HandlerOptions
	LogValuer      = slog.LogValuer
	Record         = slog.Record
	Leveler        = slog.Leveler
)

var (
	SetDefault = slog.SetDefault //nolint:gochecknoglobals // alias
	Default    = slog.Default    //nolint:gochecknoglobals // alias

	SetLogLoggerLevel = slog.SetLogLoggerLevel //nolint:gochecknoglobals // alias

	NewTextHandler = slog.NewTextHandler //nolint:gochecknoglobals // alias
	NewJSONHandler = slog.NewJSONHandler //nolint:gochecknoglobals // alias
	New            = slog.New            //nolint:gochecknoglobals // alias

	StringAttr   = slog.String   //nolint:gochecknoglobals // alias
	BoolAttr     = slog.Bool     //nolint:gochecknoglobals // alias
	Float64Attr  = slog.Float64  //nolint:gochecknoglobals // alias
	AnyAttr      = slog.Any      //nolint:gochecknoglobals // alias
	DurationAttr = slog.Duration //nolint:gochecknoglobals // alias
	IntAttr      = slog.Int      //nolint:gochecknoglobals // alias
	Int64Attr    = slog.Int64    //nolint:gochecknoglobals // alias
	Uint64Attr   = slog.Uint64   //nolint:gochecknoglobals // alias

	StringValue = slog.StringValue //nolint:gochecknoglobals // alias
	GroupValue  = slog.GroupValue  //nolint:gochecknoglobals // alias
	Group       = slog.Group       //nolint:gochecknoglobals // alias
)

func Float32Attr(key string, val float32) Attr {
	return slog.Float64(key, float64(val))
}

func UInt32Attr(key string, val uint32) Attr {
	return slog.Int(key, int(val))
}

func Int32Attr(key string, val int32) Attr {
	return slog.Int(key, int(val))
}

func TimeAttr(key string, time time.Time) Attr {
	return slog.String(key, time.String())
}

func ErrAttr(err error) Attr {
	return slog.String("error", err.Error())
}

func Debug(msg string, args ...any) {
	log(context.Background(), slog.Default(), LevelDebug, msg, args...)
}

func Info(msg string, args ...any) {
	log(context.Background(), slog.Default(), LevelInfo, msg, args...)
}

func Warn(msg string, args ...any) {
	log(context.Background(), slog.Default(), LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	log(context.Background(), slog.Default(), LevelError, msg, args...)
}

func Fatal(msg string, args ...any) {
	log(context.Background(), slog.Default(), LevelFatal, msg, args...)
	os.Exit(1)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	log(ctx, Default(), LevelDebug, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	log(ctx, Default(), LevelInfo, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	log(ctx, Default(), LevelWarn, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	log(ctx, Default(), LevelError, msg, args...)
}

func FatalContext(ctx context.Context, msg string, args ...any) {
	log(ctx, Default(), LevelFatal, msg, args...)
	os.Exit(1)
}

func log(ctx context.Context, logger *Logger, level Level, msg string, args ...any) {
	if !logger.Enabled(ctx, level) {
		return
	}
	var pc uintptr //nolint:varnamelen // go internal
	var pcs [1]uintptr

	runtime.Callers(3, pcs[:]) //nolint:mnd // skip [runtime.Callers, this function, this function's caller]
	pc = pcs[0]

	rec := slog.NewRecord(time.Now(), level, msg, pc)
	rec.Add(args...)

	_ = logger.Handler().Handle(ctx, rec) // _ - go internal
}

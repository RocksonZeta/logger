package logger

import (
	"fmt"
	"runtime/debug"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"github.com/rs/zerolog"
)

var PackageField = "package"
var ModuleField = "module"
var FuncName = "func"

//TimeFormat  default timefield format
var TimeFormat = "2006-01-02 15:04:05.999"

type FileLogger struct {
	zerolog.Logger
	rotateLogs *rotatelogs.RotateLogs
}

type Options struct {
	Level string
	//File eg. access_log.%Y%m%d
	File string
	//current log file link
	FileLink     string
	MaxAge       time.Duration
	RotationTime time.Duration
	ForceNewFile bool
}

func rotateLogOptions(options Options) []rotatelogs.Option {
	var args []rotatelogs.Option
	if options.FileLink != "" {
		args = append(args, rotatelogs.WithLinkName(options.FileLink))
	}
	if options.MaxAge <= 0 {
		args = append(args, rotatelogs.WithMaxAge(-1))
	} else {
		args = append(args, rotatelogs.WithMaxAge(options.MaxAge))
	}
	if options.RotationTime > 0 {
		args = append(args, rotatelogs.WithRotationTime(options.RotationTime))
	} else {
		args = append(args, rotatelogs.WithRotationTime(24*time.Hour))
	}
	if options.ForceNewFile {
		args = append(args, rotatelogs.ForceNewFile())
	}
	return args
}

//NewLogger
func NewLogger(options Options) FileLogger {
	zerolog.TimeFieldFormat = TimeFormat
	fileLogger := FileLogger{}
	rotateLogs, err := rotatelogs.New(options.File, rotateLogOptions(options)...)
	if err != nil {
		fmt.Println("failed to create rotatelogs: ", err)
		return fileLogger
	}
	fileLogger.rotateLogs = rotateLogs

	logger := zerolog.New(fileLogger.rotateLogs).With().Timestamp().Logger()
	level, err := zerolog.ParseLevel(options.Level)
	if err != nil {
		fmt.Println(err)
	}
	zerolog.ErrorMarshalFunc = func(err error) interface{} {
		return string(debug.Stack())
	}
	fileLogger.Logger = logger.Level(level)
	return fileLogger
}

type ModuleHook struct {
	pkg string
	mod string
}

func (h ModuleHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if h.pkg != "" {
		e.Str(PackageField, h.pkg)
	}
	if h.mod != "" {
		e.Str(ModuleField, h.mod)
	}
}

//Fork new log for module
func (f FileLogger) Fork(pkg, mod string) FileLogger {
	return FileLogger{Logger: f.Hook(ModuleHook{pkg: pkg, mod: mod})}
}

type Event struct {
	*zerolog.Event
}

func (f FileLogger) Trace() *Event {
	return &Event{f.Logger.Trace()}
}
func (f FileLogger) Debug() *Event {
	return &Event{f.Logger.Debug()}
}
func (f FileLogger) Info() *Event {
	return &Event{f.Logger.Info()}
}
func (f FileLogger) Warn() *Event {
	return &Event{f.Logger.Warn()}
}
func (f FileLogger) Error() *Event {
	return &Event{f.Logger.Error()}
}
func (f FileLogger) Fatal() *Event {
	return &Event{f.Logger.Fatal()}
}
func (f FileLogger) Panic() *Event {
	return &Event{f.Logger.Panic()}
}
func (f FileLogger) WithLevel(level zerolog.Level) *Event {
	return &Event{f.Logger.WithLevel(level)}
}

//Func add func field in log
func (e *Event) Func(funcName string) *Event {
	e.Event.Str(FuncName, funcName)
	return e
}
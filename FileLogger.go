package logger

import (
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"github.com/rs/zerolog"
)

var PackageField = "package"
var ModuleField = "module"
var FuncName = "func"
var LocalIpField = "localIps"

//TimeFormat  default timefield format
var TimeFormat = "2006-01-02 15:04:05.999"

type FileLogger struct {
	zerolog.Logger
	rotateLogs *rotatelogs.RotateLogs
	options    Options
	ips        []string
}

type Options struct {
	Console bool
	Level   string
	//File eg. access_log.%Y%m%d
	File string
	//current log file link
	FileLink     string
	MaxAge       float64 // days
	RotationTime float64 //days
	ForceNewFile bool
	ShowLocalIp  bool
}

func rotateLogOptions(options Options) []rotatelogs.Option {
	var args []rotatelogs.Option
	if options.FileLink != "" {
		args = append(args, rotatelogs.WithLinkName(options.FileLink))
	}
	if options.MaxAge <= 0 {
		args = append(args, rotatelogs.WithMaxAge(-1))
	} else {
		args = append(args, rotatelogs.WithMaxAge(time.Duration(options.MaxAge*float64(24)*float64(time.Hour))))
	}
	if options.RotationTime > 0 {
		args = append(args, rotatelogs.WithRotationTime(time.Duration(options.RotationTime*float64(24)*float64(time.Hour))))
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
	var ips []string
	var err error
	if options.ShowLocalIp {
		ips, err = localIPv4s()
		if err != nil {
			fmt.Println("NewLogger failed on get local IPv4s: ", err)
		}
	}
	fileLogger := FileLogger{options: options, ips: ips}
	rotateLogs, err := rotatelogs.New(options.File, rotateLogOptions(options)...)
	if err != nil {
		fmt.Println("NewLogger failed on creating rotatelogs: ", err)
		return fileLogger
	}
	fileLogger.rotateLogs = rotateLogs

	logger := zerolog.New(fileLogger).With().Timestamp().Logger()
	level, err := zerolog.ParseLevel(options.Level)
	if err != nil {
		fmt.Println("NewLogger failed on parse level:"+options.Level, err)
	}
	zerolog.ErrorMarshalFunc = func(err error) interface{} {
		stack := debug.Stack()
		if options.Console {
			os.Stderr.Write(stack)
		}
		return string(stack)
	}
	fileLogger.Logger = logger.Level(level)
	return fileLogger
}

type ModuleHook struct {
	pkg string
	mod string
	ips []string
}

func (h ModuleHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if h.pkg != "" {
		e.Str(PackageField, h.pkg)
	}
	if h.mod != "" {
		e.Str(ModuleField, h.mod)
	}
	if len(h.ips) > 0 {
		e.Strs(LocalIpField, h.ips)
	}
}

//Fork new log for module
func (f FileLogger) Fork(pkg, mod string) FileLogger {
	return FileLogger{Logger: f.Hook(ModuleHook{pkg: pkg, mod: mod, ips: f.ips})}
}
func (f FileLogger) WithPkg(pkg string) FileLogger {
	return FileLogger{Logger: f.Hook(ModuleHook{pkg: pkg, ips: f.ips})}
}
func (f FileLogger) Write(p []byte) (n int, err error) {
	if f.options.File == "" {
		return 0, nil
	}
	if f.options.Console {
		fmt.Print(string(p))
	}
	return f.rotateLogs.Write(p)
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

// switch ----------------------

func (f FileLogger) TraceEnabled() bool {
	return f.GetLevel() <= zerolog.TraceLevel
}
func (f FileLogger) DebugEnabled() bool {
	return f.GetLevel() <= zerolog.DebugLevel
}
func (f FileLogger) InfoEnabled() bool {
	return f.GetLevel() <= zerolog.InfoLevel
}
func (f FileLogger) WarnEnabled() bool {
	return f.GetLevel() <= zerolog.WarnLevel
}
func (f FileLogger) ErrorEnabled() bool {
	return f.GetLevel() <= zerolog.ErrorLevel
}
func (f FileLogger) FatalEnabled() bool {
	return f.GetLevel() <= zerolog.FatalLevel
}

// method wrap ----------------------
type Event struct {
	*zerolog.Event
}

func (e *Event) Str(key, val string) *Event {
	e.Event.Str(key, val)
	return e
}

func (e *Event) Err(err error) *Event {
	e.Event.Err(err)
	return e
}
func (e *Event) Stack() *Event {
	e.Event.Stack()
	return e
}
func (e *Event) Interface(key string, value interface{}) *Event {
	e.Event.Interface(key, value)
	return e
}

func (e *Event) Int(key string, val int) *Event {
	e.Event.Int(key, val)
	return e
}
func (e *Event) Int32(key string, val int32) *Event {
	e.Event.Int32(key, val)
	return e
}
func (e *Event) Int64(key string, val int64) *Event {
	e.Event.Int64(key, val)
	return e
}

// handy fns ----------------------

//Func add func field in log
func (e *Event) Func(funcName string) *Event {
	e.Event.Str(FuncName, funcName)
	return e
}
func (e *Event) F(funcName string) *Event {
	return e.Func(funcName)
}
func (e *Event) Module(name string) *Event {
	return e.Str(ModuleField, name)
}
func (e *Event) M(name string) *Event {
	return e.Module(name)
}
func (e *Event) Pkg(name string) *Event {
	return e.Str(PackageField, name)
}
func (e *Event) P(name string) *Event {
	return e.Pkg(name)
}

func localIPv4s() ([]string, error) {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}

	return ips, nil
}

package logger_test

import (
	"errors"
	"testing"

	"github.com/RocksonZeta/logger"
)

var options = logger.Options{Console: true, Level: "debug", File: "test_log.%Y%m%d.log", ForceNewFile: true, MaxAge: 1, ShowLocalIp: true}
var serviceLog logger.FileLogger = logger.NewLogger(options)

var log = serviceLog.WithPkg("good/service/users")

func TestLog(t *testing.T) {
	log.Trace().M("User").Func("TestWrite").Interface("user", User{Id: 1}).Send()
	log.Debug().M("User").Func("TestWrite").Str("sql", "select * from User where id=?").Int("id", 1).Send()
	log.Info().M("User").Func("Signin").Str("uid", "1").Msg("hello")
	log.Warn().M("User").Func("Signin").Str("uid", "1").Msg("bad user")
	log.Error().M("User").Stack().Err(errors.New("oh no")).Send()
}

type User struct {
	Id int
}

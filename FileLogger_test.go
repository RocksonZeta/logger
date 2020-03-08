package logger_test

import (
	"errors"
	"testing"
	"time"

	"github.com/RocksonZeta/logger"
)

var serviceLog = logger.NewLogger(logger.Options{Console: true, Level: "debug", File: "test_log.%Y%m%d.log", ForceNewFile: true, MaxAge: 7 * 24 * time.Hour})
var log = serviceLog.Fork("good/service/users", "Users")

func TestLog(t *testing.T) {
	log.Trace().Func("TestWrite").Interface("user", User{Id: 1}).Send()
	log.Debug().Func("TestWrite").Str("sql", "select * from User where id=?").Int("id", 1).Send()
	log.Info().Func("Signin").Str("uid", "1").Msg("hello")
	log.Warn().Func("Signin").Str("uid", "1").Msg("bad user")
	log.Error().Stack().Err(errors.New("oh no")).Send()
}

type User struct {
	Id int
}

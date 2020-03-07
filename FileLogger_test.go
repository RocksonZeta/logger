package logger_test

import (
	"errors"
	"testing"

	"github.com/RocksonZeta/summer/logger"
)

type User struct {
	Id int
}

func TestWrite(t *testing.T) {
	logger := logger.NewLogger(logger.Options{Level: "debug", File: "test_log.%Y%m%d.log", ForceNewFile: true})
	log := logger.Fork("logger_test", "")
	log.Debug().Func("TestWrite").Interface("obj", User{Id: 1}).Msg("hello1")
	log.Info().Func("hello").Str("uid", "1").Msg("hello")
	log.Error().Stack().Err(errors.New("hello1")).Send()
}

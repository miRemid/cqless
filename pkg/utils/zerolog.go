package utils

import (
	"os"
	"time"

	"github.com/miRemid/cqless/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

/*
TODO: Zerolog的初始化工作，包括但不限于
- 日志文件输出
- 滚动日志输出
- 格式优化
*/
func init() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RubyDate}).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Caller().
		Int("pid", os.Getpid()).
		Logger()
	log.Logger = logger
}

func InitLogger(config *types.Logger) {
	var logger zerolog.Logger
	if config.Debug {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RubyDate}).
			Level(zerolog.DebugLevel).
			With().
			Timestamp().
			Caller().
			Int("pid", os.Getpid()).
			Logger()
	} else {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RubyDate}).
			Level(zerolog.InfoLevel).
			With().
			Timestamp().
			Logger()
	}
	log.Logger = logger
}

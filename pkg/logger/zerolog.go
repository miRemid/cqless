package logger

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/miRemid/cqless/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

/*
TODO: Zerolog的初始化工作，包括但不限于
- 日志文件输出
- 滚动日志输出
- 格式优化
*/
// func init() {
// 	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RubyDate}).
// 		Level(zerolog.TraceLevel).
// 		With().
// 		Timestamp().
// 		Caller().
// 		Int("pid", os.Getpid()).
// 		Logger()
// }

func InitLogger(config *types.LoggerConfig) {
	if types.DEBUG == "TRUE" {
		log.Debug().Msg("running in debug mode")
		log.Logger = zerolog.New(
			zerolog.ConsoleWriter{
				Out: os.Stdout, TimeFormat: time.RubyDate}).
			Level(zerolog.DebugLevel).
			With().
			Timestamp().
			Caller().
			Int("pid", os.Getpid()).
			Logger()
	} else {
		var writers []io.Writer
		writers = append(writers, zerolog.ConsoleWriter{
			Out: os.Stdout, TimeFormat: time.RubyDate})
		if config.EnableSaveFile {
			writers = append(writers, &lumberjack.Logger{
				Filename:   path.Join(config.SavePath, "cqless.log"),
				MaxBackups: config.MaxBackups,
				MaxSize:    config.MaxSize,
				MaxAge:     config.MaxAge,
			})
		}
		mw := io.MultiWriter(writers...)
		log.Logger = zerolog.New(mw).
			Level(zerolog.InfoLevel).
			With().
			Timestamp().
			Logger()
	}
}

type moduleHook struct {
	module string
}

func (hook moduleHook) Run(e *zerolog.Event, level zerolog.Level, message string) {
	e.Str("module", hook.module)
}

func ModuleHook(module string) zerolog.Hook {
	return moduleHook{
		module: module,
	}
}

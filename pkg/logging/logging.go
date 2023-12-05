package logging

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shurubtsov/lamoda-test-task/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type logger struct {
	zerolog.Logger
}

var (
	instance *logger
	once     sync.Once
)

// GetLogger функция для получения инстанса логгера.
func GetLogger() *logger {
	return instance
}

func init() {

	cfg := config.GetConfig()

	once.Do(func() {

		// управление проверкой вывода логов
		checkLogOutput := map[string]io.Writer{
			"develop": os.Stderr,
			"prod":    io.Discard,
		}

		/*
			Уровни логирования делятся на константы библиотеки от -1 до 5
			Чем ниже уровень логгирования тем больше логов будет показано.

			Таким образом при выборе уровня Trace (-1) будут показаны все остальные логи.
			Самый высокий уровень у Fatal (5), при его выборе остальные будут отсеяны.
		*/
		checkLogLevel := map[int]zerolog.Level{
			-1: zerolog.TraceLevel,
			0:  zerolog.DebugLevel,
			1:  zerolog.InfoLevel,
			2:  zerolog.WarnLevel,
			3:  zerolog.ErrorLevel,
			4:  zerolog.FatalLevel,
		}

		// проверка места для вывода логов
		out, ok := checkLogOutput[cfg.Logging.Output]
		if !ok {
			out = os.Stdout
		}

		// проверка уровня логирования
		level, ok := checkLogLevel[cfg.Logging.Level]
		if !ok {
			level = zerolog.NoLevel
		}
		zerolog.SetGlobalLevel(level)

		// стартовая настройка логгера
		output := zerolog.ConsoleWriter{
			Out:        out,
			NoColor:    true,
			TimeFormat: time.UnixDate,
		}

		/*
			Следующие функции образуют
			кастомные настройки формата логгирования.

			Формат вывода уровня логгирования,
			Формат сообщения,
			Формат отображения функции где вызван лог
		*/
		output.FormatLevel = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("(%-3s) |", i))
		}
		output.FormatMessage = func(i interface{}) string {
			return fmt.Sprintf("[%s]", i)
		}

		// добавление трейсера для ошибок с вызовом ".Stack()"
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
			return file + " | " + runtime.FuncForPC(pc).Name() + "():" + strconv.Itoa(line)
		}

		logg := zerolog.New(output).With().Timestamp().Caller().Logger()
		logg.Info().Msg("Getting logger")

		instance = &logger{Logger: logg}
	})
}

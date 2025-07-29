package logger

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

type writerHook struct {
	Writer    []io.Writer
	LogLevels []logrus.Level
}

func (hook *writerHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	for _, w := range hook.Writer {
		w.Write([]byte(line))
	}
	return err
}

var e *logrus.Entry

type Logger struct {
	*logrus.Entry
}

func GetLogger() Logger {
	return Logger{e}
}

func (l *Logger) GetLoggerWithField(k string, v any) Logger {
	return Logger{l.WithField(k, v)}
}

func (hook *writerHook) Levels() []logrus.Level {
	return hook.LogLevels
}

func Init() {
	l := logrus.New()
	l.SetReportCaller(true)
	l.Formatter = &logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
		ForceColors:   true,
	}

	err := os.MkdirAll("logs", 0777)
	if err != nil {
		panic(err)
	}

	logFile, err := os.OpenFile("logs/logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		panic(err)
	}

	l.SetOutput(io.Discard)

	l.AddHook(&writerHook{
		Writer:    []io.Writer{logFile, os.Stdout},
		LogLevels: logrus.AllLevels,
	})

	l.SetLevel(logrus.TraceLevel)

	e = logrus.NewEntry(l)
}

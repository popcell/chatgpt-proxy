package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"runtime"
)

var (
	log *logrus.Logger
)

func init() {
	log = logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

func getFileFromPath(path string) string {
	return filepath.Base(filepath.Dir(path)) + "/" + filepath.Base(path)
}
func Ansi(colorString string) func(...interface{}) string {
	return func(args ...interface{}) string {
		return fmt.Sprintf(colorString, fmt.Sprint(args...))
	}
}

var (
	Green  = Ansi("\033[1;32m%s\033[0m")
	Yellow = Ansi("\033[1;33m%s\033[0m")
	Red    = Ansi("\033[1;31m%s\033[0m")
)

func Info(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	// 记录日志消息
	log.WithFields(logrus.Fields{
		"file": fmt.Sprintf("%s:%d", getFileFromPath(file), line),
	}).Info(Green(args...))
}

func Warn(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	// 记录日志消息
	log.WithFields(logrus.Fields{
		"file": fmt.Sprintf("%s:%d", getFileFromPath(file), line),
	}).Warn(Yellow(args...))
}

func Error(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	// 记录日志消息
	log.WithFields(logrus.Fields{
		"file": fmt.Sprintf("%s:%d", getFileFromPath(file), line),
	}).Error(Red(args...))
}
func Panic(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	// 记录日志消息
	log.WithFields(logrus.Fields{
		"file": fmt.Sprintf("%s:%d", getFileFromPath(file), line),
	}).Panic(Red(args...))
}

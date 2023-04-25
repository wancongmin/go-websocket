package mylog

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// 继承 Logger
type MyLogger struct {
	*log.Logger
	mu sync.Mutex
}

// 日志路径
var path = "logs/"

func NewLogger(prefix string) MyLogger {
	file := time.Now().Format("20060102") + ".log"
	logFile, err := os.OpenFile(path+file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open mylog file failed, err:", err)
		panic(err)
	}
	if prefix != "" {
		prefix = "[" + prefix + "]"
	}
	logger := log.New(logFile, prefix, log.LstdFlags|log.Llongfile)
	return MyLogger{Logger: logger}
}

func (l *MyLogger) Output2(calldepth int, s string) error {
	return l.Logger.Output(calldepth, s)
}

func Output(s, prefix string) error {
	var std = NewLogger(prefix)
	return std.Logger.Output(2, s)
}

func Info(s string) {
	var std = NewLogger("info")
	err := std.Logger.Output(2, s)
	if err != nil {
		return
	}
}

func Error(s string) {
	var std = NewLogger("error")
	err := std.Logger.Output(2, s)
	if err != nil {
		return
	}
}

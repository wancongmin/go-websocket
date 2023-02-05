package mylog

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

//继承 Logger
type MyLogger struct {
	*log.Logger
	mu sync.Mutex
}

//日志路径
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
	l.Logger.Output(calldepth, s)
	return nil
}

func Output(s, prefix string) error {
	var std = NewLogger(prefix)
	return std.Logger.Output(2, s)
}

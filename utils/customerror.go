package utils

import (
	"fmt"
	"runtime/debug"
	"websocket/lib/mylog"
)

func CustomError() {
	err := recover()
	if err != nil {
		s := string(debug.Stack())
		_ = mylog.Output(fmt.Sprintf("err=%v, stack=%s\n", err, s), "error")
		err := mylog.Output("Recover:"+fmt.Sprintf("%v", err), "error")
		if err != nil {
			fmt.Println("日志写入错误:" + err.Error())
			return
		}
	}
}

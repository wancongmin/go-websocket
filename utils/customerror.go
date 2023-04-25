package utils

import (
	"fmt"
	"websocket/lib/mylog"
)

func CustomError() {
	err := recover()
	if err != nil {
		fmt.Println(err)
		err := mylog.Output("Recover:"+fmt.Sprintf("%v", err), "error")
		if err != nil {
			fmt.Println("日志写入错误:" + err.Error())
			return
		}
	}
}

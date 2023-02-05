package utils

import (
	"fmt"
	"websocket/lib/mylog"
)

func CustomError() {
	err := recover()
	if err != nil {
		fmt.Println(err)
		err := mylog.Output(fmt.Sprintf("%v", err), "error")
		if err != nil {
			return
		}
	}
}

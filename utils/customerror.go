package utils

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"websocket/config"
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

func CdnUrl(url string) string {
	pattern := `^(https?://)?([a-zA-Z0-9]+(-[a-zA-Z0-9]+)*\.)+[a-zA-Z]{2,}(/.*)?$`
	// 编译正则表达式
	regex, err := regexp.Compile(pattern)
	if err != nil {
		mylog.Error("regexp compile error:" + err.Error())
		return url
	}
	// 匹配路径
	if regex.MatchString(url) {
		return url
	} else {
		var conf = &config.Conf{}
		var domain string
		err := config.ConfFile.Section("conf").MapTo(conf)
		if err != nil {
			domain = ""
		} else {
			domain = conf.OssUrl
		}
		return domain + url
	}
}

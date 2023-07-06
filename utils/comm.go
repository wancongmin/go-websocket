package utils

import (
	"math/rand"
	"regexp"
	"strings"
	"time"
	"websocket/config"
	"websocket/lib/mylog"
)

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

// 缩略图
func Thumb(url, w string) string {
	url = CdnUrl(url)
	url = url + "?x-oss-process=image/resize,w_" + w + ",m_lfit"
	return url
}

// 圆角缩略图
func RoundThumb(url, w, r string) string {
	url = CdnUrl(url)
	if r != "0" {
		url = url + "?x-oss-process=image/resize,w_" + w + ",m_lfit/rounded-corners,r_" + r
	} else {
		url = url + "?x-oss-process=image/resize,w_" + w + ",m_lfit"
	}
	return url
}

var CHARS = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
	"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}

/*
RandNumString  生成随机数字字符串([0~9])

	lenNum 长度
*/
func RandNumString(lenNum int) string {
	str := strings.Builder{}
	length := 10
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < lenNum; i++ {
		str.WriteString(CHARS[52+rand.Intn(length)])
	}
	return str.String()
}

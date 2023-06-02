package utils

import (
	"regexp"
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
	url = url + "?x-oss-process=image/resize,w_" + w + ",m_lfit/rounded-corners,r_" + r
	return url
}

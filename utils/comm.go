package utils

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"websocket/config"
	"websocket/lib/db"
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

// Thumb 缩略图
func Thumb(url, w string) string {
	url = CdnUrl(url)
	url = url + "?x-oss-process=image/resize,w_" + w + ",m_lfit"
	return url
}

// RoundThumb 圆角缩略图
func RoundThumb(url, r string) string {
	url = CdnUrl(url)
	return url + "/rounded-corners,r_" + r + "/format,png"
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

// EarthDistance 计算经纬度距离
func EarthDistance(latitude1, longitude1, latitude2, longitude2 string) (float64, error) {
	lat1, err := strconv.ParseFloat(latitude1, 64)
	if err != nil {
		return 0, err
	}
	lng1, err := strconv.ParseFloat(longitude1, 64)
	if err != nil {
		return 0, err
	}
	lat2, err := strconv.ParseFloat(latitude2, 64)
	if err != nil {
		return 0, err
	}
	lng2, err := strconv.ParseFloat(longitude2, 64)
	if err != nil {
		return 0, err
	}
	if lat1 == 0 || lng1 == 0 || lat2 == 0 || lng2 == 0 {
		return 0, errors.New("参数错误")
	}
	radius := 6378.137
	rad := math.Pi / 180.0
	lat1 = lat1 * rad
	lng1 = lng1 * rad
	lat2 = lat2 * rad
	lng2 = lng2 * rad
	theta := lng2 - lng1
	dist := math.Acos(math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta))
	return Decimal(dist * radius * 1000), nil
}

func Decimal(num float64) float64 {
	num, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", num), 64)
	return num
}

type ConfigData struct {
	Id    int
	Name  string
	Value string
}

func GetConfVal(name string) string {
	var conf ConfigData
	db.Db.Table("fa_config").
		Where("name = ?", name).
		First(&conf)
	return conf.Value
}

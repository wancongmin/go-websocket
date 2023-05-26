package token

import (
	"crypto/hmac"
	"errors"
	"fmt"
	"golang.org/x/crypto/ripemd160"
	"websocket/config"
	"websocket/lib/db"
)

type Token struct {
	UserId     uint32
	Token      string
	Expiretime int
	Createtime int
}

/**
 * 获取加密后的Token
 */
func getEncryptedToken(token string) (string, error) {
	var conf = &config.Token{}
	err := config.ConfFile.Section("token").MapTo(conf)
	if err != nil {
		return "", err
	}
	hash := hmac.New(ripemd160.New, []byte(conf.Key))
	hash.Write([]byte(token))
	hashString := fmt.Sprintf("%x", hash.Sum(nil))
	return hashString, nil
}

func Get(token string) (Token, error) {
	var tokenObj Token
	encryptedToken, err := getEncryptedToken(token)
	if err != nil {
		return tokenObj, err
	}
	db.Db.Table("fa_user_token").Where("token = ?", encryptedToken).First(&tokenObj)
	if tokenObj.UserId == 0 {
		return tokenObj, errors.New("token error")
	}
	return tokenObj, nil
}

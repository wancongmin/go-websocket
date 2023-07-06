package comm

import (
	"websocket/lib/db"
)

type UserFriend struct {
	Id         int `gorm:"id"`
	UserId     int `gorm:"user_id"`
	FriendId   int
	Type       int `gorm:"type"`
	UserShow   int
	FriendShow int
	nickname   string
	avatar     string
	gender     int
	FNickname  string
	FAvatar    string
	FGender    int
}

func GetFriendLocation(userId uint32) []User {
	var friendList []UserFriend
	db.Db.Table("fa_user_friend").
		Select("id,user_id,friend_id,type").
		Where(UserFriend{Type: 2}).
		Where(
			db.Db.Where("user_id = ?", userId).Or("friend_id = ?", userId),
		).
		Find(&friendList)
	var users []User
	for _, friend := range friendList {
		var uid uint32
		if friend.UserId == int(userId) {
			uid = uint32(friend.FriendId)
		} else {
			uid = uint32(friend.UserId)
		}
		user := GetUserLocation(uid)
		if user.Id == 0 {
			continue
		}
		users = append(users, user)
	}
	return users
}

package model

import "websocket/lib/db"

type User struct {
	Id       uint32
	Nickname string
	Mobile   string
	Avatar   string
}

type UserFriend struct {
	Id         int
	UserId     int
	FriendId   int
	Type       int
	UserShow   int
	FriendShow int
	nickname   string
	avatar     string
	gender     int
	FNickname  string
	FAvatar    string
	FGender    int
}

type ClubActivityOrder struct {
	Id           int
	UserId       int
	Status       string
	Type         int
	PositionShow int
}

type ClubJoin struct {
	Id           int
	UserId       int
	ClubId       int
	Status       int
	PositionShow int
}

func GetFriendLocation(userId int) []UserFriend {
	var friendList []UserFriend
	db.Db.Table("fa_user_friend f").
		Select("f.id,f.user_id,f.user_id,f.type,uu.nickname,uu.avatar,uu.gender,fu.nickname" +
			" f_nickname,fu.avatar f_avatar,fu.gender f_gender").
		Where(UserFriend{Type: 2}).
		Where(
			db.Db.Where("f.user_id = ?", userId).Or("f.friend_id = ?", userId),
		).
		Joins("join fa_user uu on uu.id = f.user_id").
		Joins("join fa_user fu on fu.id = f.friend_id").
		Find(&friendList)

	//for _, item := range friendList {
	//	if userId == item.UserId {
	//
	//	}
	//}
	return friendList
}

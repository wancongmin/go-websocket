package club

import (
	"websocket/lib/db"
	"websocket/model/comm"
)

type ClubActivityOrder struct {
	Id           int
	ActivityId   int
	UserId       int
	Status       string
	Type         int
	PositionShow int
}

func GetActivityMemberLocation(activityId int, userId uint32) []comm.User {
	var users []comm.User
	if activityId == 0 {
		return users
	}
	var members []ClubActivityOrder
	db.Db.Table("fa_club_activity_order").
		Select("id,user_id").
		Where(ClubActivityOrder{ActivityId: activityId, PositionShow: 1}).
		Where("status IN ?", []string{"1", "2", "3"}).
		Where("user_id <> ?", userId).
		Group("user_id").
		Find(&members)
	for _, member := range members {
		user := comm.GetUserLocation(uint32(member.UserId))
		if user.Id == 0 {
			continue
		}
		users = append(users, user)
	}
	return users
}

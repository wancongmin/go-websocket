package club

import (
	"websocket/lib/db"
	"websocket/model/comm"
)

type ClubJoin struct {
	Id           int
	UserId       int
	ClubId       int
	Status       int
	PositionShow int
}

func GetClubMemberLocation(clubId int, userId uint32) []comm.User {
	var users []comm.User
	if clubId == 0 {
		return users
	}
	var members []ClubJoin
	db.Db.Table("fa_club_join").
		Select("id,user_id").
		Where(ClubJoin{ClubId: clubId, Status: 2, PositionShow: 1}).
		Where("user_id <> ?", userId).
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

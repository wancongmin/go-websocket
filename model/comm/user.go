package comm

type User struct {
	Id          uint32
	Nickname    string
	Mobile      string
	Avatar      string
	Gender      int
	Longitude   string
	Latitude    string
	Electricity string
	GhostType   int `gorm:"ghost_type"`
	GhostTime   int `gorm:"ghost_time"`
	ChooseType  int `gorm:"-"`
}

package model

import (
	"time"
)

// 房间结构
type GameRoom struct {
	Id              int
	UserId          int
	MasterId        int
	HidingSecond    int
	ArrestSecond    int
	Status          int
	StartHidingTime int
	StartArrestTime int
	EndTime         int
	Coordinate      string
	ChatRoom        string
}

type GameRoomChecker struct {
	interval time.Duration //  Heartbeat detection interval(检测时间间隔)
	quitChan chan bool     // Quit signal(退出信号)
	GameRoom GameRoom
}

func NewGameRoomChecker(interval time.Duration) *GameRoomChecker {
	gameRoom := &GameRoomChecker{
		interval: interval,
		quitChan: make(chan bool),
	}

	return gameRoom
}

// 游戏检测器
func (r *GameRoomChecker) start() {
	ticker := time.NewTicker(r.interval)
	for {
		select {
		case <-ticker.C:
			r.check()
		case <-r.quitChan:
			ticker.Stop()
			return
		}
	}
}

// 检测
func (h *GameRoomChecker) check() (err error) {
	return err
}

func (h *GameRoomChecker) Start() {
	go h.start()
}

func (h *GameRoomChecker) Stop() {
	go func() {
		h.quitChan <- true
	}()
}

package internal

import (
	"common"
	"conf"
	"game/poker"
	"msg"
	"time"

	"github.com/name5566/leaf/log"
)

func (user *User) checkRoomMinChips(minChips int64, create bool) bool {
	if minChips > user.baseData.userData.Chips {
		//机器人需要对其金币进行修改
		if user.isRobot() {
			user.baseData.userData.Chips += minChips
			return true
		}
		if create {
			user.WriteMsg(&msg.S2C_CreateRoom{
				Error:    msg.S2C_CreateRoom_LackOfChips,
				MinChips: minChips,
			})
		} else {
			user.WriteMsg(&msg.S2C_EnterRoom{
				Error:    msg.S2C_EnterRoom_LackOfChips,
				MinChips: minChips,
			})
		}
		return false
	}
	return true
}

func (user *User) checkRoomMaxChips(maxChips int64, create bool) bool {
	if user.baseData.userData.Chips > maxChips {
		if user.isRobot() {
			user.baseData.userData.Chips = maxChips
			return true
		}
		if create {
			user.WriteMsg(&msg.S2C_CreateRoom{
				Error:    msg.S2C_CreateRoom_MaxChipsLimit,
				MaxChips: maxChips,
			})
		} else {
			user.WriteMsg(&msg.S2C_EnterRoom{
				Error:    msg.S2C_EnterRoom_MaxChipsLimit,
				MaxChips: maxChips,
			})
		}
		return false
	}
	return true
}

func (user *User) createLandlordRoom(rule *poker.LandlordRule) {
	roomNumber := getRoomNumber()
	if _, ok := roomNumberRooms[roomNumber]; ok {
		user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_InnerError})
		return
	}
	room := newLandlordRoom(rule)
	log.Debug("userID: %v 创建斗地主房间: %v, 类型: %v, 底分: %v, 红包: %v", user.baseData.userData.UserID, roomNumber, rule.RoomType, rule.BaseScore, rule.RedPacketType)
	room.number = roomNumber
	roomNumberRooms[roomNumber] = room
	room.ownerUserID = user.baseData.userData.UserID
	room.creatorUserID = user.baseData.userData.UserID
	user.enterRoom(room)
}

func (user *User) createOrEnterPracticeRoom() {
	for _, r := range roomNumberRooms {
		room := r.(*LandlordRoom)
		if room.rule.RoomType == roomPractice && !room.full() {
			user.enterRoom(r)
			return
		}
	}
	rule := &poker.LandlordRule{
		RoomType:   roomPractice,
		MaxPlayers: 3,
	}
	user.createLandlordRoom(rule)
}

func (user *User) createOrEnterBaseScoreMatchingRoom(baseScore int) {
	/*
		minChips := int64(baseScore) * 10
		switch baseScore {
		case 500:
			minChips = 1000
		}
		if !user.checkRoomMinChips(minChips, false) {
			return
		}
		var tickets int64
		switch baseScore {
		case 500:
			if !user.checkRoomMaxChips(60000, true) {
				return
			}
			tickets = 240
		case 3000:
			tickets = 1800
		case 10000:
			tickets = 3600
		default:
			user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_RuleError})
			return
		}
	*/
	data := getMatch(baseScore)
	/*
	   检查入场最低金币数量
	*/
	if !user.checkRoomMinChips(int64(data.BaseScore), false) {
		return
	}
	/*
	   检查最高入场金币数量
	*/
	if !user.checkRoomMaxChips(int64(data.MaxScore), false) {
		return
	}
	//优先进入真实玩家场
	if !user.isRobot() {
		if user.enterBaseScoreMatchingRoom(baseScore, 2, true) {
			return
		}
		if user.enterBaseScoreMatchingRoom(baseScore, 1, true) {
			return
		}
	}
	//后续加入机器人场
	if user.isRobot() {
		if user.enterBaseScoreMatchingRoom(baseScore, 2, false) {
			return
		}
		if user.enterBaseScoreMatchingRoom(baseScore, 1, false) {
			return
		}
	}
	tickets := data.Tickets
	minChips := data.MinScore
	if user.isRobot() { // 只进入房间不创建房间
		user.WriteMsg(&msg.S2C_EnterRoom{Error: msg.S2C_EnterRoom_Unknown})
		return
	}
	rule := &poker.LandlordRule{
		RoomType:   roomBaseScoreMatching,
		MaxPlayers: 3,
		BaseScore:  baseScore,
		MinChips:   int64(minChips),
		Tickets:    int64(tickets),
	}
	user.createLandlordRoom(rule)
}

func (user *User) createBasePrivateRoom(baseScore int) {
	minChips := int64(baseScore) * 10
	if !user.checkRoomMinChips(minChips, true) {
		return
	}
	var tickets int64
	switch baseScore {
	case 30000:
		tickets = 4500
	case 50000:
		tickets = 6000
	case 100000:
		tickets = 8500
	default:
		user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_RuleError})
		return
	}
	rule := &poker.LandlordRule{
		RoomType:   roomBaseScorePrivate,
		MaxPlayers: 3,
		BaseScore:  baseScore,
		Tickets:    tickets,
		MinChips:   minChips,
	}
	user.createLandlordRoom(rule)
}

func (user *User) createVIPPrivateRoom(maxPlayers int) {
	if common.Index([]int{2, 3}, maxPlayers) == -1 {
		user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_RuleError})
		return
	}
	var minChips int64 = 10000
	if !user.checkRoomMinChips(minChips, true) {
		return
	}
	rule := &poker.LandlordRule{
		RoomType:   roomVIPPrivate,
		MaxPlayers: maxPlayers,
		MinChips:   minChips,
	}
	user.createLandlordRoom(rule)
}

func (user *User) createOrEnterRedPacketMatchingRoom(redPacketType int) {
	if common.Index([]int{1, 10}, redPacketType) == -1 {
		user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_RuleError})
		return
	}
	if redPacketType == 1 {
		if time.Now().Hour() < conf.GetOneRedpacketInfo().Start || time.Now().Hour() > conf.GetOneRedpacketInfo().End {
			user.WriteMsg(&msg.S2C_EnterRoom{Error: msg.S2C_EnterRoom_NotRightNow})
			return
		}
	}
	if redPacketType == 10 {
		if time.Now().Hour() < conf.GetTenRedpacketInfo().Start || time.Now().Hour() > conf.GetTenRedpacketInfo().End {
			user.WriteMsg(&msg.S2C_EnterRoom{Error: msg.S2C_EnterRoom_NotRightNow})
			return
		}
	}

	var minChips int64
	switch redPacketType {
	case 10:
		minChips = conf.GetTenRedpacketInfo().Chips
	default:
		minChips = conf.GetOneRedpacketInfo().Chips
	}

	if !user.checkRoomMinChips(minChips, false) {
		return
	}
	//真实玩家优先跟真实玩家一起打红包赛
	if !user.isRobot() {
		if user.enterRedPacketMatchingRoom(redPacketType, 2, true) {
			return
		}
		if user.enterRedPacketMatchingRoom(redPacketType, 1, true) {
			return
		}
	}
	if user.isRobot() {
		if user.enterRedPacketMatchingRoom(redPacketType, 2, false) {
			return
		}
		if user.enterRedPacketMatchingRoom(redPacketType, 1, false) {
			return
		}
	}
	if user.isRobot() { // 只进入房间不创建房间
		user.WriteMsg(&msg.S2C_EnterRoom{Error: msg.S2C_EnterRoom_Unknown})
		return
	}
	rule := &poker.LandlordRule{
		RoomType:      roomRedPacketMatching,
		MaxPlayers:    3,
		RedPacketType: redPacketType,
		MinChips:      minChips,
		BaseScore:     int(minChips),
	}
	user.createLandlordRoom(rule)
}

func (user *User) createRedPacketPrivateRoom(redPacketType int) {
	if time.Now().Hour() < conf.GetHundredRedpacketInfo().Start || time.Now().Hour() > conf.GetHundredRedpacketInfo().End {
		user.WriteMsg(&msg.S2C_EnterRoom{Error: msg.S2C_EnterRoom_NotRightNow})
		return
	}
	var minChips int64
	switch redPacketType {
	case 100:
		minChips = conf.GetHundredRedpacketInfo().Chips
	case 999:
		minChips = 498 * 10000
	default:
		user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_RuleError})
		return
	}
	if !user.checkRoomMinChips(minChips, true) {
		return
	}
	rule := &poker.LandlordRule{
		RoomType:      roomRedPacketPrivate,
		MaxPlayers:    3,
		RedPacketType: redPacketType,
		MinChips:      minChips,
	}
	user.createLandlordRoom(rule)
}

/*

//如果房间里面有一个真实玩家,可以先加入一个机器人(房间最多只能有一个机器人)
		lable := false
		if room.RealPlayer() == 1 && user.isRobot() {
			lable = true
		}

		//如果房间里面有两个
		if room.RealPlayer() != 2 && user.isRobot() && !room.RealPlayer() == 1 {

		}

*/
func (user *User) enterBaseScoreMatchingRoom(baseScore int, playerNumber int, real bool) bool {
	for _, r := range roomNumberRooms {
		room := r.(*LandlordRoom)
		if real {
			if !conf.Server.Model {
				if room.rule.RoomType == roomBaseScoreMatching && room.rule.BaseScore == baseScore && room.RealPlayer() == playerNumber && !room.full() {
					user.enterRoom(r)
					return true
				}
			}
			if !room.loginIPs[user.baseData.userData.LoginIP] && room.rule.RoomType == roomBaseScoreMatching && room.rule.BaseScore == baseScore && len(room.positionUserIDs) == playerNumber {

				if !room.playTogether(user) {
					user.enterRoom(r)
					return true
				}
			}
		} else {
			if !conf.Server.Model {
				if room.RealPlayer() == 1 && len(room.positionUserIDs) == 1 && user.isRobot() || !user.isRobot() || room.RealPlayer() == 2 {
					if room.rule.RoomType == roomBaseScoreMatching && room.rule.BaseScore == baseScore && len(room.positionUserIDs) == playerNumber && !room.full() {
						user.enterRoom(r)
						return true
					}
				}
			}
			if !room.loginIPs[user.baseData.userData.LoginIP] && room.RealPlayer() == 1 && len(room.positionUserIDs) == 1 && user.isRobot() || !user.isRobot() || room.RealPlayer() == 2 {
				if room.rule.RoomType == roomBaseScoreMatching && room.rule.BaseScore == baseScore && len(room.positionUserIDs) == playerNumber && !room.full() {
					if !room.playTogether(user) {
						user.enterRoom(r)
						return true
					}
				}
			}
		}
	}
	return false
}

func (user *User) enterRedPacketMatchingRoom(redPacketType int, playerNumber int, real bool) bool {
	for _, r := range roomNumberRooms {
		room := r.(*LandlordRoom)
		//!room.loginIPs[user.baseData.userData.LoginIP] &&
		if real {
			//测试环境
			if !conf.Server.Model {
				if room.rule.RoomType == roomRedPacketMatching && room.rule.RedPacketType == redPacketType && room.RealPlayer() == playerNumber && !room.full() {
					user.enterRoom(r)
					return true
				}
			}
			if !room.loginIPs[user.baseData.userData.LoginIP] && room.rule.RoomType == roomRedPacketMatching && room.rule.RedPacketType == redPacketType && room.RealPlayer() == playerNumber && !room.full() {
				if !room.playTogether(user) {
					user.enterRoom(r)
				}
				return true
			}
		} else {
			if !conf.Server.Model {
				if room.RealPlayer() == 1 && len(room.positionUserIDs) == 1 && user.isRobot() || !user.isRobot() || room.RealPlayer() == 2 {
					if room.rule.RoomType == roomRedPacketMatching && room.rule.RedPacketType == redPacketType && len(room.positionUserIDs) == playerNumber && !room.full() {
						user.enterRoom(r)
						return true
					}
				}
			}
			if room.RealPlayer() == 1 && len(room.positionUserIDs) == 1 && user.isRobot() || !user.isRobot() || room.RealPlayer() == 2 {
				if !room.loginIPs[user.baseData.userData.LoginIP] && room.rule.RoomType == roomRedPacketMatching && room.rule.RedPacketType == redPacketType && len(room.positionUserIDs) == playerNumber && !room.full() {
					if !room.playTogether(user) {
						user.enterRoom(r)
						return true
					}
				}
			}
		}
	}
	return false
}

func (user *User) enterRoom(r interface{}) {
	room := r.(*LandlordRoom)
	sitDown := room.Enter(user)
	if sitDown {
		userIDRooms[user.baseData.userData.UserID] = r
		user.baseData.ownerUserID = room.ownerUserID
	}
}

func (user *User) exitRoom(r interface{}, forcible bool) {
	room := r.(*LandlordRoom)
	if room.state == roomGame {
		if forcible {
			user.WriteMsg(&msg.S2C_ExitRoom{Error: msg.S2C_ExitRoom_GamePlaying})
		}
	} else {
		room.Exit(user)
	}
}

func (user *User) setVIPRoomChips(r interface{}, chips int64) {
	room := r.(*LandlordRoom)
	if room.rule.RoomType != roomVIPPrivate {
		return
	}
	if playerData, ok := room.userIDPlayerDatas[user.baseData.userData.UserID]; ok {
		playerData.vipChips = chips
		broadcast(&msg.S2C_SetVIPRoomChips{
			Error:    msg.S2C_SetVIPChips_OK,
			Position: playerData.position,
			Chips:    playerData.vipChips,
		}, room.positionUserIDs, -1)
	}
}
func getMatch(baseScore int) conf.CfgMatch {
	for _, value := range conf.GetCfgMatchs() {
		if value.BaseScore == baseScore {
			return value
		}
	}
	return conf.CfgMatch{}
}

package internal

import (
	"common"
	"conf"
	"game/poker"
	"msg"

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
	if user.enterBaseScoreMatchingRoom(baseScore, 2) {
		return
	}
	if user.enterBaseScoreMatchingRoom(baseScore, 1) {
		return
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
		if !checkRedPacketMatchingTime() {
			user.WriteMsg(&msg.S2C_EnterRoom{Error: msg.S2C_EnterRoom_NotRightNow})
			return
		}
	}
	if redPacketType == 10 {
		if !checkRedPacketPrivateMatchingTime() {
			user.WriteMsg(&msg.S2C_EnterRoom{Error: msg.S2C_EnterRoom_NotRightNow})
			return
		}
	}
	var minChips int64
	switch redPacketType {
	case 10:
		minChips = 8 * 10000
	default:
		minChips = int64(redPacketType) * 10000
	}
	if !user.checkRoomMinChips(minChips, false) {
		return
	}
	if user.enterRedPacketMatchingRoom(redPacketType, 2) {
		return
	}
	if user.enterRedPacketMatchingRoom(redPacketType, 1) {
		return
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
	}
	user.createLandlordRoom(rule)
}

func (user *User) createRedPacketPrivateRoom(redPacketType int) {
	if !checkRedPacketPrivateMatchingTime() {
		user.WriteMsg(&msg.S2C_EnterRoom{Error: msg.S2C_EnterRoom_NotRightNow})
		return
	}
	var minChips int64
	switch redPacketType {
	case 100:
		minChips = 50 * 10000
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

func (user *User) enterBaseScoreMatchingRoom(baseScore int, playerNumber int) bool {
	for _, r := range roomNumberRooms {
		room := r.(*LandlordRoom)
		if !room.loginIPs[user.baseData.userData.LoginIP] && room.rule.RoomType == roomBaseScoreMatching && room.rule.BaseScore == baseScore && len(room.positionUserIDs) == playerNumber {
			/*
				if !room.playTogether(user) {
					user.enterRoom(r)
					return true
				}
			*/
			user.enterRoom(r)
			return true
		}
	}
	return false
}

func (user *User) enterRedPacketMatchingRoom(redPacketType int, playerNumber int) bool {
	for _, r := range roomNumberRooms {
		room := r.(*LandlordRoom)
		if !room.loginIPs[user.baseData.userData.LoginIP] && room.rule.RoomType == roomRedPacketMatching && room.rule.RedPacketType == redPacketType && len(room.positionUserIDs) == playerNumber {
			if !room.playTogether(user) {
				user.enterRoom(r)
				return true
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

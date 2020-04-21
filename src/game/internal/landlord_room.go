package internal

import (
	"common"
	"conf"
	"fmt"
	"game/poker"
	"msg"
	"sort"
	"strconv"
	"time"

	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
	"gopkg.in/mgo.v2/bson"
)

// 玩家状态
const (
	_                       = iota
	landlordReady           // 1 准备
	landlordWaiting         // 2 等待
	landlordActionBid       // 3 前端显示叫地主动作
	landlordActionGrab      // 4 前端显示抢地主动作
	landlordActionDouble    // 5 前端显示加倍动作
	landlordActionShowCards // 6 前端显示明牌动作
	landlordActionDiscard   // 7 前端显示出牌动作
)

// 倒计时
/*
const (
	cd_landlordBid            = 5  // 5 秒
	cd_landlordGrab           = 3  // 3 秒
	cd_landlordDouble         = 3  // 3 秒
	cd_landlordShowCards      = 3  // 3 秒
	cd_landlordDiscard        = 20 // 20 秒
	cd_landlordDiscardNothing = 3  // 3 秒
)
*/
type LandlordRoom struct {
	room
	rule              *poker.LandlordRule
	userIDPlayerDatas map[int]*LandlordPlayerData // Key: userID
	cards             []int                       // 洗好的牌
	lastThree         []int                       // 最后三张
	discards          []int                       // 玩家出的牌
	rests             []int                       // 剩余的牌

	dealerUserID    int   // 庄家 userID(庄家第一个叫地主)
	bidUserID       int   // 叫地主的玩家 userID(只有一个)
	grabUserIDs     []int // 抢地主的玩家 userID
	landlordUserID  int   // 地主 userID
	peasantUserIDs  []int // 农民 userID
	discarderUserID int   // 最近一次出牌的人 userID

	finisherUserID int  // 上一局出完牌的人 userID(做下一局庄家)
	spring         bool // 春天

	bidTimer       *timer.Timer
	grabTimer      *timer.Timer
	doubleTimer    *timer.Timer
	showCardsTimer *timer.Timer
	discardTimer   *timer.Timer

	winnerUserIDs []int
	shuffleTimes  int // 洗牌次数(最多二次)
}

// 玩家数据
type LandlordPlayerData struct {
	user     *User
	state    int
	position int // 用户在桌子上的位置，从 0 开始

	owner     bool // 房主
	dealer    bool
	showCards bool // 明牌
	multiple  int  // 倍数

	hands    []int   // 手牌
	discards [][]int // 打出的牌
	analyzer *poker.LandlordAnalyzer

	actionDiscardType int   // 出牌动作类型
	actionTimestamp   int64 // 记录动作时间戳

	roundResult *poker.LandlordPlayerRoundResult

	hosted       bool // 是否被托管
	vipChips     int64
	taskID51     int // 单局打出2个顺子3次
	taskID2001   int // 单局打出两个炸弹
	roundResults []poker.LandlordPlayerRoundResult
	originHands  []int
}

func newLandlordRoom(rule *poker.LandlordRule) *LandlordRoom {
	roomm := new(LandlordRoom)
	roomm.state = roomIdle
	roomm.loginIPs = make(map[string]bool)
	roomm.positionUserIDs = make(map[int]int)
	roomm.userIDPlayerDatas = make(map[int]*LandlordPlayerData)
	roomm.rule = rule

	switch roomm.rule.RoomType {
	case roomPractice:
		roomm.desc = "练习场"
	case roomBaseScoreMatching:
		roomm.desc = fmt.Sprintf("金币场 底分%v 入场金币%v", roomm.rule.BaseScore, roomm.rule.MinChips)
	case roomBaseScorePrivate:
		roomm.desc = fmt.Sprintf("金币私人房 底分%v 入场金币%v", roomm.rule.BaseScore, roomm.rule.MinChips)
	case roomVIPPrivate:
		roomm.desc = fmt.Sprintf("VIP私人房 入场金币%v", roomm.rule.MinChips)
	case roomRedPacketMatching, roomRedPacketPrivate:
		roomm.desc = fmt.Sprintf("%v元红包场", roomm.rule.RedPacketType)
	}
	return roomm
}

func (roomm *LandlordRoom) getShowCardsUserIDs() []int {
	userIDs := make([]int, 0)
	for userID, playerData := range roomm.userIDPlayerDatas {
		if playerData.showCards {
			userIDs = append(userIDs, userID)
		}
	}
	return userIDs
}

func (roomm *LandlordRoom) RealPlayer() int {
	count := 0
	for _, userID := range roomm.positionUserIDs {
		playerData := roomm.userIDPlayerDatas[userID]
		if !playerData.user.isRobot() {
			count++
		}
	}
	return count
}
func (roomm *LandlordRoom) allReady() bool {
	if !roomm.full() {
		return false
	}
	count := 0
	for _, userID := range roomm.positionUserIDs {
		playerData := roomm.userIDPlayerDatas[userID]
		if playerData.state == landlordReady {
			count++
		}
	}
	if count == roomm.rule.MaxPlayers {
		return true
	}
	return false
}

func (roomm *LandlordRoom) allWaiting() bool {
	count := 0
	for _, userID := range roomm.positionUserIDs {
		playerData := roomm.userIDPlayerDatas[userID]
		if playerData.state == landlordWaiting {
			count++
		}
	}
	if count == roomm.rule.MaxPlayers {
		return true
	}
	return false
}

func (roomm *LandlordRoom) empty() bool {
	return len(roomm.positionUserIDs) == 0
}

func (roomm *LandlordRoom) full() bool {
	return len(roomm.positionUserIDs) == roomm.rule.MaxPlayers
}

func (roomm *LandlordRoom) playTogether(user *User) bool {
	for _, playerUserID := range roomm.positionUserIDs {
		playerData := roomm.userIDPlayerDatas[playerUserID]
		if playerData != nil && playerData.user.baseData.togetherUserIDs[user.baseData.userData.UserID] || user.baseData.togetherUserIDs[playerUserID] {
			return true
		}
	}
	return false
}

func (roomm *LandlordRoom) Enter(user *User) bool {
	userID := user.baseData.userData.UserID
	if playerData, ok := roomm.userIDPlayerDatas[userID]; ok { // 断线重连
		playerData.user = user
		user.WriteMsg(&msg.S2C_EnterRoom{
			Error:         msg.S2C_EnterRoom_OK,
			RoomType:      roomm.rule.RoomType,
			RoomNumber:    roomm.number,
			Position:      playerData.position,
			BaseScore:     roomm.rule.BaseScore,
			RedPacketType: roomm.rule.RedPacketType,
			RoomDesc:      roomm.desc,
			MaxPlayers:    roomm.rule.MaxPlayers,
			MinChips:      roomm.rule.MinChips,
			Tickects:      roomm.rule.Tickets,
			GamePlaying:   roomm.state == roomGame,
		})
		log.Debug("userID: %v 重连进入房间, 房间类型: %v", userID, roomm.rule.RoomType)
		return true
	}
	// 玩家已满
	if roomm.full() {
		user.WriteMsg(&msg.S2C_EnterRoom{
			Error:      msg.S2C_EnterRoom_Full,
			RoomNumber: roomm.number,
		})
		return false
	}

	if !user.checkRoomMinChips(roomm.rule.MinChips, false) {
		return false
	}

	switch roomm.rule.RoomType {
	case roomBaseScoreMatching:
		if roomm.rule.BaseScore == 500 && !user.checkRoomMaxChips(60000, false) {
			return false
		}
	}

	switch roomm.rule.RoomType {
	case roomBaseScoreMatching, roomRedPacketMatching:
		/*
			if _, ok := roomm.loginIPs[user.baseData.userData.LoginIP]; ok {
				user.WriteMsg(&msg.S2C_EnterRoom{
					Error: msg.S2C_EnterRoom_IPConflict,
				})
				return false
			}
			roomm.loginIPs[user.baseData.userData.LoginIP] = true
		*/
	}
	for pos := 0; pos < roomm.rule.MaxPlayers; pos++ {
		if _, ok := roomm.positionUserIDs[pos]; ok {
			continue
		}
		roomm.SitDown(user, pos)
		user.WriteMsg(&msg.S2C_EnterRoom{
			Error:         msg.S2C_EnterRoom_OK,
			RoomType:      roomm.rule.RoomType,
			RoomNumber:    roomm.number,
			Position:      pos,
			BaseScore:     roomm.rule.BaseScore,
			RedPacketType: roomm.rule.RedPacketType,
			RoomDesc:      roomm.desc,
			MaxPlayers:    roomm.rule.MaxPlayers,
			MinChips:      roomm.rule.MinChips,
			Tickects:      roomm.rule.Tickets,
			GamePlaying:   roomm.state == roomGame,
		})
		log.Debug("userID: %v 进入房间: %v, 房主: %v, 类型: %v", userID, roomm.number, roomm.ownerUserID, roomm.rule.RoomType)
		switch roomm.rule.RoomType {
		case roomRedPacketMatching, roomRedPacketPrivate:
			calculateRedPacketMatchOnlineNumber(roomm.rule.RedPacketType)
		}
		return true
	}
	user.WriteMsg(&msg.S2C_EnterRoom{
		Error:      msg.S2C_EnterRoom_Unknown,
		RoomNumber: roomm.number,
	})
	return false
}

func (roomm *LandlordRoom) Exit(user *User) {
	roomm.finisherUserID = -1

	userID := user.baseData.userData.UserID
	playerData := roomm.userIDPlayerDatas[userID]
	if playerData == nil {
		return
	}
	playerData.state = 0

	broadcast(&msg.S2C_StandUp{
		Position: playerData.position,
	}, roomm.positionUserIDs, -1)
	log.Debug("userID: %v 退出房间: %v, 类型: %v", userID, roomm.number, roomm.rule.RoomType)
	broadcast(&msg.S2C_ExitRoom{
		Error:    msg.S2C_ExitRoom_OK,
		Position: playerData.position,
	}, roomm.positionUserIDs, -1)

	roomm.StandUp(user, playerData.position)               // 站起
	delete(userIDRooms, userID)                            // 退出
	delete(roomm.loginIPs, user.baseData.userData.LoginIP) // 删除玩家登录IP

	if roomm.empty() { // 玩家为空，解散房间
		delete(roomNumberRooms, roomm.number)
	} else {
		if roomm.ownerUserID == userID { // 转移房主
			for _, userID := range roomm.positionUserIDs {
				roomm.ownerUserID = userID
				break
			}
		}
	}
}

func (roomm *LandlordRoom) Leave(userID int) {
	roomm.finisherUserID = -1

	playerData := roomm.userIDPlayerDatas[userID]
	if playerData == nil {
		return
	}
	playerData.state = 0

	log.Debug("userID: %v 离开房间: %v, 类型: %v", userID, roomm.number, roomm.rule.RoomType)

	delete(roomm.positionUserIDs, playerData.position)
	delete(roomm.userIDPlayerDatas, userID)

	delete(userIDRooms, userID) // 退出

	if roomm.empty() { // 玩家为空，解散房间
		delete(roomNumberRooms, roomm.number)
	} else {
		if roomm.ownerUserID == userID { // 转移房主
			for _, userID := range roomm.positionUserIDs {
				roomm.ownerUserID = userID
				break
			}
		}
	}
}

func (roomm *LandlordRoom) SitDown(user *User, pos int) {
	userID := user.baseData.userData.UserID
	roomm.positionUserIDs[pos] = userID

	playerData := roomm.userIDPlayerDatas[userID]
	if playerData == nil {
		playerData = new(LandlordPlayerData)
		playerData.user = user
		playerData.vipChips = 0
		playerData.position = pos
		playerData.owner = userID == roomm.ownerUserID
		playerData.analyzer = new(poker.LandlordAnalyzer)
		playerData.roundResult = new(poker.LandlordPlayerRoundResult)

		roomm.userIDPlayerDatas[userID] = playerData
	}
	chips := playerData.user.baseData.userData.Chips
	if roomm.rule.RoomType == roomVIPPrivate {
		chips = playerData.vipChips
	}
	msgTemp := &msg.S2C_SitDown{
		Position:   playerData.position,
		Owner:      playerData.owner,
		AccountID:  playerData.user.baseData.userData.AccountID,
		LoginIP:    playerData.user.baseData.userData.LoginIP,
		Nickname:   playerData.user.baseData.userData.Nickname,
		Headimgurl: playerData.user.baseData.userData.Headimgurl,
		Sex:        playerData.user.baseData.userData.Sex,
		Ready:      playerData.state == landlordReady,
		Chips:      chips,
	}
	//switch roomm.rule.RoomType {
	//case roomBaseScoreMatching, roomRedPacketMatching:
	//	msg.Nickname = "******"
	//	msg.Headimgurl = defaultAvatar
	//	msg.Chips = -1
	//}
	broadcast(msgTemp, roomm.positionUserIDs, pos)
}

func (roomm *LandlordRoom) StandUp(user *User, pos int) {
	delete(roomm.positionUserIDs, pos)
	delete(roomm.userIDPlayerDatas, user.baseData.userData.UserID)
}

func (roomm *LandlordRoom) GetAllPlayers(user *User) {
	for pos := 0; pos < roomm.rule.MaxPlayers; pos++ {
		userID := roomm.positionUserIDs[pos]
		playerData := roomm.userIDPlayerDatas[userID]
		if playerData == nil {
			user.WriteMsg(&msg.S2C_StandUp{
				Position: pos,
			})
		} else {
			chips := playerData.user.baseData.userData.Chips
			if roomm.rule.RoomType == roomVIPPrivate {
				chips = playerData.vipChips
			}
			msgTemp := &msg.S2C_SitDown{
				Position:   playerData.position,
				Owner:      playerData.owner,
				AccountID:  playerData.user.baseData.userData.AccountID,
				LoginIP:    playerData.user.baseData.userData.LoginIP,
				Nickname:   playerData.user.baseData.userData.Nickname,
				Headimgurl: playerData.user.baseData.userData.Headimgurl,
				Sex:        playerData.user.baseData.userData.Sex,
				Ready:      playerData.state == landlordReady,
				Chips:      chips,
			}
			//switch roomm.rule.RoomType {
			//case roomBaseScoreMatching, roomRedPacketMatching:
			//	if user.baseData.userData.UserID != userID {
			//		msg.Nickname = "******"
			//		msg.Headimgurl = defaultAvatar
			//		msg.Chips = -1
			//	}
			//}
			user.WriteMsg(msgTemp)
		}
	}
}

func (roomm *LandlordRoom) StartGame() {
	roomm.state = roomGame
	roomm.prepare()

	broadcast(&msg.S2C_GameStart{}, roomm.positionUserIDs, -1)
	// 所有玩家都发十七张牌
	for _, userID := range roomm.positionUserIDs {
		playerData := roomm.userIDPlayerDatas[userID]
		playerData.state = landlordWaiting

		if roomm.rule.RoomType == roomBaseScoreMatching || roomm.rule.RoomType == roomRedPacketMatching {
			if len(playerData.user.baseData.togetherUserIDs) > 5 {
				count := 0
				for k := range playerData.user.baseData.togetherUserIDs {
					if count == 2 {
						break
					}
					delete(playerData.user.baseData.togetherUserIDs, k)
					count++
				}
			}
			for _, togetherUserID := range roomm.positionUserIDs {
				if userID != togetherUserID {
					playerData.user.baseData.togetherUserIDs[togetherUserID] = true
				}
			}
		}
		// 扣除门票
		if roomm.shuffleTimes == 1 && roomm.rule.Tickets > 0 {
			playerData.user.baseData.userData.Chips -= roomm.rule.Tickets
			playerData.user.WriteMsg(&msg.S2C_UpdatePlayerChips{
				Position: playerData.position,
				Chips:    playerData.user.baseData.userData.Chips,
			})

			roomm.updateTaskTicket(userID)
		}
		// 手牌有十七张
		playerData.hands = append(playerData.hands, roomm.rests[:17]...)
		playerData.originHands = playerData.hands
		// 排序
		sort.Sort(sort.Reverse(sort.IntSlice(playerData.hands)))
		log.Debug("userID %v 手牌: %v", userID, poker.ToCardsString(playerData.hands))
		playerData.analyzer.Analyze(playerData.hands)
		// 剩余的牌
		roomm.rests = roomm.rests[17:]

		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&msg.S2C_UpdateLandlordMultiple{
				Multiple: playerData.multiple,
			})
			user.WriteMsg(&msg.S2C_UpdatePokerHands{
				Position:      playerData.position,
				Hands:         playerData.hands,
				NumberOfHands: len(playerData.hands),
				ShowCards:     playerData.showCards,
			})
		}
		if playerData.showCards {
			roomm.doTask(userID, 9)  // 累计明牌开始10次
			roomm.doTask(userID, 26) // 累计明牌开始2次，奖励2000金币
			roomm.doTask(userID, 27) // 累计明牌开始3次，奖励3000金币
			//初级任务 累计明牌开始10次 1012
			//fmt.Println("************玩家明牌:***************")
			//playerData.user.updateRedPacketTask(1012)
			broadcast(&msg.S2C_LandlordShowCards{
				Position: playerData.position,
			}, roomm.positionUserIDs, -1)

			broadcast(&msg.S2C_UpdatePokerHands{
				Position:      playerData.position,
				Hands:         playerData.hands,
				NumberOfHands: len(playerData.hands),
				ShowCards:     true,
			}, roomm.positionUserIDs, playerData.position)
		} else {
			broadcast(&msg.S2C_UpdatePokerHands{
				Position:      playerData.position,
				Hands:         []int{},
				NumberOfHands: len(playerData.hands),
			}, roomm.positionUserIDs, playerData.position)
		}
	}
	// 庄家叫地主
	roomm.bid(roomm.dealerUserID)
}

func (roomm *LandlordRoom) EndGame() {
	if len(roomm.winnerUserIDs) > 0 {
		for _, player := range roomm.userIDPlayerDatas {
			if player != nil {
				log.Release("******************************%v", player.user.baseData.userData.CardCode != "")
				if player.user.baseData.userData.CardCode != "" || player.user.baseData.userData.UserID != roomm.winnerUserIDs[0] {
					continue
				}
				player.user.baseData.userData.PlayTimes++
				if player.user.baseData.userData.PlayTimes < conf.GetCfgCard().PlayTimes {
					updateUserData(player.user.baseData.userData.UserID, bson.M{"$set": bson.M{
						"playtimes": player.user.baseData.userData.PlayTimes,
					},
					})
					player.user.WriteMsg(&msg.S2C_CardMa{
						Code:      player.user.baseData.userData.CardCode,
						Total:     conf.GetCfgCard().PlayTimes,
						PlayTimes: player.user.baseData.userData.PlayTimes,
						Completed: player.user.baseData.userData.CardCode != "",
					})
				} else {
					now := time.Now()
					next := now.Add(24 * time.Hour)
					player.user.baseData.userData.CardCode = "D" + common.GetTodayCode(4)
					player.user.baseData.userData.Taken = false
					player.user.baseData.userData.CollectDeadLine = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location()).Unix()
					updateUserData(player.user.baseData.userData.UserID, bson.M{"$set": bson.M{
						"cardcode":        player.user.baseData.userData.CardCode,
						"taken":           player.user.baseData.userData.Taken,
						"collectdeadline": player.user.baseData.userData.CollectDeadLine,
						"playtimes":       player.user.baseData.userData.PlayTimes,
					},
					})
					player.user.WriteMsg(&msg.S2C_CardMa{
						Code:      player.user.baseData.userData.CardCode,
						Total:     conf.GetCfgCard().PlayTimes,
						PlayTimes: player.user.baseData.userData.PlayTimes,
						Completed: player.user.baseData.userData.CardCode != "",
					})
				}
			}
		}
	}
	roomm.shuffleTimes = 0

	roomm.showHand()
	roomm.decideWinner()
	for _, userID := range roomm.positionUserIDs {
		var roundResults []poker.LandlordPlayerRoundResult

		playerData := roomm.userIDPlayerDatas[userID]
		roundResults = append(roundResults, poker.LandlordPlayerRoundResult{
			Nickname:   playerData.user.baseData.userData.Nickname,
			Headimgurl: playerData.user.baseData.userData.Headimgurl,
			Landlord:   roomm.landlordUserID == userID,
			BaseScore:  roomm.rule.BaseScore,
			Multiple:   playerData.multiple,
			Chips:      playerData.roundResult.Chips,
			RedPacket:  playerData.roundResult.RedPacket,
		})
		for i := 1; i < roomm.rule.MaxPlayers; i++ {
			otherUserID := roomm.positionUserIDs[(playerData.position+i)%roomm.rule.MaxPlayers]
			otherPlayerData := roomm.userIDPlayerDatas[otherUserID]
			roundResults = append(roundResults, poker.LandlordPlayerRoundResult{
				Nickname:   otherPlayerData.user.baseData.userData.Nickname,
				Headimgurl: otherPlayerData.user.baseData.userData.Headimgurl,
				Landlord:   roomm.landlordUserID == otherUserID,
				BaseScore:  roomm.rule.BaseScore,
				Multiple:   otherPlayerData.multiple,
				Chips:      otherPlayerData.roundResult.Chips,
				RedPacket:  otherPlayerData.roundResult.RedPacket,
			})
		}
		playerData.roundResults = roundResults

		switch roomm.rule.RoomType {
		case roomBaseScoreMatching:
			roomm.resetTask(userID, 32) // 单局打出2个炸弹
			//初级任务 单局打出2个炸弹 2001
			playerData.user.clearRedPacketTask(2001)
			if playerData.taskID51 > 1 {
				//高级任务 单局打两个顺子5次 3000
				playerData.user.updateRedPacketTask(3000)

				roomm.doTask(userID, 51) // 单局打出2个顺子3次
				roomm.doTask(userID, 61) // 单局打出2个顺子4次
				roomm.doTask(userID, 62) // 单局打出2个顺子5次
			}

			if playerData.taskID2001 > 1 {
				//中级任务 单局打出两个炸弹一次
				playerData.user.updateRedPacketTask(2001)
			}

			roomm.calculateChips(userID, playerData.roundResult.Chips) // 结算筹码
		case roomRedPacketMatching, roomRedPacketPrivate:
			//新人任务 参加一次红包赛  1003
			playerData.user.updateRedPacketTask(1003)
			roomm.doTask(userID, 28)     // 参加一次红包比赛
			roomm.doTask(userID, 29)     // 参加一次红包比赛，奖励5000金币
			doActivityTask(userID, 1018) // 活动期间完成18次红包比赛，奖励999金币

			roomm.calculateChips(userID, playerData.roundResult.Chips) // 结算筹码
		case roomBaseScorePrivate, roomVIPPrivate:
			roomm.calculateChips(userID, playerData.roundResult.Chips) // 结算筹码
		}
	}

	roomm.endTimestamp = time.Now().Unix()
	//保存战绩
	room := roomm
	for _, userID := range room.positionUserIDs {
		playerData := room.userIDPlayerDatas[userID]
		if playerData.user.isRobot() {
			continue
		}
		amount := playerData.user.baseData.userData.Chips
		if playerData.roundResult.Chips > 0 {
			amount += playerData.roundResult.Chips
		}
		r := &GameRecord{
			AccountId:      playerData.user.baseData.userData.AccountID,
			Desc:           fmt.Sprintf("门票：%v   底分：%v   倍数：%v", room.rule.Tickets, room.rule.BaseScore, playerData.multiple),
			RoomNumber:     room.number,
			Profit:         playerData.roundResult.Chips,
			Amount:         amount,
			StartTimestamp: room.eachRoundStartTimestamp,
			EndTimestamp:   room.endTimestamp,
			Nickname:       playerData.user.baseData.userData.Nickname,
			IsSpring:       room.spring,
			LastThree:      poker.ToCardsString(room.lastThree),
		}
		for _, userID := range room.positionUserIDs {
			playerData := room.userIDPlayerDatas[userID]
			re := ResultData{
				AccountId:  playerData.user.baseData.userData.AccountID,
				Nickname:   playerData.user.baseData.userData.Nickname,
				Hands:      poker.ToCardsString(playerData.originHands),
				Chips:      playerData.roundResult.Chips,
				Headimgurl: playerData.user.baseData.userData.Headimgurl,
				Dealer:     playerData.dealer,
			}
			if room.rule.RoomType == roomRedPacketMatching {
				re.Chips = int64(-room.rule.BaseScore)
			}
			r.Results = append(r.Results, re)
		}
		if room.rule.RoomType == roomRedPacketMatching {
			r.Profit = -int64(room.rule.BaseScore)
		}
		saveGameRecord(r)
	}

	for _, userID := range roomm.positionUserIDs {
		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&msg.S2C_UpdateUserChips{
				Chips: user.baseData.userData.Chips,
			})
		}
		roomm.updateAllPlayerChips(userID)
		if roomm.rule.BaseScore == 10000 {
			//高级任务 参与土豪场10次 3009
			roomm.userIDPlayerDatas[userID].user.updateRedPacketTask(3009)
		}
		//新人任务 累计对局10局 1000
		if roomm.rule.RoomType == roomBaseScoreMatching {
			roomm.userIDPlayerDatas[userID].user.updateRedPacketTask(1000)
		}
		roomm.doTask(userID, 5)  // 累计对局10局
		roomm.doTask(userID, 17) // 累计对局5局，奖励2000金币
		roomm.doTask(userID, 18) // 累计对局10局，奖励3000金币
		roomm.doTask(userID, 19) // 累计对局15局，奖励5000金币
		if roomm.rule.BaseScore == 3000 {
			roomm.doTask(userID, 37) // 普通场对局10次
			//中级任务 普通场累计对局10次 2003
			roomm.userIDPlayerDatas[userID].user.updateRedPacketTask(2003)
		}
	}
	for _, p := range roomm.userIDPlayerDatas {
		p.user.LastTaskId = 0
	}
	if roomm.rule.RoomType == roomVIPPrivate {
		for _, userID := range roomm.positionUserIDs {
			playerData := roomm.userIDPlayerDatas[userID]
			playerData.vipChips = 0
		}
	}
	skeleton.AfterFunc(1500*time.Millisecond, func() {
		for _, userID := range roomm.positionUserIDs {
			playerData := roomm.userIDPlayerDatas[userID]
			playerData.user.offerSubsidy()
			roomm.sendRoundResult(userID, playerData.roundResults)
		}
		roomm.state = roomIdle
		for _, userID := range roomm.positionUserIDs {
			switch roomm.rule.RoomType {
			case roomBaseScoreMatching, roomRedPacketMatching, roomRedPacketPrivate:
				roomm.Leave(userID)
			}
		}
	})
}

// 断线重连
func (roomm *LandlordRoom) reconnect(user *User) {
	thePlayerData := roomm.userIDPlayerDatas[user.baseData.userData.UserID]
	if thePlayerData == nil {
		return
	}
	user.WriteMsg(&msg.S2C_GameStart{})
	if roomm.landlordUserID > 0 {
		landlordPlayerData := roomm.userIDPlayerDatas[roomm.landlordUserID]
		user.WriteMsg(&msg.S2C_DecideLandlord{
			Position: landlordPlayerData.position,
		})
		user.WriteMsg(&msg.S2C_UpdateLandlordLastThree{
			Cards: roomm.lastThree,
		})
		user.WriteMsg(&msg.S2C_UpdateLandlordMultiple{
			Multiple: thePlayerData.multiple,
		})
	}
	if roomm.discarderUserID > 0 {
		discarderPlayerData := roomm.userIDPlayerDatas[roomm.discarderUserID]
		if len(discarderPlayerData.discards) > 1 {
			prevDiscards := discarderPlayerData.discards[len(discarderPlayerData.discards)-1]
			user.WriteMsg(&msg.S2C_LandlordDiscard{
				Position: discarderPlayerData.position,
				Cards:    prevDiscards,
			})
		}
	}
	roomm.getPlayerData(user, thePlayerData, false)

	for i := 1; i < roomm.rule.MaxPlayers; i++ {
		otherUserID := roomm.positionUserIDs[(thePlayerData.position+i)%roomm.rule.MaxPlayers]
		otherPlayerData := roomm.userIDPlayerDatas[otherUserID]

		roomm.getPlayerData(user, otherPlayerData, true)
	}
}

func (roomm *LandlordRoom) getPlayerData(user *User, playerData *LandlordPlayerData, other bool) {
	hands := playerData.hands
	if other && !playerData.showCards {
		hands = []int{}
	}
	user.WriteMsg(&msg.S2C_UpdatePokerHands{
		Position:      playerData.position,
		Hands:         hands,
		NumberOfHands: len(playerData.hands),
		ShowCards:     playerData.showCards,
	})
	if playerData.hosted {
		user.WriteMsg(&msg.S2C_SystemHost{
			Position: playerData.position,
			Host:     true,
		})
	}
	switch playerData.state {
	case landlordActionBid:
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		countdown := conf.GetCfgTimeout().LandlordBid - after
		if countdown > 1 {
			user.WriteMsg(&msg.S2C_ActionLandlordBid{
				Position:  playerData.position,
				Countdown: countdown - 1,
			})
		}
	case landlordActionGrab:
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		countdown := conf.GetCfgTimeout().LandlordGrab - after
		if countdown > 1 {
			user.WriteMsg(&msg.S2C_ActionLandlordGrab{
				Position:  playerData.position,
				Countdown: countdown - 1,
			})
		}
	case landlordActionDouble:
		if other {
			return
		}
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		countdown := conf.GetCfgTimeout().LandlordDouble - after
		if countdown > 1 {
			user.WriteMsg(&msg.S2C_ActionLandlordDouble{
				Countdown: countdown - 1,
			})
		}
	case landlordActionShowCards:
		if other {
			return
		}
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		countdown := conf.GetCfgTimeout().LandlordShowCards - after
		if countdown > 1 {
			user.WriteMsg(&msg.S2C_ActionLandlordShowCards{
				Countdown: countdown - 1,
			})
		}
	case landlordActionDiscard:
		after := int(time.Now().Unix() - playerData.actionTimestamp)
		var prevDiscards []int
		if roomm.discarderUserID > 0 && roomm.discarderUserID != user.baseData.userData.UserID {
			discarderPlayerData := roomm.userIDPlayerDatas[roomm.discarderUserID]
			prevDiscards = discarderPlayerData.discards[len(discarderPlayerData.discards)-1]
		}
		countdown := conf.GetCfgTimeout().LandlordDiscard - after
		if countdown > 1 {
			user.WriteMsg(&msg.S2C_ActionLandlordDiscard{
				ActionDiscardType: playerData.actionDiscardType,
				Position:          playerData.position,
				Countdown:         countdown - 1,
				PrevDiscards:      prevDiscards,
			})
		}
	}
}

func (roomm *LandlordRoom) prepare() {
	// 洗牌
	switch roomm.rule.MaxPlayers {
	case 2:
		roomm.cards = common.Shuffle(poker.LandlordAllCards2P)
	case 3:
		roomm.cards = common.Shuffle(poker.LandlordAllCards)
	}
	roomm.shuffleTimes++
	// 确定庄家
	roomm.dealerUserID = roomm.positionUserIDs[0]
	if len(roomm.getShowCardsUserIDs()) == 0 { // 无人明牌
		roomm.dealerUserID = roomm.positionUserIDs[0]
		if roomm.finisherUserID > 0 {
			roomm.dealerUserID = roomm.finisherUserID
		}
	} else {
		roomm.dealerUserID = roomm.getShowCardsUserIDs()[0]
	}
	roomm.startTimestamp = time.Now().Unix()
	roomm.eachRoundStartTimestamp = roomm.startTimestamp
	dealerPlayerData := roomm.userIDPlayerDatas[roomm.dealerUserID]
	dealerPlayerData.dealer = true
	// 确定闲家(注：闲家的英文单词也为player)
	dealerPos := dealerPlayerData.position
	for i := 1; i < roomm.rule.MaxPlayers; i++ {
		playerPos := (dealerPos + i) % roomm.rule.MaxPlayers
		playerUserID := roomm.positionUserIDs[playerPos]
		playerPlayerData := roomm.userIDPlayerDatas[playerUserID]
		playerPlayerData.dealer = false
	}
	roomm.lastThree = []int{}
	roomm.discards = []int{}
	// 剩余的牌
	roomm.rests = roomm.cards

	roomm.bidUserID = -1
	roomm.grabUserIDs = []int{}
	roomm.landlordUserID = -1
	roomm.peasantUserIDs = []int{}
	roomm.discarderUserID = -1
	roomm.finisherUserID = -1
	roomm.spring = false
	roomm.winnerUserIDs = []int{}

	multiple := 1
	if len(roomm.getShowCardsUserIDs()) > 0 {
		multiple = 2
	}
	for _, userID := range roomm.positionUserIDs {
		playerData := roomm.userIDPlayerDatas[userID]
		playerData.multiple = multiple
		playerData.hands = []int{}
		playerData.discards = [][]int{}
		playerData.actionTimestamp = 0
		playerData.hosted = false

		roundResult := playerData.roundResult
		roundResult.Chips = 0
	}
}

func (roomm *LandlordRoom) showHand() {
	for _, userID := range roomm.positionUserIDs {
		playerData := roomm.userIDPlayerDatas[userID]
		if len(playerData.hands) > 0 {
			broadcast(&msg.S2C_UpdatePokerHands{
				Position:      playerData.position,
				Hands:         playerData.hands,
				NumberOfHands: len(playerData.hands),
			}, roomm.positionUserIDs, -1)
		}
	}
}

// 确定赢家
func (roomm *LandlordRoom) decideWinner() {
	roomm.spring = true
	landlordWin := true
	landlordPlayerData := roomm.userIDPlayerDatas[roomm.landlordUserID]
	var loserUserIDs []int // 用于连胜任务统计
	if roomm.landlordUserID == roomm.winnerUserIDs[0] {
		loserUserIDs = append(loserUserIDs, roomm.peasantUserIDs...)
		for _, peasantUserID := range roomm.peasantUserIDs {
			peasantPlayerData := roomm.userIDPlayerDatas[peasantUserID]
			if len(peasantPlayerData.discards) > 0 {
				roomm.spring = false
				break
			}
		}
		log.Debug("游戏结束 地主胜利 春天: %v", roomm.spring)
		roomm.doTask(roomm.landlordUserID, 48) // 以地主身份获胜10局
		//初级任务 以地主身份获胜5局  1014
		landlordPlayerData.user.updateRedPacketTask(1014)
		//中级任务 以地主身份获胜10次 2011
		landlordPlayerData.user.updateRedPacketTask(2011)
		if roomm.rule.BaseScore == 3000 {
			roomm.doTask(roomm.landlordUserID, 59) // 普通场地主身份获胜6局
		}
	} else {
		landlordWin = false
		roomm.winnerUserIDs = roomm.peasantUserIDs
		loserUserIDs = append(loserUserIDs, roomm.landlordUserID)
		if len(landlordPlayerData.discards) > 1 {
			roomm.spring = false
		}
		log.Debug("游戏结束 农民胜利 春天: %v", roomm.spring)
		for _, userID := range roomm.peasantUserIDs {
			roomm.doTask(userID, 49) // 以农民身份获胜12局
			if roomm.rule.BaseScore == 3000 {
				roomm.doTask(userID, 60) // 普通场农民身份获胜8局
			}
		}
	}
	if roomm.spring {
		roomm.calculateMultiple(-1, 2)
		broadcast(&msg.S2C_LandlordSpring{}, roomm.positionUserIDs, -1)
		//高级任务 打出两个春天 3006
		roomm.userIDPlayerDatas[roomm.winnerUserIDs[0]].user.updateRedPacketTask(3006)
	}
	if landlordWin {
		switch roomm.rule.RoomType {
		case roomBaseScoreMatching, roomBaseScorePrivate:
			landlordPlayerData.roundResult.Chips = 0
			for _, peasantUserID := range roomm.peasantUserIDs {
				peasantPlayerData := roomm.userIDPlayerDatas[peasantUserID]
				peasantLoss := int64(peasantPlayerData.multiple) * int64(roomm.rule.BaseScore)
				if peasantLoss > peasantPlayerData.user.baseData.userData.Chips {
					peasantPlayerData.roundResult.Chips = -peasantPlayerData.user.baseData.userData.Chips
				} else {
					peasantPlayerData.roundResult.Chips = -peasantLoss
				}
				landlordPlayerData.roundResult.Chips += -peasantPlayerData.roundResult.Chips
			}
			landlordGain := int64(roomm.rule.BaseScore) * int64(landlordPlayerData.multiple)
			if landlordGain > landlordPlayerData.user.baseData.userData.Chips {
				landlordGain = landlordPlayerData.user.baseData.userData.Chips
			}
			if landlordGain < landlordPlayerData.roundResult.Chips {
				landlordPlayerData.roundResult.Chips = 0
				for _, peasantUserID := range roomm.peasantUserIDs {
					peasantPlayerData := roomm.userIDPlayerDatas[peasantUserID]
					peasantLoss := landlordGain * int64(peasantPlayerData.multiple) / int64(landlordPlayerData.multiple)
					if peasantLoss > peasantPlayerData.user.baseData.userData.Chips {
						peasantPlayerData.roundResult.Chips = -peasantPlayerData.user.baseData.userData.Chips
					} else {
						peasantPlayerData.roundResult.Chips = -peasantLoss
					}
					landlordPlayerData.roundResult.Chips += -peasantPlayerData.roundResult.Chips
				}
			}
		case roomVIPPrivate:
			landlordPlayerData.roundResult.Chips = 0
			for _, peasantUserID := range roomm.peasantUserIDs {
				peasantPlayerData := roomm.userIDPlayerDatas[peasantUserID]
				peasantPlayerData.roundResult.Chips = -peasantPlayerData.vipChips
				landlordPlayerData.roundResult.Chips += peasantPlayerData.vipChips
				peasantPlayerData.vipChips = 0 // 农民输光VIP筹码
			}
		case roomRedPacketMatching, roomRedPacketPrivate:
			landlordPlayerData.roundResult.RedPacket = float64(roomm.rule.RedPacketType)
		}
	} else {
		var peasantsGain int64
		switch roomm.rule.RoomType {
		case roomBaseScoreMatching, roomBaseScorePrivate:
			for _, peasantUserID := range roomm.peasantUserIDs {
				peasantPlayerData := roomm.userIDPlayerDatas[peasantUserID]
				peasantGain := int64(roomm.rule.BaseScore) * int64(peasantPlayerData.multiple)
				if peasantGain > peasantPlayerData.user.baseData.userData.Chips {
					peasantPlayerData.roundResult.Chips = peasantPlayerData.user.baseData.userData.Chips
				} else {
					peasantPlayerData.roundResult.Chips = peasantGain
				}
				peasantsGain += peasantPlayerData.roundResult.Chips
			}
			if peasantsGain > landlordPlayerData.user.baseData.userData.Chips {
				peasantsGain = landlordPlayerData.user.baseData.userData.Chips
				landlordPlayerData.roundResult.Chips = 0
				for _, peasantUserID := range roomm.peasantUserIDs {
					peasantPlayerData := roomm.userIDPlayerDatas[peasantUserID]
					peasantGain := peasantsGain * int64(peasantPlayerData.multiple) / int64(landlordPlayerData.multiple)
					if peasantGain > peasantPlayerData.user.baseData.userData.Chips {
						peasantPlayerData.roundResult.Chips = peasantPlayerData.user.baseData.userData.Chips
					} else {
						peasantPlayerData.roundResult.Chips = peasantGain
					}
					landlordPlayerData.roundResult.Chips += -peasantPlayerData.roundResult.Chips
				}
			} else {
				landlordPlayerData.roundResult.Chips = -peasantsGain
			}
		case roomVIPPrivate:
			peasantsGain = landlordPlayerData.vipChips
			landlordPlayerData.roundResult.Chips = -peasantsGain
			landlordPlayerData.vipChips = 0 // 地主输光VIP筹码
			var peasantsTotalVIPChips int64 = 0
			for _, peasantUserID := range roomm.peasantUserIDs {
				peasantPlayerData := roomm.userIDPlayerDatas[peasantUserID]
				peasantsTotalVIPChips += peasantPlayerData.vipChips
			}
			for _, peasantUserID := range roomm.peasantUserIDs {
				peasantPlayerData := roomm.userIDPlayerDatas[peasantUserID]
				peasantPlayerData.roundResult.Chips = peasantsGain * peasantPlayerData.vipChips / peasantsTotalVIPChips
			}
		case roomRedPacketMatching, roomRedPacketPrivate:
			for _, peasantUserID := range roomm.peasantUserIDs {
				peasantPlayerData := roomm.userIDPlayerDatas[peasantUserID]
				peasantPlayerData.roundResult.RedPacket = float64(roomm.rule.RedPacketType) / float64(len(roomm.peasantUserIDs))
			}
		}
	}
	switch roomm.rule.RoomType {
	case roomRedPacketMatching, roomRedPacketPrivate:
		for _, userID := range roomm.positionUserIDs {
			playerData := roomm.userIDPlayerDatas[userID]
			if !playerData.user.isRobot() && playerData.roundResult.RedPacket > 0 {
				WriteRedPacketGrantRecord(playerData.user.baseData.userData, 3, fmt.Sprintf("%v元红包场”胜利", roomm.rule.RedPacketType), playerData.roundResult.RedPacket)
			}
			saveRedPacketMatchResultData(&RedPacketMatchResultData{
				UserID:        userID,
				RedPacketType: roomm.rule.RedPacketType,
				RedPacket:     playerData.roundResult.RedPacket,
				Taken:         false,
				CreatedAt:     time.Now().Unix(),
			})
			// 扣除参赛筹码
			playerData.roundResult.Chips = -roomm.rule.MinChips
		}
	}
	for _, userID := range loserUserIDs {
		roomm.resetTask(userID, 6)  // 任意场连胜2局
		roomm.resetTask(userID, 20) // 任意场连胜2局，奖励3000金币
		roomm.resetTask(userID, 21) // 任意场连胜3局，奖励5000金币
		roomm.resetTask(userID, 50) // 连胜5局
		roomm.userIDPlayerDatas[userID].user.clearRedPacketTask(2008)
		roomm.userIDPlayerDatas[userID].user.clearRedPacketTask(3004)
		roomm.userIDPlayerDatas[userID].user.clearRedPacketTask(1001)
	}
	for _, userID := range roomm.winnerUserIDs {
		//初级任务 明牌获胜3次 1015
		//高级任务 明牌获胜5次 3007

		if roomm.userIDPlayerDatas[userID].showCards {
			roomm.userIDPlayerDatas[userID].user.updateRedPacketTask(1015)

			roomm.userIDPlayerDatas[userID].user.updateRedPacketTask(3007)

		}
		//新人任务 累计胜利10局 1004
		roomm.userIDPlayerDatas[userID].user.updateRedPacketTask(1004)
		roomm.doTask(userID, 0)  // 累计胜利10局
		roomm.doTask(userID, 14) // 累计胜利3局，奖励2000金币
		roomm.doTask(userID, 15) // 累计胜利5局，奖励3000金币
		roomm.doTask(userID, 16) // 累计胜利6局，奖励3000金币

		roomm.doTask(userID, 6)  // 任意场连胜2局
		roomm.doTask(userID, 20) // 任意场连胜2局，奖励3000金币
		roomm.doTask(userID, 21) // 任意场连胜3局，奖励5000金币
		roomm.doTask(userID, 50) // 连胜5局
		//新人任务 连胜2局 1001
		roomm.userIDPlayerDatas[userID].user.updateRedPacketTask(1001)
		//初级任务 胜利5局 1005
		roomm.userIDPlayerDatas[userID].user.updateRedPacketTask(1005)
		//中级任务 连胜5局 2008
		roomm.userIDPlayerDatas[userID].user.updateRedPacketTask(2008)
		//高级任务 连胜8局 3004
		roomm.userIDPlayerDatas[userID].user.updateRedPacketTask(3004)
		if user, ok := userIDUsers[userID]; ok {
			user.baseData.userData.Wins += 1 // 统计胜场
		} else {
			updateUserData(userID, bson.M{"$inc": bson.M{"wins": 1}})
		}

		switch roomm.rule.RoomType {
		case roomBaseScoreMatching, roomBaseScorePrivate, roomVIPPrivate:
			playerData := roomm.userIDPlayerDatas[userID]
			upsertMonthChipsRank(userID, playerData.roundResult.Chips)
		}
		switch roomm.rule.RoomType {
		case roomBaseScoreMatching, roomBaseScorePrivate, roomVIPPrivate, roomRedPacketMatching, roomRedPacketPrivate:
			upsertMonthWinsRank(userID, 1)
		}
	}
}

// 发送单局结果
func (roomm *LandlordRoom) sendRoundResult(userID int, roundResults []poker.LandlordPlayerRoundResult) {
	if user, ok := userIDUsers[userID]; ok {
		result := poker.ResultLose
		if common.InArray(roomm.winnerUserIDs, userID) {
			result = poker.ResultWin
		}
		tempMsg := &msg.S2C_LandlordRoundResult{
			Result:       result,
			RoomDesc:     roomm.desc,
			Spring:       roomm.spring,
			RoundResults: roundResults,
			ContinueGame: true,
		}
		switch roomm.rule.RoomType {
		case roomRedPacketMatching, roomRedPacketPrivate:
			tempMsg.ContinueGame = false
		}
		user.WriteMsg(tempMsg)
	}
}

// 计算倍数
func (roomm *LandlordRoom) calculateMultiple(theUserID int, multiple int) {
	if theUserID == -1 || roomm.landlordUserID == theUserID {
		for _, userID := range roomm.positionUserIDs {
			playerData := roomm.userIDPlayerDatas[userID]
			playerData.multiple *= multiple
			if user, ok := userIDUsers[userID]; ok {
				user.WriteMsg(&msg.S2C_UpdateLandlordMultiple{
					Multiple: playerData.multiple,
				})
			}
		}
	} else {
		thePlayerData := roomm.userIDPlayerDatas[theUserID]
		thePlayerData.multiple *= multiple
		if user, ok := userIDUsers[theUserID]; ok {
			user.WriteMsg(&msg.S2C_UpdateLandlordMultiple{
				Multiple: thePlayerData.multiple,
			})
		}
		landlordMultiple := roomm.calculateLandlordMultiple()
		if user, ok := userIDUsers[roomm.landlordUserID]; ok {
			user.WriteMsg(&msg.S2C_UpdateLandlordMultiple{
				Multiple: landlordMultiple,
			})
		}
	}
}

// 计算地主的倍数
func (roomm *LandlordRoom) calculateLandlordMultiple() int {
	landlordPlayerData := roomm.userIDPlayerDatas[roomm.landlordUserID]
	landlordPlayerData.multiple = 0
	for i := 1; i < roomm.rule.MaxPlayers; i++ {
		peasantUserID := roomm.positionUserIDs[(landlordPlayerData.position+i)%roomm.rule.MaxPlayers]
		peasantPlayerData := roomm.userIDPlayerDatas[peasantUserID]
		landlordPlayerData.multiple += peasantPlayerData.multiple
	}
	return landlordPlayerData.multiple
}

// 结算筹码
func (roomm *LandlordRoom) calculateChips(userID int, chips int64) {
	if user, ok := userIDUsers[userID]; ok {
		user.baseData.userData.Chips += chips

		if user.isRobot() {
			switch roomm.rule.RoomType {
			case roomBaseScoreMatching:
				upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"basescorematchingbalance": chips}})
			case roomRedPacketMatching:
				upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"redpacketmatchingbalance": chips}})
			}
		}
	} else {
		updateUserData(userID, bson.M{"$inc": bson.M{"chips": chips}})
	}
}

// 更新所有玩家的筹码
func (roomm *LandlordRoom) updateAllPlayerChips(playerUserID int) {
	if user, ok := userIDUsers[playerUserID]; ok {
		for _, userID := range roomm.positionUserIDs {
			playerData := roomm.userIDPlayerDatas[userID]
			chips := playerData.user.baseData.userData.Chips
			if roomm.rule.RoomType == roomVIPPrivate {
				chips = playerData.vipChips
			}
			user.WriteMsg(&msg.S2C_UpdatePlayerChips{
				Position: playerData.position,
				Chips:    chips,
			})
		}
	}
}

// 换桌(练习场、底分匹配场、红包匹配场才可以换桌)
func (roomm *LandlordRoom) changeTable(user *User) {
	userID := user.baseData.userData.UserID
	playerData := roomm.userIDPlayerDatas[userID]
	if playerData == nil {
		return
	}
	for i := 1; i < roomm.rule.MaxPlayers; i++ {
		otherUserID := roomm.positionUserIDs[(playerData.position+i)%roomm.rule.MaxPlayers]
		if otherUserID > 0 {
			otherPlayerData := roomm.userIDPlayerDatas[otherUserID]
			if otherPlayerData != nil {
				user.WriteMsg(&msg.S2C_StandUp{Position: otherPlayerData.position})
			}
		}
	}
	broadcast(&msg.S2C_StandUp{
		Position: playerData.position,
	}, roomm.positionUserIDs, -1)

	broadcast(&msg.S2C_ExitRoom{
		Error:    msg.S2C_ExitRoom_OK,
		Position: playerData.position,
	}, roomm.positionUserIDs, playerData.position)

	roomm.StandUp(user, playerData.position)               // 站起
	delete(userIDRooms, userID)                            // 退出
	delete(roomm.loginIPs, user.baseData.userData.LoginIP) // 删除玩家登录IP
	if roomm.empty() {                                     // 玩家为空，解散房间
		log.Debug("userID: %v 退出房间，房间解散", userID)
		delete(roomNumberRooms, roomm.number)
	} else {
		if roomm.ownerUserID == userID { // 转移房主
			for _, userID := range roomm.positionUserIDs {
				roomm.ownerUserID = userID
				break
			}
			log.Debug("userID: %v 退出房间, 转移房主 :%v", userID, roomm.ownerUserID)
			user.baseData.ownerUserID = roomm.ownerUserID
		} else {
			log.Debug("userID: %v 退出房间", userID)
		}
	}
	switch roomm.rule.RoomType {
	case roomPractice:
		for _, r := range roomNumberRooms {
			room := r.(*LandlordRoom)
			if room.rule.RoomType == roomPractice && !room.full() && room.ownerUserID != user.baseData.ownerUserID {
				user.enterRoom(r)
				return
			}
		}
	case roomBaseScoreMatching:
		for _, r := range roomNumberRooms {
			room := r.(*LandlordRoom)
			if !room.loginIPs[user.baseData.userData.LoginIP] && room.rule.RoomType == roomBaseScoreMatching && room.rule.BaseScore == roomm.rule.BaseScore && !room.full() && room.ownerUserID != user.baseData.ownerUserID {
				if !room.playTogether(user) {
					user.enterRoom(r)
					return
				}
			}
		}
	case roomRedPacketMatching:
		if !checkRedPacketMatchingTime() {
			user.WriteMsg(&msg.S2C_EnterRoom{Error: msg.S2C_EnterRoom_NotRightNow})
			return
		}
		for _, r := range roomNumberRooms {
			room := r.(*LandlordRoom)
			if !room.loginIPs[user.baseData.userData.LoginIP] && room.rule.RoomType == roomRedPacketMatching && room.rule.RedPacketType == roomm.rule.RedPacketType && !room.full() && room.ownerUserID != user.baseData.ownerUserID {
				if !room.playTogether(user) {
					user.enterRoom(r)
					return
				}
			}
		}
	}
	user.createLandlordRoom(roomm.rule)
}

func (room *LandlordRoom) updateTaskTicket(userID int) {
	playerData := room.userIDPlayerDatas[userID]
	for _, task := range playerData.user.baseData.taskIDTaskDatas {
		if TaskList[task.TaskID].Type == taskRedPacket && task.Progress < TaskList[task.TaskID].Total {
			upsertTaskTicket(bson.M{"userid": userID, "finish": false, "taskid": task.TaskID},
				bson.M{"$inc": bson.M{"ticket_" + strconv.Itoa(int(room.rule.Tickets)): 1}})
		}
	}
}

package internal

import (
	"common"
	"conf"
	"game/poker"
	"math/rand"
	"msg"
	"sort"
	"time"

	"github.com/name5566/leaf/log"
)

// 叫地主
func (room *LandlordRoom) bid(userID int) {
	playerData := room.userIDPlayerDatas[userID]
	playerData.state = landlordActionBid

	broadcast(&msg.S2C_ActionLandlordBid{
		Position:  playerData.position,
		Countdown: conf.GetCfgTimeout().LandlordBid,
	}, room.positionUserIDs, -1)

	playerData.actionTimestamp = time.Now().Unix()
	log.Debug("等待 userID %v 叫地主", userID)
	room.bidTimer = skeleton.AfterFunc((time.Duration(conf.GetCfgTimeout().LandlordBid+2))*time.Second, func() {
		log.Debug("userID %v 自动不叫", userID)
		room.doBid(userID, false)
	})
}

// 抢地主
func (room *LandlordRoom) grab(userID int) {
	playerData := room.userIDPlayerDatas[userID]
	playerData.state = landlordActionGrab

	broadcast(&msg.S2C_ActionLandlordGrab{
		Position:  playerData.position,
		Countdown: conf.GetCfgTimeout().LandlordGrab,
	}, room.positionUserIDs, -1)

	playerData.actionTimestamp = time.Now().Unix()
	log.Debug("等待 userID %v 抢地主", userID)
	room.grabTimer = skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordGrab+2)*time.Second, func() {
		log.Debug("userID %v 自动不抢", userID)
		room.doGrab(userID, false)
	})
}

// 确定地主
func (room *LandlordRoom) decideLandlord(userID int) {
	broadcast(&msg.S2C_ClearAction{}, room.positionUserIDs, -1)

	room.doTask(userID, 7)  // 累计当地主10次
	room.doTask(userID, 54) // 当地主15次

	room.landlordUserID = userID
	playerData := room.userIDPlayerDatas[room.landlordUserID]
	//新人任务 累计当地主5次 1002
	if room.rule.RoomType == roomBaseScoreMatching {
		playerData.user.updateRedPacketTask(1002)
		//初级任务 当地主10次 （1010）
		playerData.user.updateRedPacketTask(1010)
	}
	if room.rule.BaseScore == 3000 {
		//中级任务 普通场当地主8次 2006
		playerData.user.updateRedPacketTask(2006)
		room.doTask(userID, 44) // 普通场当地主8次
	}
	//高级任务 当地主20次 3001
	playerData.user.updateRedPacketTask(3001)
	for i := 1; i < room.rule.MaxPlayers; i++ {
		peasantUserID := room.positionUserIDs[(playerData.position+i)%room.rule.MaxPlayers]
		room.peasantUserIDs = append(room.peasantUserIDs, peasantUserID)
	}
	room.calculateLandlordMultiple()

	broadcast(&msg.S2C_DecideLandlord{
		Position: playerData.position,
	}, room.positionUserIDs, -1)
	// 最后三张
	room.lastThree = room.rests[:3]
	room.rests = []int{}
	sort.Sort(sort.Reverse(sort.IntSlice(room.lastThree)))
	log.Debug("三张: %v", poker.ToCardsString(room.lastThree))

	broadcast(&msg.S2C_UpdateLandlordLastThree{
		Cards: room.lastThree,
	}, room.positionUserIDs, -1)

	playerData.hands = append(playerData.hands, room.lastThree...)
	playerData.taskID51 = 0 // 单局打出2个顺子3次计数初始化
	sort.Sort(sort.Reverse(sort.IntSlice(playerData.hands)))

	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&msg.S2C_UpdatePokerHands{
			Position:      playerData.position,
			Hands:         playerData.hands,
			NumberOfHands: len(playerData.hands),
			ShowCards:     playerData.showCards,
		})
		user.WriteMsg(&msg.S2C_UpdateLandlordMultiple{
			Multiple: playerData.multiple,
		})
	}
	if playerData.showCards {
		broadcast(&msg.S2C_UpdatePokerHands{
			Position:      playerData.position,
			Hands:         playerData.hands,
			NumberOfHands: len(playerData.hands),
			ShowCards:     true,
		}, room.positionUserIDs, playerData.position)
	} else {
		broadcast(&msg.S2C_UpdatePokerHands{
			Position:      playerData.position,
			Hands:         []int{},
			NumberOfHands: len(playerData.hands),
		}, room.positionUserIDs, playerData.position)
	}
	switch room.rule.RoomType {
	case roomPractice:
		skeleton.AfterFunc(1*time.Second, func() {
			room.showCards()
		})
	case roomBaseScoreMatching, roomBaseScorePrivate:
		skeleton.AfterFunc(1*time.Second, func() {
			room.double()
		})
	case roomVIPPrivate, roomRedPacketMatching, roomRedPacketPrivate:
		broadcast(&msg.S2C_ClearAction{}, room.positionUserIDs, -1)
		room.discard(room.landlordUserID, poker.ActionLandlordDiscardMust)
	}
}

// 加倍
func (room *LandlordRoom) double() {
	actionTimestamp := time.Now().Unix()
	for _, userID := range room.positionUserIDs {
		playerData := room.userIDPlayerDatas[userID]
		playerData.state = landlordActionDouble
		playerData.actionTimestamp = actionTimestamp

		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&msg.S2C_ActionLandlordDouble{
				Countdown: conf.GetCfgTimeout().LandlordDouble,
			})
		}
	}
	log.Debug("等待所有人加倍")
	room.doubleTimer = skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordDouble+2)*time.Second, func() {
		for _, userID := range room.positionUserIDs {
			playerData := room.userIDPlayerDatas[userID]
			if playerData.state == landlordActionDouble {
				log.Debug("userID %v 自动不加倍", userID)
				room.doDouble(userID, false)
			}
		}
	})
}

// 明牌
func (room *LandlordRoom) showCards() {
	broadcast(&msg.S2C_ClearAction{}, room.positionUserIDs, -1)

	actionTimestamp := time.Now().Unix()
	switch room.rule.RoomType {
	case roomBaseScoreMatching:
		landlordPlayData := room.userIDPlayerDatas[room.landlordUserID]
		if landlordPlayData.showCards {
			room.discard(room.landlordUserID, poker.ActionLandlordDiscardMust)
			return
		}
		landlordPlayData.state = landlordActionShowCards
		landlordPlayData.actionTimestamp = actionTimestamp
		landlordPlayData.user.WriteMsg(&msg.S2C_ActionLandlordShowCards{
			Countdown: conf.GetCfgTimeout().LandlordShowCards,
		})

		log.Debug("等待地主明牌")
		room.showCardsTimer = skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordShowCards+2)*time.Second, func() {
			if landlordPlayData.state == landlordActionShowCards {
				log.Debug("地主: %v 自动不明牌", room.landlordUserID)
				room.doShowCards(room.landlordUserID, false)
			}
		})
	default:
		if len(room.getShowCardsUserIDs()) == room.rule.MaxPlayers {
			room.discard(room.landlordUserID, poker.ActionLandlordDiscardMust)
			return
		}
		for _, userID := range room.positionUserIDs {
			playerData := room.userIDPlayerDatas[userID]
			if !playerData.showCards {
				playerData.state = landlordActionShowCards
				playerData.actionTimestamp = actionTimestamp
				playerData.user.WriteMsg(&msg.S2C_ActionLandlordShowCards{
					Countdown: conf.GetCfgTimeout().LandlordShowCards,
				})
			}
		}
		log.Debug("等待其他人明牌")
		room.showCardsTimer = skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().LandlordShowCards+2)*time.Second, func() {
			for _, userID := range room.positionUserIDs {
				playerData := room.userIDPlayerDatas[userID]
				if playerData.state == landlordActionShowCards {
					log.Debug("userID %v 自动不明牌", userID)
					room.doShowCards(userID, false)
				}
			}
		})
	}
}

// 出牌
func (room *LandlordRoom) discard(userID int, actionDiscardType int) {
	playerData := room.userIDPlayerDatas[userID]
	playerData.state = landlordActionDiscard
	playerData.actionDiscardType = actionDiscardType

	broadcast(&msg.S2C_ActionLandlordDiscard{
		ActionDiscardType: poker.ActionLandlordDiscardAlternative,
		Position:          playerData.position,
		Countdown:         conf.GetCfgTimeout().LandlordDiscard,
	}, room.positionUserIDs, playerData.position)

	prevDiscards := []int{}
	countdown := conf.GetCfgTimeout().LandlordDiscard
	hint := make([][]int, 0)
	switch playerData.actionDiscardType {
	case poker.ActionLandlordDiscardNothing:
		if playerData.hosted {
			goto HOST
		}
		countdown = conf.GetCfgTimeout().LandlordDiscardNothing
	case poker.ActionLandlordDiscardAlternative:
		discarderPlayerData := room.userIDPlayerDatas[room.discarderUserID]
		prevDiscards = discarderPlayerData.discards[len(discarderPlayerData.discards)-1]
		if poker.CompareLandlordDiscard(playerData.hands, prevDiscards) {
			goto DISCARD_HANDS
		}
		if playerData.hosted {
			goto HOST
		}
		hint = poker.GetDiscardHint(prevDiscards, playerData.hands)
		log.Debug("提示出牌: %v", poker.ToMeldsString(hint))
	case poker.ActionLandlordDiscardMust:
		if len(playerData.hands) > 1 && poker.GetLandlordCardsType(playerData.hands[:2]) == poker.KingBomb && poker.GetLandlordCardsType(playerData.hands[2:]) > poker.Error {
			goto DISCARD_KINGBOMB
		}
		handsType := poker.GetLandlordCardsType(playerData.hands)
		if handsType != poker.Error && handsType != poker.FourDualsolo && handsType != poker.FourDualpair {
			goto DISCARD_HANDS
		}
		if playerData.hosted {
			goto HOST
		}
	}
	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&msg.S2C_ActionLandlordDiscard{
			ActionDiscardType: playerData.actionDiscardType,
			Position:          playerData.position,
			Countdown:         countdown,
			PrevDiscards:      prevDiscards,
			Hint:              hint,
		})
	}
	playerData.actionTimestamp = time.Now().Unix()
	log.Debug("等待 userID %v 出牌 动作: %v", userID, playerData.actionDiscardType)
	room.discardTimer = skeleton.AfterFunc(time.Duration(countdown+2)*time.Second, func() {
		switch playerData.actionDiscardType {
		case poker.ActionLandlordDiscardNothing:
			log.Debug("userID %v 自动不出", userID)
			room.doDiscard(userID, []int{})
		default:
			room.doSystemHost(userID, true)
		}
	})
	return
HOST: // 托管出牌
	skeleton.AfterFunc(1500*time.Millisecond, func() {
		room.doHostDiscard(userID)
	})
	return
DISCARD_HANDS: // 自动出手牌
	skeleton.AfterFunc(1500*time.Millisecond, func() {
		log.Debug("userID %v 自动出: %v", userID, poker.ToCardsString(playerData.hands))
		room.doDiscard(userID, playerData.hands)
	})
	return
DISCARD_KINGBOMB: // 自动出王炸
	skeleton.AfterFunc(1500*time.Millisecond, func() {
		log.Debug("userID %v 自动出王炸", userID)
		room.doDiscard(userID, []int{53, 52})
	})
}

func (room *LandlordRoom) doPrepare(userID int, showCards bool) {
	playerData := room.userIDPlayerDatas[userID]
	if room.rule.RoomType == roomVIPPrivate && playerData.vipChips == 0 {
		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&msg.S2C_SetVIPRoomChips{
				Error:    msg.S2C_SetVIPChips_ChipsUnset,
				Position: playerData.position,
			})
		}
		return
	}

	playerData.state = landlordReady
	playerData.showCards = showCards

	broadcast(&msg.S2C_Prepare{
		Position: playerData.position,
		Ready:    true,
	}, room.positionUserIDs, -1)

	if room.allReady() {
		room.state = roomGame
		skeleton.AfterFunc(1*time.Second, func() {
			room.StartGame()
		})
	}
}

func (room *LandlordRoom) doBid(userID int, bid bool) {
	playerData := room.userIDPlayerDatas[userID]
	if playerData.state != landlordActionBid {
		return
	}
	room.bidTimer.Stop()
	playerData.state = landlordWaiting

	broadcast(&msg.S2C_LandlordBid{
		Position: playerData.position,
		Bid:      bid,
	}, room.positionUserIDs, -1)

	dealerPlayerData := room.userIDPlayerDatas[room.dealerUserID]
	nextUserID := room.positionUserIDs[(playerData.position+1)%room.rule.MaxPlayers]
	lastPos := (dealerPlayerData.position + room.rule.MaxPlayers - 1) % room.rule.MaxPlayers
	if bid {
		room.bidUserID = userID
		if room.rule.BaseScore == 3000 {
			//中级任务 普通场叫地主6次 2007
			playerData.user.updateRedPacketTask(2007)
			room.doTask(userID, 45) // 普通场叫地主6次
		}
		if playerData.position == lastPos {
			skeleton.AfterFunc(1*time.Second, func() {
				room.decideLandlord(userID)
			})
		} else {
			room.grab(nextUserID)
		}
	} else {
		if playerData.position == lastPos {
			if room.shuffleTimes == 2 {
				skeleton.AfterFunc(1*time.Second, func() {
					landlordUserID := room.positionUserIDs[rand.Intn(room.rule.MaxPlayers)]
					room.decideLandlord(landlordUserID)
				})
				return
			}
			if len(room.getShowCardsUserIDs()) == 0 {
				skeleton.AfterFunc(1*time.Second, func() {
					room.StartGame()
				})
			} else {
				skeleton.AfterFunc(1*time.Second, func() {
					room.decideLandlord(room.getShowCardsUserIDs()[0])
				})
			}
		} else {
			room.bid(nextUserID)
		}
	}
}

func (room *LandlordRoom) doGrab(userID int, grab bool) {
	playerData := room.userIDPlayerDatas[userID]
	if playerData.state != landlordActionGrab {
		return
	}
	room.grabTimer.Stop()
	playerData.state = landlordWaiting

	broadcast(&msg.S2C_LandlordGrab{
		Position: playerData.position,
		Grab:     grab,
		Again:    userID == room.bidUserID,
	}, room.positionUserIDs, -1)

	dealerPlayerData := room.userIDPlayerDatas[room.dealerUserID]
	nextUserID := room.positionUserIDs[(playerData.position+1)%room.rule.MaxPlayers]
	lastPos := (dealerPlayerData.position + room.rule.MaxPlayers - 1) % room.rule.MaxPlayers
	if grab {
		room.doTask(userID, 8) // 累计抢地主10次
		//初级任务 抢地主10次 1011
		playerData.user.updateRedPacketTask(1011)
		room.calculateMultiple(-1, 2)
		room.grabUserIDs = append(room.grabUserIDs, userID)
		if userID == room.bidUserID {
			skeleton.AfterFunc(1*time.Second, func() {
				room.decideLandlord(userID)
			})
		} else {
			if playerData.position == lastPos {
				room.grab(room.bidUserID)
			} else {
				room.grab(nextUserID)
			}
		}
	} else {
		numberOfGrab := len(room.grabUserIDs)
		if userID == room.bidUserID {
			skeleton.AfterFunc(1*time.Second, func() {
				room.decideLandlord(room.grabUserIDs[numberOfGrab-1])
			})
		} else {
			if playerData.position == lastPos {
				if numberOfGrab == 0 {
					skeleton.AfterFunc(1*time.Second, func() {
						room.decideLandlord(room.bidUserID)
					})
				} else {
					room.grab(room.bidUserID)
				}
			} else {
				room.grab(nextUserID)
			}
		}
	}
}

func (room *LandlordRoom) doDouble(userID int, double bool) {
	playerData := room.userIDPlayerDatas[userID]
	if playerData.state != landlordActionDouble {
		return
	}
	playerData.state = landlordWaiting

	broadcast(&msg.S2C_LandlordDouble{
		Position: playerData.position,
		Double:   double,
	}, room.positionUserIDs, -1)

	if double {
		room.doTask(userID, 10) // 累计加倍底分10次
		//初级任务 菜鸟场加倍底分4次 1016
		if room.rule.BaseScore == 500 {
			playerData.user.updateRedPacketTask(1016)
		}
		if room.rule.BaseScore == 3000 {
			//中级任务 普通场加倍底分4次 2005
			playerData.user.updateRedPacketTask(2005)
			//高级任务 普通场加倍底分8次 3008
			playerData.user.updateRedPacketTask(3008)
			room.doTask(userID, 43) // 普通场加倍8次
		}

		room.calculateMultiple(userID, 2)
	}
	if room.allWaiting() {
		room.doubleTimer.Stop()
		skeleton.AfterFunc(1*time.Second, func() {
			room.showCards()
		})
	}
}

func (room *LandlordRoom) doShowCards(userID int, showCards bool) {
	playerData := room.userIDPlayerDatas[userID]
	if playerData.state != landlordActionShowCards {
		return
	}
	playerData.state = landlordWaiting
	playerData.showCards = showCards

	if playerData.showCards {
		room.calculateMultiple(userID, 2)

		broadcast(&msg.S2C_LandlordShowCards{
			Position: playerData.position,
		}, room.positionUserIDs, -1)

		broadcast(&msg.S2C_UpdatePokerHands{
			Position:      playerData.position,
			Hands:         playerData.hands,
			NumberOfHands: len(playerData.hands),
			ShowCards:     true,
		}, room.positionUserIDs, playerData.position)
		//初级任务 累计明牌开始10次 1012
		playerData.user.updateRedPacketTask(1012)
		//高级任务  普通场明牌开始10次 3005
		if room.rule.BaseScore == 3000 {
			playerData.user.updateRedPacketTask(3005)
		}
	}
	if room.allWaiting() {
		room.showCardsTimer.Stop()
		broadcast(&msg.S2C_ClearAction{}, room.positionUserIDs, -1)
		room.discard(room.landlordUserID, poker.ActionLandlordDiscardMust)
	}
}

// cards 长度为零时表示不出
func (room *LandlordRoom) doDiscard(userID int, cards []int) {
	playerData := room.userIDPlayerDatas[userID]
	if playerData.state != landlordActionDiscard {
		return
	}
	cards = poker.ReSortLandlordCards(cards)
	cardsLen := len(cards)
	cardsType := poker.GetLandlordCardsType(cards)
	contain := common.Contain(playerData.hands, cards)

	var prevDiscards []int
	if room.discarderUserID > 0 && room.discarderUserID != userID {
		discarderPlayerData := room.userIDPlayerDatas[room.discarderUserID]
		prevDiscards = discarderPlayerData.discards[len(discarderPlayerData.discards)-1]
	}
	if cardsLen == 0 && playerData.actionDiscardType == poker.ActionLandlordDiscardMust ||
		cardsLen > 0 && playerData.actionDiscardType == poker.ActionLandlordDiscardNothing ||
		cardsLen > 0 && !contain || cardsLen > 0 && cardsType == poker.Error ||
		cardsLen > 0 && playerData.actionDiscardType == poker.ActionLandlordDiscardAlternative && !poker.CompareLandlordDiscard(cards, prevDiscards) {
		if user, ok := userIDUsers[userID]; ok {
			after := int(time.Now().Unix() - playerData.actionTimestamp)
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
		return
	}
	if room.discardTimer != nil {
		room.discardTimer.Stop()
		room.discardTimer = nil
	}
	playerData.state = landlordWaiting

	broadcast(&msg.S2C_LandlordDiscard{
		Position: playerData.position,
		Cards:    cards,
	}, room.positionUserIDs, -1)

	nextUserID := room.positionUserIDs[(playerData.position+1)%room.rule.MaxPlayers]
	if cardsLen == 0 {
		log.Debug("userID %v 不出", userID)
		if room.discarderUserID == nextUserID {
			room.discard(nextUserID, poker.ActionLandlordDiscardMust)
		} else {
			nextUserPlayerData := room.userIDPlayerDatas[nextUserID]
			if poker.CompareLandlordHands(prevDiscards, nextUserPlayerData.hands) {
				room.discard(nextUserID, poker.ActionLandlordDiscardNothing)
			} else {
				room.discard(nextUserID, poker.ActionLandlordDiscardAlternative)
			}
		}
		return
	}
	switch cardsType {
	case poker.AirplaneChain, poker.TrioSoloAirplane, poker.TrioPairChain:
		room.doTask(userID, 1)  // 累计打出3个飞机
		room.doTask(userID, 23) // 累计打出3个飞机，奖励3000金币
		//初级任务 累计打出3个飞机 1006
		playerData.user.updateRedPacketTask(1006)
		if room.rule.BaseScore == 3000 {
			room.doTask(userID, 38) // 普通场打出1个飞机
			room.doTask(userID, 55) // 普通场打出2个飞机
		}
	case poker.PairSisters:
		room.doTask(userID, 2)  // 累计打出5个连对
		room.doTask(userID, 33) // 累计打出6个连对
		//初级任务 累计打出5个连对 1007
		playerData.user.updateRedPacketTask(1007)
		//中级任务 累计打出10个连对 2002
		playerData.user.updateRedPacketTask(2002)
		if room.rule.BaseScore == 3000 {
			room.doTask(userID, 39) // 普通场打出2个连对
			room.doTask(userID, 56) // 普通场打出4个连对
		}
	case poker.KingBomb:
		room.doTask(userID, 3) // 累计打出3个炸弹
		//初级任务 累计打出3个炸弹 1008
		playerData.user.updateRedPacketTask(1008)
		//初级任务 累计打出3个王炸 1009
		playerData.user.updateRedPacketTask(1009)
		//中级任务 累计打出5个王炸
		playerData.user.updateRedPacketTask(2000)
		room.doTask(userID, 4)  // 累计打出3个王炸
		room.doTask(userID, 24) // 累计打出2个王炸，奖励3000金币
		room.doTask(userID, 25) // 累计打出3个炸弹，奖励3000金币

		room.doTask(userID, 30) // 打出4个炸弹
		room.doTask(userID, 31) // 打出4个王炸
		room.doTask(userID, 34) // 打出5个炸弹
		room.doTask(userID, 46) // 打出6个炸弹
		room.doTask(userID, 47) // 打出5个王炸
		//中级任务 单局打出2个炸弹
		//playerData.user.updateRedPacketTask(2001)
		playerData.taskID2001++
		room.doTask(userID, 32) // 单局打出2个炸弹
		if room.rule.BaseScore == 3000 {
			room.doTask(userID, 40) // 普通场打出2个炸弹
			room.doTask(userID, 41) // 普通场打出2个王炸
			room.doTask(userID, 57) // 普通场打出3个炸弹
		}

		room.calculateMultiple(-1, 2)
	case poker.Bomb:
		//初级任务 累计打出3个炸弹 1008
		playerData.user.updateRedPacketTask(1008)
		room.doTask(userID, 3)  // 累计打出3个炸弹
		room.doTask(userID, 25) // 累计打出3个炸弹，奖励3000金币
		playerData.taskID2001++
		room.doTask(userID, 30) // 打出4个炸弹
		room.doTask(userID, 32) // 单局打出2个炸弹
		room.doTask(userID, 34) // 打出5个炸弹
		room.doTask(userID, 46) // 打出6个炸弹

		if room.rule.BaseScore == 3000 {
			room.doTask(userID, 40) // 普通场打出2个炸弹
			room.doTask(userID, 57) // 普通场打出3个炸弹
		}
		room.calculateMultiple(-1, 2)
	case poker.SoloChain:
		room.doTask(userID, 36) // 打出10个顺子
		playerData.taskID51++   // 单局打出2个顺子3次

		if room.rule.BaseScore == 3000 {
			//中级任务 普通场打出6个顺子 2004
			playerData.user.updateRedPacketTask(2004)
			room.doTask(userID, 42) // 普通场打出6个顺子
			room.doTask(userID, 58) // 普通场打出10个顺子
		}
	case poker.TrioPair:
		room.doTask(userID, 52) // 打出8次三带二
		//中级任务 打出8次三带二 2009
		playerData.user.updateRedPacketTask(2009)

		if room.rule.BaseScore == 3000 {
			room.doTask(userID, 63) // 普通场打出8次三带二
			//高级任务 普通场打出8次三带二 3002
			playerData.user.updateRedPacketTask(3002)
		}
	case poker.FourDualsolo:
		//中级任务 打出三次四带二 2010
		playerData.user.updateRedPacketTask(2010)
		room.doTask(userID, 53) // 打出3次四带二
	}
	room.discarderUserID = userID
	room.discards = append(room.discards, cards...)
	playerData.discards = append(playerData.discards, cards)
	playerData.hands = common.Remove(playerData.hands, cards)
	log.Debug("userID %v, 出牌: %v, 剩余: %v", userID, poker.ToCardsString(cards), poker.ToCardsString(playerData.hands))
	if playerData.showCards {
		broadcast(&msg.S2C_UpdatePokerHands{
			Position:      playerData.position,
			Hands:         playerData.hands,
			NumberOfHands: len(playerData.hands),
			ShowCards:     true,
		}, room.positionUserIDs, -1)
	} else {
		if user, ok := userIDUsers[userID]; ok {
			user.WriteMsg(&msg.S2C_UpdatePokerHands{
				Position:      playerData.position,
				Hands:         playerData.hands,
				NumberOfHands: len(playerData.hands),
			})
		}
		broadcast(&msg.S2C_UpdatePokerHands{
			Position:      playerData.position,
			Hands:         []int{},
			NumberOfHands: len(playerData.hands),
		}, room.positionUserIDs, playerData.position)
	}
	if len(playerData.hands) == 0 {
		room.winnerUserIDs = append(room.winnerUserIDs, userID)
		skeleton.AfterFunc(1*time.Second, func() {
			room.EndGame()
		})
		return
	}
	if room.discarderUserID == nextUserID {
		room.discard(nextUserID, poker.ActionLandlordDiscardMust)
	} else {
		nextUserPlayerData := room.userIDPlayerDatas[nextUserID]
		if poker.CompareLandlordHands(cards, nextUserPlayerData.hands) {
			room.discard(nextUserID, poker.ActionLandlordDiscardNothing)
		} else {
			room.discard(nextUserID, poker.ActionLandlordDiscardAlternative)
		}
	}
}

// 托管出牌
func (room *LandlordRoom) doHostDiscard(userID int) {
	playerData := room.userIDPlayerDatas[userID]
	if playerData.state != landlordActionDiscard {
		return
	}
	switch playerData.actionDiscardType {
	case poker.ActionLandlordDiscardNothing:
		room.doDiscard(userID, []int{})
		return
	case poker.ActionLandlordDiscardAlternative:
		discarderPlayerData := room.userIDPlayerDatas[room.discarderUserID]
		prevDiscards := discarderPlayerData.discards[len(discarderPlayerData.discards)-1]
		hint := poker.GetDiscardHint(prevDiscards, playerData.hands)
		if len(hint) == 0 {
			room.doDiscard(userID, []int{})
			return
		}
		hintType := poker.GetLandlordCardsType(hint[0])
		if len(playerData.discards) == 0 { // 还没有出过牌
			switch hintType {
			case poker.KingBomb, poker.Bomb:
				room.doDiscard(userID, []int{})
			}
		}
		if playerData.user.isRobot() {
			if common.InArray(room.peasantUserIDs, playerData.user.baseData.userData.UserID) &&
				common.InArray(room.peasantUserIDs, discarderPlayerData.user.baseData.userData.UserID) {
				switch hintType {
				case poker.Solo:
					if poker.CardValue(hint[0][0]) < 12 {
						room.doDiscard(userID, hint[0])
						return
					}
				case poker.Pair:
					if poker.CardValue(hint[0][0]) < 11 {
						room.doDiscard(userID, hint[0])
						return
					}
				}
				room.doDiscard(userID, []int{})
				return
			}
		}
		log.Debug("userID %v 托管出牌: %v", userID, poker.ToCardsString(hint[0]))
		room.doDiscard(userID, hint[0])
		return
	case poker.ActionLandlordDiscardMust:
		analyzer := new(poker.LandlordAnalyzer)
		minCards := analyzer.GetMinDiscards(playerData.hands)
		log.Debug("userID %v 托管出牌: %v", userID, poker.ToCardsString(minCards))
		room.doDiscard(userID, minCards)
		return
	}
}

func (room *LandlordRoom) doSystemHost(userID int, host bool) {
	playerData := room.userIDPlayerDatas[userID]
	if playerData.hosted == host {
		return
	}
	playerData.hosted = host
	// 托管不让别人知道
	playerData.user.WriteMsg(&msg.S2C_SystemHost{
		Position: playerData.position,
		Host:     host,
	})
	//broadcast(&msg.S2C_SystemHost{
	//	Position: playerData.position,
	//	Host:     host,
	//}, room.positionUserIDs, -1)

	if host {
		room.doHostDiscard(userID)
	}
}

// 重置任务
func (room *LandlordRoom) resetTask(userID int, taskID int) {
	if room.rule.RoomType == roomBaseScoreMatching {
		if user, ok := userIDUsers[userID]; ok {
			if task, ok := user.baseData.taskIDTaskDatas[taskID]; ok {
				if task.Progress < TaskList[taskID].Total {
					task.Progress = 0
					user.WriteMsg(&msg.S2C_UpdateTaskProgress{
						TaskID:   task.TaskID,
						Progress: task.Progress,
					})
				}
			}
		} else {
			resetTaskProgress(userID, taskID)
		}
	}
}

// 做任务
func (room *LandlordRoom) doTask(userID int, taskID int) {
	if room.rule.RoomType == roomBaseScoreMatching || taskID == 28 || taskID == 29 {
		if user, ok := userIDUsers[userID]; ok {
			user.doTask(taskID)
		} else {
			addTaskProgress(userID, taskID)
		}
	}
}

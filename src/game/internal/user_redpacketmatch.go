package internal

import (
	"common"
	"conf"
	"msg"
	"time"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

// 计算红包比赛在线人数
func calculateRedPacketMatchOnlineNumber(redPacketType int) {
	switch redPacketType {
	case 1: // 红包匹配场
		redPacketMatchOnlineNumber[0]++
	case 10: // 红包匹配场
		redPacketMatchOnlineNumber[1]++
	case 100: // 红包私人房
		redPacketMatchOnlineNumber[2]++
	case 999: // 红包私人房
		redPacketMatchOnlineNumber[3]++
	}
}

// 发送红包比赛在线人数
func (user *User) sendRedPacketMatchOnlineNumber() {
	//一元红包场
	if time.Now().Hour() < conf.GetOneRedpacketInfo().Start || time.Now().Hour() > conf.GetOneRedpacketInfo().End {
		redPacketMatchOnlineNumber[0] = 0
	}
	//十元红包场
	if time.Now().Hour() < conf.GetTenRedpacketInfo().Start || time.Now().Hour() > conf.GetTenRedpacketInfo().End {
		redPacketMatchOnlineNumber[1] = 0
	}
	if time.Now().Hour() < conf.GetHundredRedpacketInfo().Start || time.Now().Hour() > conf.GetHundredRedpacketInfo().End {
		redPacketMatchOnlineNumber[2] = 0
	}
	//十元红包场
	redPacketMatchOnlineNumber[3] = 0
	user.WriteMsg(&msg.S2C_UpdateRedPacketMatchOnlineNumber{
		Numbers: redPacketMatchOnlineNumber,
	})
}

// 发送未领取的红包比赛奖励数量
func (user *User) sendUntakenRedPacketMatchPrizeNumber() {
	count := 0
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		count, _ = db.DB(DB).C("redpacketmatchresult").
			Find(bson.M{"userid": user.baseData.userData.UserID, "redpacket": bson.M{"$gt": 0}, "taken": false}).Count()
	}, func() {
		user.WriteMsg(&msg.S2C_UpdateUntakenRedPacketMatchPrizeNumber{
			Number: count,
		})
	})
}

func (user *User) sendRedPacketMatchRecord(pageNumber int, pageSize int) {
	resultData := new(RedPacketMatchResultData)
	var items []msg.RedPacketMatchRecordItem
	count := 0
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		count, _ = db.DB(DB).C("redpacketmatchresult").Find(bson.M{"userid": user.baseData.userData.UserID, "redpacket": bson.M{"$gt": 0}}).Count()

		iter := db.DB(DB).C("redpacketmatchresult").Find(bson.M{"userid": user.baseData.userData.UserID, "redpacket": bson.M{"$gt": 0}}).
			Sort("-createdat").Skip((pageNumber - 1) * pageSize).Limit(pageSize).Iter()
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
		for iter.Next(&resultData) {
			items = append(items, msg.RedPacketMatchRecordItem{
				ID:            resultData.ID,
				RedPacketType: resultData.RedPacketType,
				RedPacket:     resultData.RedPacket,
				Taken:         resultData.Taken,
				Date:          time.Unix(resultData.CreatedAt, 0).Format("2006/01/02 15:04:05"),
				CardCode:      resultData.CardCode,
			})
		}
	}, func() {
		user.WriteMsg(&msg.S2C_RedPacketMatchRecord{
			Items:      items,
			Total:      count,
			PageNumber: pageNumber,
			PageSize:   pageSize,
		})
	})
}

func (user *User) takeRedPacketMatchPrize(id bson.ObjectId) {
	if user.baseData.userData.CircleID < 1 {
		user.WriteMsg(&msg.S2C_TakeRedPacketMatchPrize{
			Error: msg.S2C_TakeRedPacketMatchPrize_CircleIDInvalid,
		})
		user.requestCircleID()
		return
	}
	userID := user.baseData.userData.UserID

	resultData := new(RedPacketMatchResultData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		err := db.DB(DB).C("redpacketmatchresult").FindId(id).One(&resultData)
		if err != nil {
			resultData = nil
			log.Debug("find redpacketmatchresult: %v error: %v", id, err)
			return
		}
	}, func() {
		if resultData == nil {
			user.WriteMsg(&msg.S2C_TakeRedPacketMatchPrize{
				Error: msg.S2C_TakeRedPacketMatchPrize_IDInvalid,
				ID:    id,
			})
			return
		}
		if resultData.UserID != user.baseData.userData.UserID || resultData.RedPacket <= 0 {
			user.WriteMsg(&msg.S2C_TakeRedPacketMatchPrize{
				Error: msg.S2C_TakeRedPacketMatchPrize_NotYetWon,
				ID:    id,
			})
			return
		}
		if resultData.Taken {
			user.WriteMsg(&msg.S2C_TakeRedPacketMatchPrize{
				Error:     msg.S2C_TakeRedPacketMatchPrize_TakeRepeated,
				ID:        id,
				RedPacket: resultData.RedPacket,
			})
			return
		}
		if resultData.Handling {
			return
		}
		updateRedPacketMatchResultData(id, bson.M{"$set": bson.M{"handling": true}}, func() {
			/*
				// 请求生成一个圈圈红包
				desc := strconv.Itoa(resultData.RedPacketType) + "元红包比赛奖励"
				user.requestCircleRedPacket(resultData.RedPacket, desc, func() {
					takeRedPacketMatchPrizeSuccess(userID, id, resultData.RedPacket)
				}, func() {
					takeRedPacketMatchPrizeFail(userID, id)
				})
			*/
			takeRedPacketMatchPrizeSuccess(userID, id, resultData.RedPacket)
		})
	})
}

func takeRedPacketMatchPrizeSuccess(userID int, id bson.ObjectId, redPacket float64) {
	var cb func()
	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&msg.S2C_TakeRedPacketMatchPrize{
			Error:     msg.S2C_TakeRedPacketMatchPrize_OK,
			ID:        id,
			RedPacket: common.Round(redPacket, 2),
		})
		cb = func() {
			if theUser, ok := userIDUsers[userID]; ok {
				theUser.sendUntakenRedPacketMatchPrizeNumber()
			}
		}
	} else {
		cb = nil
	}
	updateRedPacketMatchResultData(id, bson.M{"$set": bson.M{"taken": true, "handling": false, "updatedat": time.Now().Unix()}}, cb)
}

func takeRedPacketMatchPrizeFail(userID int, id bson.ObjectId) {
	if user, ok := userIDUsers[userID]; ok {
		user.WriteMsg(&msg.S2C_TakeRedPacketMatchPrize{
			Error: msg.S2C_TakeRedPacketMatchPrize_Error,
			ID:    id,
		})
	}
	updateRedPacketMatchResultData(id, bson.M{"$set": bson.M{"handling": false}}, nil)
}

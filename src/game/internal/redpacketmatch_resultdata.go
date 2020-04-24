package internal

import (
	"encoding/json"
	"game/circle"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

type RedPacketMatchResultData struct {
	ID            bson.ObjectId `bson:"_id"`
	UserID        int
	RedPacketType int     // 红包种类(元): 1、5、10、50
	RedPacket     float64 // 红包奖励
	Taken         bool    // 奖励是否被领取
	Handling      bool    // 处理中
	CreatedAt     int64
	UpdatedAt     int64
	CardCode      string //红包码
}

func saveRedPacketMatchResultData(resultData *RedPacketMatchResultData) {
	temp := &struct {
		UserID        int
		RedPacketType int     // 红包种类(元): 1、5、10、50
		RedPacket     float64 // 红包奖励
		Taken         bool    // 是否领取
		CreatedAt     int64
		CardCode      string //  红包码
	}{}
	temp.UserID = resultData.UserID
	temp.RedPacketType = resultData.RedPacketType
	temp.RedPacket = resultData.RedPacket
	temp.Taken = resultData.Taken
	temp.CreatedAt = resultData.CreatedAt
	temp.CardCode = ""
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		if temp.RedPacket > 0 {
			temp1 := &struct {
				Code string
				Data string
			}{}
			r := new(circle.RedPacketCodeInfo)
			r.Sum = float64(temp.RedPacket)
			param, _ := json.Marshal(r)
			json.Unmarshal(circle.DoRequestRepacketCode(string(param)), temp1)
			log.Release("玩家用户Id:%v请求%v红包码:%v", temp.UserID, temp.RedPacket, temp1.Data)
			temp.CardCode = temp1.Data
		}
		err := db.DB(DB).C("redpacketmatchresult").Insert(temp)
		if err != nil {
			log.Error("insert redpacketmatchresult data error: %v", err)
		}
	}, nil)
}

func updateRedPacketMatchResultData(id bson.ObjectId, update interface{}, cb func()) {
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		_, err := db.DB(DB).C("redpacketmatchresult").UpsertId(id, update)
		if err != nil {
			log.Error("upsert redpacketmatchresult %v data error: %v", id, err)
		}
	}, func() {
		if cb != nil {
			cb()
		}
	})
}

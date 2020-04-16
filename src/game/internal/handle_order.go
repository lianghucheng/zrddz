package internal

import (
	"msg"
	"time"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

// 验证用户是否存在，存在则存储订单信息
func startWXPayOrder(outTradeNo string, accountID, totalFee int, cb func()) {
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		userData := new(UserData)
		err := db.DB(DB).C("users").Find(bson.M{"accountid": accountID}).One(userData)
		if err != nil {
			log.Debug("find accountID: %v error: %v", accountID, err)
			return
		}
		temp := &struct {
			UserID     int
			OutTradeNo string
			Success    bool
			TotalFee   int
			CreatedAt  int64
		}{
			UserID:     userData.UserID,
			OutTradeNo: outTradeNo,
			TotalFee:   totalFee,
			CreatedAt:  time.Now().Unix(),
		}
		_, err = db.DB(DB).C("wxpayresult").Upsert(bson.M{"outtradeno": outTradeNo}, bson.M{"$set": temp})
		if err != nil {
			log.Debug("upsert userID: %v error: %v", userData.UserID, err)
		}
	}, func() {
		if cb != nil {
			cb()
		}
	})
}

func finishWXPayOrder(outTradeNo string, totalFee int, valid bool) {
	temp := &struct {
		UserID     int
		OutTradeNo string
		Success    bool
		TotalFee   int
		Valid      bool
		UpdatedAt  int64
	}{}
	userData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		err := db.DB(DB).C("wxpayresult").Find(bson.M{"outtradeno": outTradeNo, "success": false}).One(temp)
		if err != nil {
			temp = nil
			log.Debug("find out_trade_no: %v error: %v", outTradeNo, err)
			return
		}
		if temp.TotalFee == totalFee {
			temp.Success = true
			temp.Valid = valid
			temp.UpdatedAt = time.Now().Unix()
			err = db.DB(DB).C("wxpayresult").Update(bson.M{"outtradeno": temp.OutTradeNo, "success": false}, bson.M{"$set": temp})
			if err != nil {
				log.Debug("update out_trade_no: %v error: %v", temp.OutTradeNo, err)
				temp = nil
			}
			if err := db.DB(DB).C("users").Find(bson.M{"_id": temp.UserID}).One(userData); err != nil {
				log.Release("read users: error: %v", err)
				userData = nil
			}
		} else {
			temp = nil
		}
	}, func() {
		if temp == nil {
			return
		}
		if userData != nil {
			userData.rebate(float64(temp.TotalFee) / 100.0)
			userData.countRecharge(float64(temp.TotalFee) / 100.0)
		}
		addChips := int64(temp.TotalFee) * 100
		switch temp.TotalFee {
		case 100:
			addChips = 8800
		case 600:
			addChips = 55600
		case 1200:
			addChips = 116000
		case 5000: // ￥50
			addChips = 550000
		case 10000: // ￥100
			addChips = 1236000
		}
		if user, ok := userIDUsers[temp.UserID]; ok {
			user.doTask(11) // 购买任意数量金币
			user.doTask(22) // 购买任意数量金币，奖励2000金币
			//新人任务 购买任意数量金币 1004
			user.updateRedPacketTask(1004)
			//初级任务 购买任意数量金币 1013
			user.updateRedPacketTask(1013)
			user.WriteMsg(&msg.S2C_PayOK{
				Chips: addChips,
			})
			user.baseData.userData.Chips += addChips
			user.WriteMsg(&msg.S2C_UpdateUserChips{
				Chips: user.baseData.userData.Chips,
			})
			if user.isRobot() {
				upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"recharge": addChips}})
			}
		} else {
			updateUserData(temp.UserID, bson.M{"$inc": bson.M{"chips": addChips}})

			addTaskProgress(temp.UserID, 11) // 购买任意数量金币
			addTaskProgress(temp.UserID, 22) // 购买任意数量金币，奖励2000金币
		}
	})
}

func startAliPayOrder(outTradeNo string, accountID int, totalAmount float64, cb func()) {
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		userData := new(UserData)
		err := db.DB(DB).C("users").Find(bson.M{"accountid": accountID}).One(userData)
		if err != nil {
			log.Debug("find accountID: %v error: %v", accountID, err)
			return
		}
		temp := &struct {
			UserID      int
			OutTradeNo  string
			Success     bool
			TotalAmount float64
			CreatedAt   int64
		}{
			UserID:      userData.UserID,
			OutTradeNo:  outTradeNo,
			TotalAmount: totalAmount,
			CreatedAt:   time.Now().Unix(),
		}
		_, err = db.DB(DB).C("alipayresult").Upsert(bson.M{"outtradeno": outTradeNo}, bson.M{"$set": temp})
		if err != nil {
			log.Debug("upsert userID: %v error: %v", userData.UserID, err)
		}
	}, func() {
		if cb != nil {
			cb()
		}
	})
}

func finishAliPayOrder(outTradeNo string, totalAmount float64, valid bool) {
	temp := &struct {
		UserID      int
		OutTradeNo  string
		Success     bool
		TotalAmount float64
		Valid       bool
		UpdatedAt   int64
	}{}
	userData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		err := db.DB(DB).C("alipayresult").Find(bson.M{"outtradeno": outTradeNo, "success": false}).One(temp)
		if err != nil {
			temp = nil
			log.Debug("find out_trade_no: %v error: %v", outTradeNo, err)
			return
		}
		if temp.TotalAmount == totalAmount {
			temp.Success = true
			temp.Valid = valid
			temp.UpdatedAt = time.Now().Unix()
			err = db.DB(DB).C("alipayresult").Update(bson.M{"outtradeno": temp.OutTradeNo, "success": false}, bson.M{"$set": temp})
			if err != nil {
				log.Debug("update out_trade_no: %v error: %v", temp.OutTradeNo, err)
				temp = nil
			}
			if err := db.DB(DB).C("users").Find(bson.M{"_id": temp.UserID}).One(userData); err != nil {
				log.Release("read users: error: %v", err)
				userData = nil
			}
		} else {
			temp = nil
		}
	}, func() {
		if temp == nil {
			return
		}
		if userData != nil {
			userData.rebate(float64(temp.TotalAmount))
			userData.countRecharge(float64(temp.TotalAmount))
		}
		addChips := int64(temp.TotalAmount * 10000)
		switch temp.TotalAmount {
		case 1:
			addChips = 8800
		case 6:
			addChips = 55600
		case 12:
			addChips = 116000
		case 50: // ￥50
			addChips = 550000
		case 100: // ￥100
			addChips = 1236000
		}
		if user, ok := userIDUsers[temp.UserID]; ok {
			user.doTask(11) // 购买任意数量金币
			user.doTask(22) // 购买任意数量金币，奖励2000金币
			//新人任务 购买任意数量金币 1004
			user.updateRedPacketTask(1004)
			//初级任务 购买任意数量金币 1013
			user.updateRedPacketTask(1013)
			user.WriteMsg(&msg.S2C_PayOK{
				Chips: addChips,
			})
			user.baseData.userData.Chips += addChips
			user.WriteMsg(&msg.S2C_UpdateUserChips{
				Chips: user.baseData.userData.Chips,
			})
			if user.isRobot() {
				upsertRobotData(time.Now().Format("20060102"), bson.M{"$inc": bson.M{"recharge": addChips}})
			}
		} else {
			updateUserData(temp.UserID, bson.M{"$inc": bson.M{"chips": addChips}})

			addTaskProgress(temp.UserID, 11) // 购买任意数量金币
			addTaskProgress(temp.UserID, 22) // 购买任意数量金币，奖励2000金币
		}
	})
}

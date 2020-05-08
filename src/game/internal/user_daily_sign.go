package internal

import (
	"common"
	"conf"
	"msg"
	"time"
)

func (user *User) dailySign() {
	user.checkDailySign()

	if user.baseData.userData.DailySign {
		return
	}

	user.baseData.userData.DailySign = true
	addChips := conf.GetCfgDailySign()[user.baseData.userData.SignTimes].Chips
	user.baseData.userData.Chips += addChips
	user.baseData.userData.SignTimes++
	saveUserData(user.baseData.userData)

	user.WriteMsg(&msg.S2C_UpdateUserChips{
		Chips:user.baseData.userData.Chips,
	})
	user.WriteMsg(&msg.S2C_DailySign{
		Chips: addChips,
	})

	user.sendDailySignItems()
}

func (user *User) sendDailySignItems() {
	user.checkDailySign()

	dailySignItems := []msg.DailySignItems{}
	for i := 0; i < user.baseData.userData.SignTimes; i++ {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Chips:conf.GetCfgDailySign()[i].Chips,
			Status:msg.SignFinish,
		})
	}
	if !user.baseData.userData.DailySign {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Chips:conf.GetCfgDailySign()[user.baseData.userData.SignTimes].Chips,
			Status:msg.SignAccess,
		})
	} else {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Chips:conf.GetCfgDailySign()[user.baseData.userData.SignTimes].Chips,
			Status:msg.SignDeny,
		})
	}

	for i := user.baseData.userData.SignTimes + 1; i < 7; i++ {
		dailySignItems = append(dailySignItems, msg.DailySignItems{
			Chips:conf.GetCfgDailySign()[i].Chips,
			Status:msg.SignDeny,
		})
	}
	user.WriteMsg(&msg.S2C_DailySignItems{
		SignItems:dailySignItems,
		IsSign:user.baseData.userData.DailySign,
	})
}

func (user *User) checkDailySign() {
	if user.baseData.userData.DailySignDeadLine < time.Now().Unix() {
		user.baseData.userData.DailySignDeadLine = common.OneDay0ClockTimestamp(time.Now().Add(24*time.Hour))
		user.baseData.userData.DailySign = false
		if time.Now().Weekday() == time.Monday {
			user.baseData.userData.SignTimes = 0
		}
		saveUserData(user.baseData.userData)
	}
}
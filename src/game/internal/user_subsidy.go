package internal

import (
	"msg"
	"time"
)

func (user *User) offerSubsidy() bool {
	if user.isRobot() || user.baseData.userData.Chips >= 1000 {
		return false
	}
	nowTime := time.Now()
	todayMidnight := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, time.Local)
	if user.baseData.userData.SubsidizedAt >= todayMidnight.Unix() {
		return false
	}
	var subsidy int64 = 2000
	user.baseData.userData.Chips += subsidy
	user.WriteMsg(&msg.S2C_OfferSubsidy{Chips: subsidy})
	user.WriteMsg(&msg.S2C_UpdateUserChips{Chips: user.baseData.userData.Chips})
	user.baseData.userData.SubsidizedAt = time.Now().Unix()
	return true
}

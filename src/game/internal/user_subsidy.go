package internal

import (
	"conf"
	"msg"
	"time"
)

func (user *User) offerSubsidy() bool {
	if user.isRobot() || user.baseData.userData.Chips >= int64(conf.Server.SubsidyLine) {
		return false
	}
	nowTime := time.Now()
	todayMidnight := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, time.Local)
	if user.baseData.userData.SubsidizedAt >= todayMidnight.Unix() {
		return false
	}
	user.baseData.userData.Chips += int64(conf.Server.SubsidyChip)
	user.WriteMsg(&msg.S2C_OfferSubsidy{Chips: int64(conf.Server.SubsidyChip)})
	user.WriteMsg(&msg.S2C_UpdateUserChips{Chips: user.baseData.userData.Chips})
	user.baseData.userData.SubsidizedAt = time.Now().Unix()
	return true
}

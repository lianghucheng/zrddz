package internal

import (
	"conf"
	"msg"
	"time"
)

func (user *User) AskSubsidyChip() {
	if user.baseData.userData.SubsidyDeadLine < time.Now().Unix() {
		next := time.Now().Add(24 * time.Hour)
		user.baseData.userData.SubsidyDeadLine = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location()).Unix()
		user.baseData.userData.SubsidyTimes = 0
		saveUserData(user.baseData.userData)
	}

	if user.baseData.userData.Chips < int64(conf.Server.SubsidyLine) && user.baseData.userData.SubsidyTimes < 2 {
		user.WriteMsg(&msg.S2C_SubsidyChip{
			SubsidyTimes: user.baseData.userData.SubsidyTimes + int64(1),
			TotalTimes:   conf.Server.SubsidyTotal,
			Chip:         conf.Server.SubsidyChip,
		})
	} else if user.baseData.userData.SubsidyTimes >= 2 {
		user.WriteMsg(&msg.S2C_SubsidyChip{
			Error: msg.SubsidyMore,
		})
	}
}

func (user *User) TakenSubsidyChip(reply bool) {
	if reply {
		if user.baseData.userData.SubsidyDeadLine < time.Now().Unix() {
			next := time.Now().Add(24 * time.Hour)
			user.baseData.userData.SubsidyDeadLine = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location()).Unix()
			user.baseData.userData.SubsidyTimes = 0
		}

		if user.baseData.userData.SubsidyTimes < 2 && user.baseData.userData.Chips < int64(conf.Server.SubsidyLine) {
			user.baseData.userData.SubsidyTimes++
			user.baseData.userData.Chips += int64(conf.Server.SubsidyChip)
			user.WriteMsg(&msg.S2C_UpdateUserChips{
				Chips: user.baseData.userData.Chips,
			})
			saveUserData(user.baseData.userData)
		} else if user.baseData.userData.SubsidyTimes >= 2 {
			user.WriteMsg(&msg.S2C_SubsidyChip{
				Error: msg.SubsidyMore,
			})
		} else if user.baseData.userData.Chips >= int64(conf.Server.SubsidyLine) {
			user.WriteMsg(&msg.S2C_SubsidyChip{
				Error: msg.SubsidyNotLack,
			})
		}
	}
}

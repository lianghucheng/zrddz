package internal

import (
	"common"
	"conf"
	"msg"
	"strings"
	"time"

	"github.com/name5566/leaf/log"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (user *User) wechatLogin(info *msg.C2S_WeChatLogin) {
	userData := new(UserData)
	firstLogin := false
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		// load userData
		err := db.DB(DB).C("users").Find(bson.M{"unionid": info.UnionID}).One(userData)
		if err == nil {
			return
		}
		if err == mgo.ErrNotFound {
			firstLogin = true
		} else {
			log.Error("load unionid %v data error: %v", info.UnionID, err)
			userData = nil
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_InnerError})
			user.Close()
			return
		}
		// new
		err = userData.initValue(info.Channel)
		if err != nil {
			log.Error("load unionid %v data error: %v", info.UnionID, err)
			userData = nil
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_InnerError})
			user.Close()
			return
		}
	}, func() {
		if userData == nil || user.state == userLogout {
			return
		}
		if userData.Role == roleBlack {
			user.WriteMsg(&msg.S2C_Close{
				Error:        msg.S2C_Close_RoleBlack,
				WeChatNumber: landlordConfigData.WeChatNumber,
			})
			user.Close()
			return
		}
		anotherLogin := false
		if oldUser, ok := userIDUsers[userData.UserID]; ok {
			if oldUser.baseData.userData.Serial != info.Serial {
				anotherLogin = true
			}
			oldUser.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_LoginRepeated})
			oldUser.Close()
			log.Debug("userID: %v 重复登录", userData.UserID)
			if oldUser == user {
				return
			}
			user.baseData = oldUser.baseData
			userData = oldUser.baseData.userData
		}
		userIDUsers[userData.UserID] = user
		if common.OneDay0ClockTimestamp(time.Now()) > userData.UpdatedAt {
			WriteSougouActivityRecord()
		}
		userData.updateWeChatInfo(info)
		user.baseData.userData = userData
		user.onLogin(firstLogin, anotherLogin)
		if firstLogin {
			inviteTask(user.baseData.userData.UnionID)
			log.Debug("userID: %v WeChat首次登录 unionid: %v, 在线人数: %v", user.baseData.userData.UserID, user.baseData.userData.UnionID, len(userIDUsers))
		} else {
			log.Debug("userID: %v WeChat登录 unionid: %v, 在线人数: %v", user.baseData.userData.UserID, user.baseData.userData.UnionID, len(userIDUsers))
		}
	})
}

func (user *User) tokenLogin(token string) {
	userData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		err := db.DB(DB).C("users").Find(bson.M{"token": token, "expireat": bson.M{"$gt": time.Now().Unix()}}).One(userData)
		if err != nil {
			log.Debug("find token %v error: %v", token, err)
			userData = nil
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_TokenInvalid})
			user.Close()
		}
	}, func() {
		if userData == nil || user.state == userLogout {
			return
		}
		if userData.Role == roleBlack {
			user.WriteMsg(&msg.S2C_Close{
				Error:        msg.S2C_Close_RoleBlack,
				WeChatNumber: landlordConfigData.WeChatNumber,
			})
			user.Close()
			return
		}
		ip := strings.Split(user.RemoteAddr().String(), ":")[0]
		if oldUser, ok := userIDUsers[userData.UserID]; ok {
			log.Debug("userID: %v 已经登录 %v %v", userData.UserID, oldUser.baseData.userData.LoginIP, ip)
			if ip == oldUser.baseData.userData.LoginIP {
				oldUser.Close()
			} else {
				user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_IPChanged})
				user.Close()
				return
			}
			user.baseData = oldUser.baseData
			userData = oldUser.baseData.userData
		}
		userIDUsers[userData.UserID] = user
		user.baseData.userData = userData
		user.onLogin(false, false)
		log.Debug("userID: %v Token登录, 在线人数: %v", userData.UserID, len(userIDUsers))
	})
}

func (user *User) usernamePasswordLogin(username string, password string) {
	userData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		// load
		err := db.DB(DB).C("users").Find(bson.M{"username": username, "password": password}).One(userData)
		if err != nil {
			log.Error("用户名: %v, 密码不正确: %v", username, err)
			userData = nil
			user.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_UsernameInvalid})
			user.Close()
		}
	}, func() {
		if userData == nil || user.state == userLogout {
			return
		}
		if userData.Role == -1 {
			user.WriteMsg(&msg.S2C_Close{
				Error:        msg.S2C_Close_RoleBlack,
				WeChatNumber: landlordConfigData.WeChatNumber,
			})
			user.Close()
			return
		}
		if oldUser, ok := userIDUsers[userData.UserID]; ok {
			oldUser.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_LoginRepeated})
			oldUser.Close()
			log.Debug("userID: %v 重复登录", userData.UserID)
			if oldUser == user {
				return
			}
			user.baseData = oldUser.baseData
			userData = oldUser.baseData.userData
		}
		userIDUsers[userData.UserID] = user
		user.baseData.userData = userData
		user.onLogin(false, false)
		log.Debug("用户名: %v 密码登录", username)
	})
}

func (user *User) logout() {
	if user.heartbeatTimer != nil {
		user.heartbeatTimer.Stop()
	}
	if user.baseData == nil {
		return
	}
	if existUser, ok := userIDUsers[user.baseData.userData.UserID]; ok {
		if existUser == user {
			log.Debug("userID: %v 登出", user.baseData.userData.UserID)
			user.onLogout()
			delete(userIDUsers, user.baseData.userData.UserID)
			user.baseData.userData.Online = false
			saveUserData(user.baseData.userData)
		}
	}
}

func (user *User) onLogin(firstLogin bool, anotherLogin bool) {
	if !user.isRobot() {
		user.baseData.userData.LoginIP = strings.Split(user.RemoteAddr().String(), ":")[0]
		user.baseData.userData.Token = common.GetToken(32)
		user.baseData.userData.ExpireAt = time.Now().Add(2 * time.Hour).Unix()
	}
	if conf.Server.FamilyActivity {
		now := time.Now()
		if user.baseData.userData.CollectDeadLine < now.Unix() {
			next := now.Add(24 * time.Hour)
			//零点计算
			user.baseData.userData.CardCode = ""
			user.baseData.userData.Taken = false
			user.baseData.userData.CollectDeadLine = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location()).Unix()
			user.baseData.userData.PlayTimes = 0
			updateUserData(user.baseData.userData.UserID, bson.M{"$set": bson.M{
				"cardcode":        user.baseData.userData.CardCode,
				"taken":           user.baseData.userData.Taken,
				"collectdeadline": user.baseData.userData.CollectDeadLine,
				"playtimes":       user.baseData.userData.PlayTimes,
			},
			})
		}
	}
	user.baseData.userData.Online = true
	if firstLogin {
		saveUserData(user.baseData.userData)
	} else {
		updateUserData(user.baseData.userData.UserID, bson.M{"$set": bson.M{"token": user.baseData.userData.Token, "online": user.baseData.userData.Online}})
	}
	user.autoHeartbeat()
	user.WriteMsg(&msg.S2C_Login{
		AccountID:       user.baseData.userData.AccountID,
		Nickname:        user.baseData.userData.Nickname,
		Headimgurl:      user.baseData.userData.Headimgurl,
		Sex:             user.baseData.userData.Sex,
		Role:            user.baseData.userData.Role,
		Token:           user.baseData.userData.Token,
		AnotherLogin:    anotherLogin,
		AnotherRoom:     userIDRooms[user.baseData.userData.UserID] != nil,
		FirstLogin:      firstLogin,
		Radio:           landlordConfigData.Radio,
		WeChatNumber:    landlordConfigData.WeChatNumber,
		CardCode:        user.baseData.userData.CardCode,
		Taken:           user.baseData.userData.Taken,
		CardCodeDesc:    conf.GetCfgDDZ().CardCodeDesc,
		PlayTimes:       user.baseData.userData.PlayTimes,
		Total:           conf.GetCfgCard().PlayTimes,
		Parentid:        user.baseData.userData.ParentId,
		GivenChips:      conf.Server.Chips,      //绑定赠送的金币数量
		FirstLoginChips: conf.Server.FirstLogin, //首次登录赠送的金币
	})

	if conf.Server.FamilyActivity {
		user.WriteMsg(&msg.S2C_CardMa{
			Code:      user.baseData.userData.CardCode,
			Total:     conf.GetCfgCard().PlayTimes,
			PlayTimes: user.baseData.userData.PlayTimes,
			Completed: user.baseData.userData.CardCode != "",
		})
	}
	user.WriteMsg(&msg.S2C_CircleLink{
		Url: conf.GetCfgLink().CircleLink,
	})
	if user.baseData.userData.Level == 0 {
		user.baseData.userData.Level = 1
	}
	//红包任务
	user.sendRedpacketTask(user.baseData.userData.Level)
	//红包记录
	user.redpacketTaskRecord()
	//请求圈圈
	user.requestCircleID()
	user.ShareInfo()

	user.sendDailySignItems()
	/*
		user.sendTaskList(firstLogin, func() {
			user.WriteMsg(&msg.S2C_Login{
				AccountID:    user.baseData.userData.AccountID,
				Nickname:     user.baseData.userData.Nickname,
				Headimgurl:   user.baseData.userData.Headimgurl,
				Sex:          user.baseData.userData.Sex,
				Role:         user.baseData.userData.Role,
				Token:        user.baseData.userData.Token,
				AnotherLogin: anotherLogin,
				AnotherRoom:  userIDRooms[user.baseData.userData.UserID] != nil,
				FirstLogin:   firstLogin,
				Radio:        landlordConfigData.Radio,
				WeChatNumber: landlordConfigData.WeChatNumber,
				CardCode:     user.baseData.userData.CardCode,
				Taken:        user.baseData.userData.Taken,
				CardCodeDesc: conf.GetCfgDDZ().CardCodeDesc,
				PlayTimes:    user.baseData.userData.PlayTimes,
				Total:        conf.GetCfgCard().PlayTimes,
			})
			if conf.Server.FamilyActivity {
				user.WriteMsg(&msg.S2C_CardMa{
					Code:      user.baseData.userData.CardCode,
					Total:     conf.GetCfgCard().PlayTimes,
					PlayTimes: user.baseData.userData.PlayTimes,
					Completed: user.baseData.userData.CardCode != "",
				})
			}
			user.redpacketTaskRecord()
			user.requestCircleID()
			//user.requestCheckCircleUserBlack()
			user.doTask(1000)
			user.offerSubsidy()
		})

	*/
}

func (user *User) onLogout() {
	if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.exitRoom(r, false)
	}
	//user.saveTaskList()
}

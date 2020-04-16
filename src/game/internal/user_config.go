package internal

import (
	"common"
	"msg"
	"net/url"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

func (user *User) setLandlordAndroidVersion(version int) {
	if version <= landlordConfigData.AndroidVersion && user.baseData.userData.Role < roleRoot {
		log.Debug("设置的斗地主安卓版本: %v 无效", version)
		user.WriteMsg(&msg.S2C_SetLandlordConfig{
			Error:          msg.S2C_SetLandlordConfig_VersionInvalid,
			AndroidVersion: version,
		})
		return
	}
	landlordConfigData.AndroidVersion = version
	saveConfigData(landlordConfigData)
	user.WriteMsg(&msg.S2C_SetLandlordConfig{
		Error:          msg.S2C_SetLandlordConfig_OK,
		AndroidVersion: version,
	})
	log.Debug("userID %v 设置斗地主安卓新版本为: %v", user.baseData.userData.UserID, version)
}

func (user *User) setLandlordIOSVersion(version int) {
	if version <= landlordConfigData.IOSVersion && user.baseData.userData.Role < roleRoot {
		log.Debug("设置的斗地主iOS版本: %v 无效", version)
		user.WriteMsg(&msg.S2C_SetLandlordConfig{
			Error:      msg.S2C_SetLandlordConfig_VersionInvalid,
			IOSVersion: version,
		})
		return
	}
	landlordConfigData.IOSVersion = version
	saveConfigData(landlordConfigData)
	user.WriteMsg(&msg.S2C_SetLandlordConfig{
		Error:      msg.S2C_SetLandlordConfig_OK,
		IOSVersion: version,
	})
	log.Debug("userID %v 设置斗地主iOS新版本为: %v", user.baseData.userData.UserID, version)
}

func (user *User) setLandlordAndroidDownloadUrl(downloadUrl string) {
	surl, err := url.Parse(downloadUrl)
	if err == nil && surl.Scheme == "" || err != nil || downloadUrl == landlordConfigData.AndroidDownloadUrl {
		log.Debug("设置的斗地主安卓下载地址: %v 无效", downloadUrl)
		user.WriteMsg(&msg.S2C_SetLandlordConfig{
			Error:              msg.S2C_SetLandlordConfig_DownloadUrlInvalid,
			AndroidDownloadUrl: downloadUrl,
		})
		return
	}
	landlordConfigData.AndroidDownloadUrl = downloadUrl
	saveConfigData(landlordConfigData)
	user.WriteMsg(&msg.S2C_SetLandlordConfig{
		Error:              msg.S2C_SetLandlordConfig_OK,
		AndroidDownloadUrl: downloadUrl,
	})
	log.Debug("userID %v 设置斗地主安卓下载地址为: %v", user.baseData.userData.UserID, downloadUrl)
}

func (user *User) setLandlordIOSDownloadUrl(downloadUrl string) {
	surl, err := url.Parse(downloadUrl)
	if err == nil && surl.Scheme == "" || err != nil || downloadUrl == landlordConfigData.IOSDownloadUrl {
		log.Debug("设置的斗地主iOS下载地址: %v 无效", downloadUrl)
		user.WriteMsg(&msg.S2C_SetLandlordConfig{
			Error:          msg.S2C_SetLandlordConfig_DownloadUrlInvalid,
			IOSDownloadUrl: downloadUrl,
		})
		return
	}
	landlordConfigData.IOSDownloadUrl = downloadUrl
	saveConfigData(landlordConfigData)
	user.WriteMsg(&msg.S2C_SetLandlordConfig{
		Error:          msg.S2C_SetLandlordConfig_OK,
		IOSDownloadUrl: downloadUrl,
	})
	log.Debug("userID %v 设置斗地主iOS下载地址为: %v", user.baseData.userData.UserID, downloadUrl)
}

func (user *User) setLandlordNotice(notice string) {
	landlordConfigData.Notice = notice
	saveConfigData(landlordConfigData)
	user.WriteMsg(&msg.S2C_SetLandlordConfig{
		Error:  msg.S2C_SetLandlordConfig_OK,
		Notice: notice,
	})
	broadcastAll(&msg.S2C_UpdateNotice{Notice: notice})
	log.Debug("userID %v 设置斗地主公告成功", user.baseData.userData.UserID)
}

func (user *User) setLandlordRadio(radio string) {
	landlordConfigData.Radio = radio
	saveConfigData(landlordConfigData)
	user.WriteMsg(&msg.S2C_SetLandlordConfig{
		Error: msg.S2C_SetLandlordConfig_OK,
		Radio: radio,
	})
	broadcastAll(&msg.S2C_UpdateRadio{Radio: radio})
	log.Debug("userID %v 设置斗地主广播成功", user.baseData.userData.UserID)
}

func (user *User) setLandlordWeChatNumber(wechatNumber string) {
	if wechatNumber == landlordConfigData.WeChatNumber {
		log.Debug("设置的斗地主客服微信号: %v 无效", wechatNumber)
		user.WriteMsg(&msg.S2C_SetLandlordConfig{
			Error:        msg.S2C_SetLandlordConfig_WeChatNumberInvalid,
			WeChatNumber: wechatNumber,
		})
		return
	}
	landlordConfigData.WeChatNumber = wechatNumber
	saveConfigData(landlordConfigData)
	user.WriteMsg(&msg.S2C_SetLandlordConfig{
		Error:        msg.S2C_SetLandlordConfig_OK,
		WeChatNumber: wechatNumber,
	})
	log.Debug("userID %v 设置斗地主客服微信号为: %v", user.baseData.userData.UserID, wechatNumber)
}

func (user *User) setUsernamePassword(username string, password string) {
	userData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		// load
		db.DB(DB).C("users").
			Find(bson.M{"username": username}).One(userData)
	}, func() {
		if user.state == userLogout {
			return
		}
		switch userData.UserID {
		case 0:
			user.baseData.userData.Username = username
			user.baseData.userData.Password = password
			updateUserData(user.baseData.userData.UserID, bson.M{"$set": bson.M{"username": username, "password": password}})
			log.Debug("userID %v 设置用户名: %v, 密码: %v", user.baseData.userData.UserID, username, password)
			return
		case user.baseData.userData.UserID:
			if user.baseData.userData.Password == password {
				log.Debug("userID %v 新、旧密码不能相同", user.baseData.userData.UserID)
				return
			}
			user.baseData.userData.Password = password
			updateUserData(user.baseData.userData.UserID, bson.M{"$set": bson.M{"password": password}})
			log.Debug("userID %v 修改密码为: %v", user.baseData.userData.UserID, password)
			return
		}
		log.Debug("用户名: %v 已占用", username)
	})
}

func toRoleString(role int) string {
	switch role {
	case roleRoot:
		return "超管"
	case roleAdmin:
		return "管理员"
	case roleAgent:
		return "代理"
	case rolePlayer:
		return "玩家"
	case roleBlack:
		return "拉黑"
	case roleRobot:
		return "机器人"
	}
	return ""
}

func (user *User) setRole(accountID int, role int) {
	if accountID == 0 {
		log.Debug("账户ID为0")
		user.WriteMsg(&msg.S2C_SetUserRole{Error: msg.S2C_SetUserRole_AccountIDInvalid})
		return
	}
	if user.baseData.userData.AccountID == accountID {
		log.Debug("不能设置自己")
		user.WriteMsg(&msg.S2C_SetUserRole{Error: msg.S2C_SetUserRole_NotYourself})
		return
	}
	if common.Index([]int{roleRobot, roleBlack, rolePlayer, roleAdmin, roleRoot}, role) == -1 {
		log.Debug("角色 %v 无效", role)
		user.WriteMsg(&msg.S2C_SetUserRole{
			Error: msg.S2C_SetUserRole_RoleInvalid,
			Role:  role,
		})
		return
	}
	if user.baseData.userData.Role < role {
		log.Debug("userID: %v 没有权限", user.baseData.userData.UserID)
		user.WriteMsg(&msg.S2C_SetUserRole{
			Error: msg.S2C_SetUserRole_PermissionDenied,
			Role:  role,
		})
		return
	}
	otherUserData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		db.DB(DB).C("users").Find(bson.M{"accountid": accountID}).One(otherUserData)
	}, func() {
		if otherUserData.UserID == 0 {
			log.Debug("账户ID: %v 的用户不存在", accountID)
			user.WriteMsg(&msg.S2C_SetUserRole{
				Error: msg.S2C_SetUserRole_AccountIDInvalid,
			})
			return
		}
		if otherUser, ok := userIDUsers[otherUserData.UserID]; ok {
			if otherUser.baseData.userData.Role == role {
				log.Debug("账户ID: %v 已经是: %v", accountID, toRoleString(role))
				user.WriteMsg(&msg.S2C_SetUserRole{
					Error: msg.S2C_SetUserRole_SetRepeated,
					Role:  role,
				})
				return
			}
			if user.baseData.userData.Role > otherUser.baseData.userData.Role {
				otherUser.baseData.userData.Role = role
				user.WriteMsg(&msg.S2C_SetUserRole{
					Error: msg.S2C_SetUserRole_OK,
					Role:  role,
				})
				log.Debug("userID %v 设置账号ID: %v为 %v", user.baseData.userData.UserID, accountID, toRoleString(role))
				if otherUser.baseData.userData.Role == roleBlack {
					otherUser.Close()
				}
				return
			}
		} else {
			if otherUserData.Role == role {
				log.Debug("账户ID: %v 已经是: %v", accountID, toRoleString(role))
				user.WriteMsg(&msg.S2C_SetUserRole{
					Error: msg.S2C_SetUserRole_SetRepeated,
					Role:  role,
				})
				return
			}
			if user.baseData.userData.Role > otherUserData.Role {
				otherUserData.Role = role
				updateUserData(otherUserData.UserID, bson.M{"$set": bson.M{"role": otherUserData.Role}})
				user.WriteMsg(&msg.S2C_SetUserRole{
					Error: msg.S2C_SetUserRole_OK,
					Role:  role,
				})
				log.Debug("userID %v 设置账号ID: %v为 %v", user.baseData.userData.UserID, accountID, toRoleString(role))
				return
			}
		}
		log.Debug("userID: %v 权限不够", user.baseData.userData.UserID)
		user.WriteMsg(&msg.S2C_SetUserRole{Error: msg.S2C_SetUserRole_PermissionDenied})
	})
}

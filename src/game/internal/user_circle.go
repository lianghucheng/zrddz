package internal

import (
	"encoding/json"
	"game/circle"
	"msg"

	"github.com/name5566/leaf/log"
)

func (user *User) requestCircleID() {
	if user.isRobot() || user.baseData.userData.CircleID > 0 {
		return
	}
	temp := &struct {
		Code     string
		CircleID int `json:"data"`
	}{}
	skeleton.Go(func() {
		req := circle.NewCircleLoginRequest(&circle.CircleUserInfo{
			UnionID:    user.baseData.userData.UnionID,
			Nickname:   user.baseData.userData.Nickname,
			Headimgurl: user.baseData.userData.Headimgurl,
			Sex:        user.baseData.userData.Sex,
		})
		data := circle.DoRequest(req.GatewayUrl, req.Method, req.Params)
		// {"id":null,"code":"0","model":null,"message":"ok","data":174}
		// log.Debug("%s", data)
		err := json.Unmarshal(data, temp)
		if err != nil || temp.Code != "0" {
			temp = nil
			return
		}
	}, func() {
		if temp != nil {
			user.baseData.userData.CircleID = temp.CircleID
		}
	})
}

func (user *User) requestCircleLoginCode(successCb func(loginCode string), failCb func()) {
	if user.baseData.userData.CircleID < 1 {
		if failCb != nil {
			failCb()
		}
		return
	}
	temp := &struct {
		Code      string
		LoginCode string `json:"data"`
	}{}
	skeleton.Go(func() {
		req := circle.NewCircleAuthorize(user.baseData.userData.CircleID)
		data := circle.DoRequest(req.GatewayUrl, req.Method, req.Params)
		// {"id":null,"code":"0","model":null,"message":"ok","data":"1730c016-01c7-4e15-85b8-986f5d812dd9"}
		// log.Debug("%s", data)
		err := json.Unmarshal(data, temp)
		if err != nil || temp.Code != "0" {
			temp = nil
			return
		}
	}, func() {
		if temp == nil {
			if failCb != nil {
				failCb()
			}
		} else {
			if successCb != nil {
				successCb(temp.LoginCode)
			}
		}
	})
}

func (user *User) requestCheckCircleUserBlack() {
	if user.isRobot() || user.baseData.userData.CircleID < 0 {
		return
	}
	temp := &struct {
		Code  string
		Black bool `json:"data"`
	}{}
	skeleton.Go(func() {
		req := circle.NewCircleCheckBlackRequest(user.baseData.userData.CircleID)
		data := circle.DoRequest(req.GatewayUrl, req.Method, req.Params)
		// {"id":null,"code":"0","model":null,"message":"ok","data":false}
		// log.Debug("%s", data)
		json.Unmarshal(data, temp)
	}, func() {
		if temp.Code == "0" && temp.Black {
			user.baseData.userData.Role = roleBlack
			user.WriteMsg(&msg.S2C_Close{
				Error:        msg.S2C_Close_RoleBlack,
				WeChatNumber: landlordConfigData.WeChatNumber,
			})
			user.Close()
		}
	})
}

// 请求生成一个圈圈红包
func (user *User) requestCircleRedPacket(redPacket float64, desc string, successCb func(), failCb func()) {
	temp := &struct {
		Code string
		Data string
	}{}
	skeleton.Go(func() {
		req := circle.NewCircleCreateRedPacketRequest(&circle.RedPacketInfo{
			UserID: user.baseData.userData.CircleID,
			Sum:    redPacket,
			Desc:   desc,
		})
		data := circle.DoRequest(req.GatewayUrl, req.Method, req.Params)
		// {"id":null,"code":"0","model":null,"message":"ok","data":"SUCCESS"}
		// log.Debug("%s", data)
		err := json.Unmarshal(data, temp)
		if err != nil || temp.Code != "0" || temp.Data != "SUCCESS" {
			log.Error("%s", data)
			temp = nil
			return
		}
	}, func() {
		if temp == nil {
			if failCb != nil {
				failCb()
			}
		} else {
			if successCb != nil {
				successCb()
			}
		}
	})
}

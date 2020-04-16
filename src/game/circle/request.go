package circle

import (
	"conf"
	"encoding/json"
	"errors"
	"github.com/name5566/leaf/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type CircleRequest struct {
	GatewayUrl string
	Method     string
	Params     string
}

type CircleUserInfo struct {
	UnionID    string `json:"unionid"`
	Nickname   string `json:"nickname"`
	Headimgurl string `json:"headimgurl"`
	Sex        int    `json:"sex"`
	Language   string `json:"language"`
	City       string `json:"city"`
	Province   string `json:"province"`
	Country    string `json:"country"`
}

type RedPacketInfo struct {
	UserID int     `json:"userid"`
	Sum    float64 `json:"sum"`
	Desc   string  `json:"desc"`
}

func NewCircleLoginRequest(info *CircleUserInfo) *CircleRequest {
	data, _ := json.Marshal(info)
	req := new(CircleRequest)
	req.GatewayUrl = "http://api.user.youxibi.com/server.do"
	req.Method = "youxibi.user.server.third.create.wechat"
	req.Params = string(data)
	return req
}

func NewCircleCreateRedPacketRequest(info *RedPacketInfo) *CircleRequest {
	data, _ := json.Marshal(info)
	req := new(CircleRequest)
	req.GatewayUrl = "http://api.circle.shenzhouxing.com/server.do"
	req.Method = "shenzhouxing.circle.server.packet.create.normal"
	req.Params = string(data)
	return req
}

func NewCircleAuthorize(userID int) *CircleRequest {
	temp := &struct {
		UserID int `json:"userid"`
	}{
		UserID: userID,
	}
	data, _ := json.Marshal(temp)
	req := new(CircleRequest)
	req.GatewayUrl = "http://api.user.youxibi.com/server.do"
	req.Method = "youxibi.user.server.create.login.code"
	req.Params = string(data)
	return req
}

// 检测圈圈用户是否被拉黑
func NewCircleCheckBlackRequest(userID int) *CircleRequest {
	temp := &struct {
		UserID int    `json:"userid"`
		Code   string `json:"code"`
	}{
		UserID: userID,
		Code:   "DEFAULT",
	}
	data, _ := json.Marshal(temp)
	req := new(CircleRequest)
	req.GatewayUrl = "http://api.user.youxibi.com/server.do"
	req.Method = "youxibi.user.server.black.user.check"
	req.Params = string(data)
	return req
}
/*
红包码请求参数
*/
type RedPacketCodeInfo struct {
	Sum float64 `json:"sum"`
}
func DoRequestRepacketCode(params string) []byte {
	p := url.Values{}
	p.Add("device", "SERVER")
	p.Add("deviceId", "CZDDZ")
	p.Add("lang", "CN")
	p.Add("method", conf.GetCfgRedpacketCode().Method)
	p.Add("params", params)
	p.Add("partnerKey", conf.GetCfgRedpacketCode().PartnerKey)
	p.Add("secretKey", conf.GetCfgRedpacketCode().SecretKey)
	p.Add("sendTime", strconv.Itoa(int(time.Now().Unix())))
	p.Add("signType", "NORMAL")
	p.Add("versionCode", "1")
	p.Add("versionName", "1.0")
	p.Add("sign", generateSign(p))
	r, err := http.PostForm(conf.GetCfgRedpacketCode().Url, p)
	if err != nil {
		log.Debug("%v", err)
		return []byte{}
	}
	defer r.Body.Close()
	result, _ := ioutil.ReadAll(r.Body)
	return result
}
func GetRedPacketCode(Fee float64) (code string, err error) {
	//请求圈圈获取红包码
	temp := &struct {
		Code string
		Data string
	}{}
	r := new(RedPacketCodeInfo)
	r.Sum = Fee
	param, _ := json.Marshal(r)
	err = json.Unmarshal(DoRequestRepacketCode(string(param)), temp)
	if temp.Code != "0" {
		log.Error("请求圈圈红包错误")
		return "", errors.New("The request error. ")
	}
	return temp.Data, err
}

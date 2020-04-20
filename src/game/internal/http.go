package internal

import (
	"conf"
	"encoding/json"
	"fmt"
	"msg"
	"net/http"
	"strconv"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"

	//_ "net/http/pprof"
	"reflect"
)

func init() {
	go startHTTPServer()
	/*go func() {
		err := http.ListenAndServe(":8888", nil)
		if err != nil {
			log.Fatal("%v", err)
		}
	}()*/
}

func startHTTPServer() {
	mux := http.NewServeMux()
	mux.Handle("/czddz/android", http.HandlerFunc(handleCZDDZAndroid))
	mux.Handle("/czddz/ios", http.HandlerFunc(handleCZDDZIOS))
	mux.Handle("/alipay", http.HandlerFunc(handleAliPay))
	mux.Handle("/wxpay", http.HandlerFunc(handleWXPay))
	mux.Handle("/invite", http.HandlerFunc(handleInvite))
	mux.HandleFunc("/exit", handleDeal)
	mux.HandleFunc("/set/system", handleSystem)
	mux.HandleFunc("/unionid", handleUnionid)
	mux.HandleFunc("/collect", handleCollect)

	mux.HandleFunc("/fakeralipay", handleFakerAliPay)
	mux.HandleFunc("/fakerwxpay", handleFakerWXPay)
	mux.HandleFunc("/fakerrprecord", handleFakerRedPacketRecord)
	err := http.ListenAndServe(conf.Server.HTTPAddr, mux)
	if err != nil {
		log.Fatal("%v", err)
	}
}

func handleCZDDZAndroid(w http.ResponseWriter, req *http.Request) {
	m := map[string]interface{}{}
	m["version"] = landlordConfigData.AndroidVersion
	m["downloadurl"] = landlordConfigData.AndroidDownloadUrl
	m["guestlogin"] = landlordConfigData.AndroidGuestLogin
	m["enteraddress"] = landlordConfigData.EnterAddress
	m["online"] = len(userIDUsers)
	data, err := json.Marshal(m)
	if err != nil {
		log.Error("marshal message %v error: %v", reflect.TypeOf(m), err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*") // 解决跨域问题
	fmt.Fprintf(w, "%s", data)
}

func handleCZDDZIOS(w http.ResponseWriter, req *http.Request) {
	m := map[string]interface{}{}
	m["version"] = landlordConfigData.IOSVersion
	m["downloadurl"] = landlordConfigData.IOSDownloadUrl
	m["guestlogin"] = landlordConfigData.IOSGuestLogin
	m["enteraddress"] = landlordConfigData.EnterAddress
	m["online"] = len(userIDUsers)
	data, err := json.Marshal(m)
	if err != nil {
		log.Error("marshal message %v error: %v", reflect.TypeOf(m), err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "%s", data)
}

func handleDeal(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	accounitd, _ := strconv.Atoi(id)
	userData := new(UserData)
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	db.DB(DB).C("users").Find(bson.M{"accountid": accounitd}).One(userData)
	if userData.UserID != 0 {
		if r, ok := userIDRooms[userData.UserID]; ok {
			roomm := r.(*LandlordRoom)
			roomm.state = roomIdle
			for _, userID := range roomm.positionUserIDs {
				switch roomm.rule.RoomType {
				case roomBaseScoreMatching, roomRedPacketMatching, roomRedPacketPrivate:
					roomm.Leave(userID)
				}
			}
			w.Write([]byte("成功"))
			return
		}
		w.Write([]byte("玩家不在游戏中"))
	}
	w.Write([]byte("玩家不存在"))
	return
}
func handleSystem(w http.ResponseWriter, req *http.Request) {
	on := req.FormValue("on")
	systemOn = (on == "true")
	w.Write([]byte("成功"))
	return
}
func handleUnionid(w http.ResponseWriter, req *http.Request) {
	cardcode := req.FormValue("cardcode")
	userData := new(UserData)
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	m := map[string]interface{}{}
	err := db.DB(DB).C("users").Find(bson.M{"cardcode": cardcode}).One(userData)
	if err != nil {
		m["code"] = 1001
		m["msg"] = err.Error()
		data, _ := json.Marshal(m)
		w.Write(data)
		return
	}
	m["code"] = 1000
	m["msg"] = "成功"
	m["data"] = userData
	data, _ := json.Marshal(m)
	w.Write(data)
	return
}
func handleCollect(w http.ResponseWriter, req *http.Request) {
	cardcode := req.FormValue("cardcode")
	userData := new(UserData)
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	err := db.DB(DB).C("users").Find(bson.M{"cardcode": cardcode}).One(userData)
	if err != nil {
		return
	}
	db.DB(DB).C("users").Update(bson.M{"cardcode": cardcode}, bson.M{
		"$set": bson.M{"taken": true},
	})
	if existUser, ok := userIDUsers[userData.UserID]; ok {
		existUser.baseData.userData.Taken = true
		existUser.WriteMsg(&msg.C2S_CardCodeState{})
	}
}

func handleFakerRedPacketRecord(w http.ResponseWriter, req *http.Request){
	accountid := req.FormValue("accountid")
	aid, _ := strconv.Atoi(accountid)
	grantType := req.FormValue("granttype")
	desc := req.FormValue("desc")
	gt,_ := strconv.Atoi(grantType)
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	userdata:=new(UserData)
	err:=db.DB(DB).C("users").Find(bson.M{"accountid":aid}) .One(userdata)
	if err != nil {
		log.Error("%v",err)
		return
	}

	WriteRedPacketGrantRecord(userdata, gt, desc, 1.1)
}
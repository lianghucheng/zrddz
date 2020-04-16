package internal

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	"time"
)

const (
	invite_success = 0
	invite_fail    = 1
	invite_missing = 2
	invite_wrong   = 3
	invite_repeat  = 4
)

// 没有加唯一索引
func handleInvite(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		w.Header().Set("Access-Control-Allow-Origin", "*") // 解决跨域问题
		unionID := r.URL.Query().Get("unionid")
		inviteID := r.URL.Query().Get("inviteid")
		result := &struct {
			Code int
		}{
			Code: -1,
		}
		if unionID == "" || inviteID == "" {
			result.Code = invite_missing
			res, _ := json.Marshal(result)
			fmt.Fprintf(w, "%v", string(res))
			return
		}
		accountID, _ := strconv.Atoi(inviteID) // 邀请人账号
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		tempData := struct {
			InviteUserID int `bson:"_id"`
		}{}
		// 检查邀请账号是否存在
		err := db.DB(DB).C("users").Find(bson.M{"accountid": accountID}).One(&tempData)
		if err != nil {
			result.Code = invite_wrong
			res, _ := json.Marshal(result)
			fmt.Fprintf(w, "%v", string(res))
			return
		}
		// 检查是否已经被邀请
		err = db.DB(DB).C("invite_userid").Find(bson.M{"unionid": unionID}).One(&tempData)
		if err == nil {
			result.Code = invite_repeat
			res, _ := json.Marshal(result)
			fmt.Fprintf(w, "%v", string(res))
			return
		}
		// 创建邀请关联
		err = db.DB(DB).C("invite_userid").Insert(&struct {
			UnionID      string
			InviteUserID int
		}{
			UnionID:      unionID,
			InviteUserID: tempData.InviteUserID,
		})
		if err != nil {
			result.Code = invite_fail
			res, _ := json.Marshal(result)
			fmt.Fprintf(w, "%v", string(res))
			return
		}
		result.Code = invite_success
		res, _ := json.Marshal(result)
		fmt.Fprintf(w, "%v", string(res))
	}
}

func inviteTask(unionid string) {
	if time.Now().Unix() < time.Date(2018, 4, 28, 0, 0, 0, 0, time.Local).Unix() ||
		time.Now().Unix() > time.Date(2018, 5, 6, 0, 0, 0, 0, time.Local).Unix() {
		return
	}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		tempData := struct {
			InviteUserID int
		}{}
		err := db.DB(DB).C("invite_userid").Find(bson.M{"unionid": unionid}).One(&tempData)
		if err != nil {
			return
		}
		doActivityTask(tempData.InviteUserID, 1017)
	}, nil)
}

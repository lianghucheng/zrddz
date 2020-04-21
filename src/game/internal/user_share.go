package internal

import (
	"common"
	"conf"
	"encoding/json"
	"fmt"
	"game/circle"

	"github.com/globalsign/mgo"

	"msg"
	"strconv"
	"time"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

type ShareTaskData struct {
	TaskID   int
	UserID   int
	Total    int
	IsFinish bool
	EndTime  int64
}

func (user *User) shareTasksInfo() {
	tasks := conf.GetCfgShareTask()
	data := []*msg.ShareTasks{}
	for _, v := range tasks {
		temp := new(msg.ShareTasks)
		temp.Desc = v.Desc
		temp.FeeDes = v.Info

		data = append(data, temp)
	}
	user.WriteMsg(&msg.S2C_ShareTasksInfo{
		ShareTasks: data,
	})
}
func (user *User) BindSharer(accountID string) {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	account, _ := strconv.Atoi(accountID)
	parentUserdata := new(UserData)
	if err := db.DB(DB).C("users").Find(bson.M{"accountid": account}).One(parentUserdata); err != nil {
		log.Release("查找绑定人账号不合法:%v", err)
		user.WriteMsg(&msg.S2C_BindSharer{
			Error: msg.BindSharerAbnormal,
		})
		return
	}

	if user.isRobot() || parentUserdata.Role == roleRobot {
		user.WriteMsg(&msg.S2C_BindSharer{
			Error: msg.BindRobot,
		})
		return
	}

	//如果玩家已经绑定了
	if user.baseData.userData.ParentId != 0 {
		user.WriteMsg(&msg.S2C_BindSharer{
			Error: msg.BindDuplicate,
		})
		return
	}

	//不能是自己
	if account == user.baseData.userData.AccountID {
		user.WriteMsg(&msg.S2C_BindSharer{
			Error: msg.BindSelf,
		})
		return
	}

	//邀请人注册时间较晚，请绑定其他推荐码
	if user.baseData.userData.CreatedAt < parentUserdata.CreatedAt {
		user.WriteMsg(&msg.S2C_BindSharer{
			Error: msg.BindShareLate,
		})
		return
	}

	userAgent := new(UserAgent)
	query := bson.M{"accountid": user.baseData.userData.AccountID}

	if err := db.DB(DB).C("userAgents").Find(query).One(userAgent); err != nil {
		log.Error(err.Error())
	}
	//不能绑定在自己的下级上面
	for i := 0; i < len(userAgent.Agents); i++ {
		for _, value := range userAgent.Agents[0].Datas {
			if strconv.Itoa(int(value.AccountId)) == accountID {
				user.WriteMsg(&msg.S2C_BindSharer{
					Error: msg.BindNotToLevel,
				})
				return
			}
		}
	}

	//更新玩家绑定页面
	user.WriteMsg(&msg.S2C_BindSharer{
		Error:    msg.BindSharerOK,
		ParentId: int64(account),
	})
	user.baseData.userData.Chips += int64(conf.Server.Chips)
	user.WriteMsg(&msg.S2C_UpdateUserChips{
		Chips: user.baseData.userData.Chips,
	})
	//玩家绑定的上级是否完成新手任务(分享至好友并成为上级合伙人)
	//userData := new(UserData)
	//err := db.DB(DB).C("users").Find(bson.M{"accountid": account}).One(userData)
	//if err == nil {
	//	tasks := new(TaskList)
	//	err := db.DB(DB).C("userDoRedpakcetTask").Find(bson.M{"userid": userData.UserID}).One(tasks)
	//	if err == nil {
	//		for key, value := range tasks.Tasks {
	//			//表示进行的任务是分享至好友并成为上级合伙人
	//			if value.State == 1 && value.ID != 1000 {
	//				break
	//			}
	//			if value.State == 1 && value.ID == 1000 {
	//				tasks.Tasks[key].State = 2
	//				tasks.Tasks[key].PlayTimes++
	//				tasks.Tasks[key+1].PlayTimes = 0
	//				tasks.Tasks[key+1].StartTime = time.Now().Unix()
	//				tasks.Tasks[key+1].State = 1
	//				if v, ok := userIDUsers[userData.UserID]; ok {
	//					v.data.redPacketTaskList = tasks.Tasks
	//					v.WriteMsg(&msg.S2C_RedpacketTask{
	//						Tasks:          v.data.redPacketTaskList,
	//						Chips:          ChangeChips[v.data.userData.Level],
	//						FreeChangeTime: v.data.userData.FreeChangeTime,
	//					})
	//				}
	//				update := &struct {
	//					Tasks     []msg.RedPacketTask
	//					UpdatedAt int64
	//				}{
	//					Tasks:     tasks.Tasks,
	//					UpdatedAt: time.Now().Unix(),
	//				}
	//				db.DB(DB).C("userDoRedpakcetTask").Upsert(bson.M{"userid": userData.UserID}, bson.M{"$set": update})
	//			}
	//		}
	//	}
	//}
	user.baseData.userData.ParentId = int64(account)
	updateUserData(user.baseData.userData.UserID, bson.M{"$set": bson.M{"parentid": account}})

	userAgentData := new(Data)
	userAgentData.Createdat = time.Now().Unix()
	userAgentData.AccountId = int64(user.baseData.userData.AccountID)

	preId := user.baseData.userData.ParentId

	for i := 0; i <= len(userAgent.Agents); i++ {
		temppreID := preId
		for j := 0; j < 8-i; j++ {
			if i == 0 {
				temppreID = user.pollingBindSharer(i+j+1, temppreID, *userAgentData)
			} else {
				length := len(userAgent.Agents[i-1].Datas)
				for k := 0; k < length-1; k++ {
					user.pollingBindSharer(i+j+1, temppreID, userAgent.Agents[i-1].Datas[k])
				}
				temppreID = user.pollingBindSharer(i+j+1, temppreID, userAgent.Agents[i-1].Datas[length-1])
			}
			if temppreID <= 0 {
				break
			}
		}
	}
	if len(userAgent.Agents) > 0 {
		userAgent.ParentId = int64(account)
		db.DB(DB).C("userAgents").Upsert(bson.M{"accountid": user.baseData.userData.AccountID}, userAgent)
	}
}

func (user *User) pollingBindSharer(level int, parentID int64, userAgent Data) int64 {
	parentUserData := new(UserData)
	if err := parentUserData.readByAccountID(parentID); err != nil {
		log.Error(err.Error())
		return -1
	}
	userAgent.AllProfit = 0
	userAgent.Profit = 0
	userAgent.InsertAgent(level, parentID, int64(parentUserData.ParentId))

	return int64(parentUserData.ParentId)
}

func (userData *UserData) finishShareTask(shareTaskID int) {

	shareTaskData := new(ShareTaskData)
	err := shareTaskData.read(userData.UserID, shareTaskID)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if !shareTaskData.IsFinish {
		shareTaskData.EndTime = time.Now().Unix()
		shareTaskData.IsFinish = true
		shareTaskData.upsert(userData.UserID, shareTaskID)
	}
}

func (ctx *ShareTaskData) read(userid int, taskid int) error {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	err := db.DB(DB).C("sharetask").Find(bson.M{"userid": userid, "taskid": taskid}).One(ctx)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (ctx *ShareTaskData) upsert(userid int, taskid int) {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	if _, err := db.DB(DB).C("sharetask").Upsert(bson.M{"userid": userid, "taskid": taskid}, ctx); err != nil {
		log.Error(err.Error())
	}
}

type ShareTakenRecord struct {
	ID           int
	Fee          float64
	Level        int
	DirID        int
	CreatedAt    int64
	IsCopy       bool
	ExchangeCode string
}

type ShareTakenRecords struct {
	AccountID    int
	TakenRecords []msg.ShareTakenRecord
}

func (userData *UserData) rebate(fee float64) {
	if userData.ParentId == 0 {
		log.Release("玩家%v进行充值操作,其无绑定任何代理", userData.AccountID)
		return
	}
	grandParentID := userData.ParentId

	//找到他有几个父节点
	count := 0
	for i := 0; i < 8; i++ { //往上轮询
		count++
		grandParentID = userData.pollingGiveAchievement(0, 0, grandParentID, i+1)
		if grandParentID <= 0 {
			break
		}
	}

	grandParentID = userData.ParentId
	fee = common.Decimal(fee * conf.Server.RebateRate / float64(count))

	//更新上级的可领取收益
	for i := 0; i < count; i++ { //往上轮询
		grandParentID = userData.pollingGiveAchievement(fee, 0, grandParentID, i+1)
		if grandParentID <= 0 {
			break
		}
	}
}

func (userData *UserData) countRecharge(recharge float64) {
	grandParentID := userData.ParentId
	if userData.ParentId == 0 {
		log.Release("玩家%v进行返利操作,其无绑定任何代理", userData.AccountID)
		return
	}
	for i := 0; i < 8; i++ { //往上轮询
		grandParentID = userData.pollingGiveAchievement(0, recharge, grandParentID, i+1)
		if grandParentID <= 0 {
			break
		}
	}
}

func (userData *UserData) giveShareAward(fee float64, dirID int64, level int) string {
	parentUserData := new(UserData)
	if err := parentUserData.readByAccountID(dirID); err != nil {
		log.Error(err.Error())
		return ""
	}
	return parentUserData.shareAwardRecord(fee, level)
}

func (userData *UserData) pollingGiveAchievement(fee, recharge float64, parentID int64, level int) int64 {
	parentUserData := new(UserData)
	if err := parentUserData.readByAccountID(parentID); err != nil && err != mgo.ErrNotFound {
		log.Error(err.Error())
		return -1
	}

	if fee > 0 {
		WriteRedPacketGrantRecord(parentUserData, 2, fmt.Sprintf("%v级下级充值“%v”元，进行返利", level, fee), fee)
	}

	_, datas := shareAbleProfitAndDatas(level, parentID)
	data := findAgentData(datas, int64(userData.AccountID))
	if data == nil {
		log.Error("找不到代理")
		return -1
	}

	data.updateIncAgent(level, parentID, map[string]float64{"profit": fee, "allprofit": fee, "recharge": recharge})

	return parentUserData.ParentId
}

func findAgentData(datas []Data, accountid int64) (rt *Data) {
	for i := 0; i < len(datas); i++ {
		if datas[i].AccountId == accountid {
			return &datas[i]
		}
	}
	return
}

func (userData *UserData) shareAwardRecord(fee float64, level int) string {
	shareTakenRecord := new(ShareTakenRecord)
	preID := 0
	temp := new(ShareTakenRecords)
	if err := temp.read(int64(userData.AccountID)); err != nil && err != mgo.ErrNotFound {
		log.Error(err.Error())
	} else {
		if len(temp.TakenRecords) > 0 {
			preID = temp.TakenRecords[0].ID
		}
	}

	shareTakenRecord.CreatedAt = time.Now().Unix()
	shareTakenRecord.Fee = common.Decimal(fee)
	shareTakenRecord.Level = level
	shareTakenRecord.DirID = int(userData.AccountID)
	shareTakenRecord.ExchangeCode = getRedPacketCode(fee)
	shareTakenRecord.ID = preID + 1

	if err := shareTakenRecord.push(int64(userData.AccountID)); err != nil {
		log.Error(err.Error())
		return ""
	}
	return shareTakenRecord.ExchangeCode
}

func (user *User) shareAwardRecord(page, per int) {
	temp := new(ShareTakenRecords)
	temp.read(int64(user.baseData.userData.AccountID))
	//过滤掉已经过期的数据
	r := new(ShareTakenRecords)
	for i := 0; i < len(temp.TakenRecords); i++ {
		if temp.TakenRecords[i].CreatedAt+86400*3 > time.Now().Unix() {
			r.TakenRecords = append(r.TakenRecords, temp.TakenRecords[i])
		}
	}
	tr := r.TakenRecords

	stRecord := []*msg.ShareTakenRecord{}
	for i := (page - 1) * per; i < (page-1)*per+per; i++ {
		if i >= len(tr) {
			break
		}

		tempRecd := new(msg.ShareTakenRecord)
		tempRecd.CreatedAt = tr[i].CreatedAt
		tempRecd.ID = tr[i].ID
		tempRecd.Fee = tr[i].Fee
		tempRecd.Level = tr[i].Level
		tempRecd.DirID = tr[i].DirID
		tempRecd.IsCopy = tr[i].IsCopy
		tempRecd.ExchangeCode = tr[i].ExchangeCode

		stRecord = append(stRecord, tempRecd)
	}

	user.WriteMsg(&msg.S2C_ShareRecord{
		Page:              page,
		Per:               per,
		Total:             len(r.TakenRecords),
		ShareTakenRecords: stRecord,
	})
}

func (user *User) isExistAward() bool {
	str := new(ShareTakenRecords)
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	query := bson.M{"accountid": user.baseData.userData.AccountID, "takenrecords.iscopy": false}
	if err := db.DB(DB).C("sharetakenrecords").Find(query).One(str); err != nil {
		for i := 0; i < 8; i++ {
			ableprofit, _ := shareAbleProfitAndDatas(i+1, int64(user.baseData.userData.AccountID))
			if ableprofit > 0 {
				return true
			}
		}
		return false
	} else {
		return true
	}
}

func getRedPacketCode(Fee float64) (code string) {
	//请求圈圈获取红包码
	temp := &struct {
		Code string
		Data string
	}{}
	r := new(circle.RedPacketCodeInfo)
	r.Sum = Fee

	param, _ := json.Marshal(r)
	json.Unmarshal(circle.DoRequestRepacketCode(string(param)), temp)

	if temp.Code != "0" {
		log.Error("请求圈圈红包错误")
		return
	}
	return temp.Data
}

func (ctx *ShareTakenRecords) read(accountid int64) error {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	query := bson.M{"accountid": accountid}
	err := db.DB(DB).C("sharetakenrecords").Find(query).One(ctx)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (ctx *ShareTakenRecord) push(accountid int64) error {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	selector := bson.M{"accountid": accountid}
	update := bson.M{"$push": bson.M{"takenrecords": bson.M{"$each": []ShareTakenRecord{*ctx}, "$sort": bson.M{"createdat": -1}}}}
	_, err := db.DB(DB).C("sharetakenrecords").Upsert(selector, update)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (user *User) copyExchangeCode(shareRecordID int) {
	strecord := new(ShareTakenRecord)
	log.Release("%+v", *strecord)
	strecord.ID = shareRecordID
	strecord.IsCopy = true

	if err := strecord.update(user.baseData.userData.AccountID); err != nil {
		user.WriteMsg(&msg.S2C_CopyExchangeCode{
			Error: msg.CopyFail,
		})
		log.Error(err.Error())
		return
	}
	user.WriteMsg(&msg.S2C_CopyExchangeCode{
		Error:         msg.CopyOK,
		ShareRecordID: shareRecordID,
	})
}

func (ctx *ShareTakenRecord) update(accountid int) error {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	selector := bson.M{"accountid": accountid, "takenrecords.id": ctx.ID}
	update := bson.M{"$set": bson.M{"takenrecords.$.iscopy": ctx.IsCopy}}
	err := db.DB(DB).C("sharetakenrecords").Update(selector, update)
	return err
}

func (user *User) achievement(level int, page int, per int) {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	userAgent := new(UserAgent)
	err := db.DB(DB).C("userAgents").Find(bson.M{"accountid": user.baseData.userData.AccountID}).One(userAgent)
	if err != nil && err != mgo.ErrNotFound {
		log.Error(err.Error())
		return
	}
	achievements := []msg.Achievement{}
	total := 0
	if err == mgo.ErrNotFound {
		user.WriteMsg(&msg.S2C_Achievement{
			Page:         page,
			Per:          per,
			Total:        total,
			Achievements: achievements,
		})
		return
	}

	if level <= len(userAgent.Agents) {
		datas := userAgent.Agents[level-1].Datas
		total = len(datas)
		for i := (page - 1) * per; i < (page-1)*per+per; i++ {
			if i >= len(datas) {
				break
			}
			achievements = append(achievements, msg.Achievement{
				Createdat: datas[i].Createdat,
				AccountId: datas[i].AccountId,
				Recharge:  datas[i].Recharge,
				AllProfit: datas[i].AllProfit,
				Profit:    datas[i].Profit,
				Updatedat: datas[i].Updatedat,
			})
		}
	}

	user.WriteMsg(&msg.S2C_Achievement{
		Page:         page,
		Per:          per,
		Total:        total,
		Achievements: achievements,
	})
}

func (user *User) ableProfit(level int) {
	ableProfit, _ := shareAbleProfitAndDatas(level, int64(user.baseData.userData.AccountID))

	user.WriteMsg(&msg.S2C_AbleProfit{
		AbleProfit: ableProfit,
	})
}

func shareAbleProfitAndDatas(level int, accountid int64) (float64, []Data) {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	userAgent := new(UserAgent)

	err := db.DB(DB).C("userAgents").Find(bson.M{"accountid": accountid}).One(userAgent)
	if err != nil && err != mgo.ErrNotFound {
		log.Error(err.Error())
		return 0, []Data{}
	}
	if err == mgo.ErrNotFound {
		return 0, []Data{}
	}
	var ableProfit float64
	if level <= len(userAgent.Agents) {
		datas := userAgent.Agents[level-1].Datas
		for _, v := range datas {
			ableProfit += v.Profit
		}
		return ableProfit, userAgent.Agents[level-1].Datas
	} else {
		return ableProfit, []Data{}
	}
}

func (user *User) agentNumbersProfit() {
	numbersProfit := make([]msg.NumbersProfit, 8)
	for i := 0; i < len(numbersProfit); i++ {
		numbersProfit[i].Level = i + 1
	}

	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	userAgent := new(UserAgent)
	if err := db.DB(DB).C("userAgents").Find(bson.M{"accountid": user.baseData.userData.AccountID}).One(userAgent); err != nil && err != mgo.ErrNotFound {
		log.Error(err.Error())
	}

	for _, v := range userAgent.Agents {
		numbersProfit[v.Level-1].Level = v.Level
		numbersProfit[v.Level-1].Profit, _ = shareAbleProfitAndDatas(v.Level, int64(user.baseData.userData.AccountID))
		numbersProfit[v.Level-1].Number = len(v.Datas)
	}

	user.WriteMsg(&msg.S2C_AgentNumbersProfit{
		NumbersProfits: numbersProfit,
	})
}

func (user *User) receiveProfit(level int) {
	fee, datas := shareAbleProfitAndDatas(level, int64(user.baseData.userData.AccountID))
	if fee == 0 {
		user.WriteMsg(&msg.S2C_ReceiveShareProfit{
			Error: msg.ReceiveFail,
		})
		return
	}
	exchangeCode := user.baseData.userData.giveShareAward(fee, int64(user.baseData.userData.AccountID), level)

	for i := 0; i < len(datas); i++ {
		datas[i].updateIncAgent(level, int64(user.baseData.userData.AccountID), map[string]float64{"profit": -datas[i].Profit})
	}

	user.WriteMsg(&msg.S2C_ReceiveShareProfit{
		Error:        msg.ReceiveOK,
		Profit:       fee,
		ExchangeCode: exchangeCode,
	})
}

func (user *User) takenProfit() {
	shareTakenRecords := new(ShareTakenRecords)
	shareTakenRecords.read(int64(user.baseData.userData.AccountID))
	var takenProfit float64
	for _, v := range shareTakenRecords.TakenRecords {
		takenProfit += v.Fee
	}

	user.WriteMsg(&msg.S2C_TakenProfit{
		TakenProfit: takenProfit,
	})
}

package internal

import (
	"common"
	"conf"
	"encoding/json"
	"fmt"
	"game/circle"
	"math/rand"
	"msg"
	"reflect"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
)

func init() {
	handler(&msg.C2S_SetSystemOn{}, handleSetSystemOn)
	handler(&msg.C2S_SetUsernamePassword{}, handleSetUsernamePassword)
	handler(&msg.C2S_SetLandlordConfig{}, handleSetLandlordConfig)
	handler(&msg.C2S_TransferChips{}, handleTransferChips)
	handler(&msg.C2S_SetUserRole{}, handleSetUserRole)

	handler(&msg.C2S_Heartbeat{}, handleHeartbeat)
	//创建斗地主私人房
	handler(&msg.C2S_CreateLandlordRoom{}, handleCreateLandlordRoom)
	//加入斗地主房间
	handler(&msg.C2S_EnterRoom{}, handleEnterRoom)
	handler(&msg.C2S_GetUserChips{}, handleGetUserChips)
	handler(&msg.C2S_SetVIPRoomChips{}, handleSetVIPRoomChips)
	handler(&msg.C2S_GetAllPlayers{}, handleGetAllPlayers)
	handler(&msg.C2S_ExitRoom{}, handleExitRoom)
	handler(&msg.C2S_LandlordPrepare{}, handleLandlordPrepare)
	handler(&msg.C2S_LandlordMatching{}, handleLandlordMatching)
	handler(&msg.C2S_LandlordBid{}, handleLandlordBid)
	handler(&msg.C2S_LandlordGrab{}, handleLandlordGrab)
	handler(&msg.C2S_LandlordDouble{}, handleLandlordDouble)
	handler(&msg.C2S_LandlordShowCards{}, handleLandlordShowCards)
	handler(&msg.C2S_LandlordDiscard{}, handleLandlordDiscard)
	handler(&msg.C2S_GetMonthChipsRank{}, handleMonthChipsRank)
	handler(&msg.C2S_GetMonthChipsRankPos{}, handleMonthChipsRankPos)
	handler(&msg.C2S_GetMonthWinsRank{}, handleMonthWinsRank)
	handler(&msg.C2S_GetMonthWinsRankPos{}, handleMonthWinsRankPos)
	handler(&msg.C2S_CleanMonthRanks{}, handleCleanMonthRanks)
	handler(&msg.C2S_SystemHost{}, handleSystemHost)
	handler(&msg.C2S_ChangeTable{}, handleChangeTable)
	handler(&msg.C2S_GetRedPacketMatchRecord{}, handleGetRedPacketMatchRecord)
	handler(&msg.C2S_TakeRedPacketMatchPrize{}, handleTakeRedPacketMatchPrize)
	handler(&msg.C2S_DoTask{}, handleDoTask)
	handler(&msg.C2S_ChangeTask{}, handleChangeTask)
	handler(&msg.C2S_FreeChangeCountDown{}, handleFreeChangeCountDown)
	handler(&msg.C2S_TakeTaskPrize{}, handleTakeTaskPrize)
	handler(&msg.C2S_FakeWXPay{}, handleFakeWXPay)
	handler(&msg.C2S_FakeAliPay{}, handleFakeAliPay)
	handler(&msg.C2S_GetCircleLoginCode{}, handleGetCircleLoginCode)

	handler(&msg.C2S_SetRobotData{}, handleSetRobotData)
	handler(&msg.C2S_TakeTaskState{}, handleTaskState)
	handler(&msg.C2S_GetShareTaskInfo{}, handleShareTasks)
	handler(&msg.C2S_GetRedpacketTaskCode{}, handleRedpacktTaskCode)
	handler(&msg.C2S_SubsidyChip{}, handleSubsidyChip)
	handler(&msg.C2S_IsExistSubsidy{}, handleIsExistSubsidy)
	handler(&msg.C2S_ShareInfo{}, handleShareInfo)
}

func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handleSetSystemOn(args []interface{}) {
	m := args[0].(*msg.C2S_SetSystemOn)
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.user == nil {
		return
	}
	user := agentInfo.user
	if user.baseData.userData.Role < roleRoot {
		log.Debug("userID: %v 没有权限", user.baseData.userData.UserID)
		user.WriteMsg(&msg.S2C_SetSystemOn{Error: msg.S2C_SetSystemOn_PermissionDenied})
		return
	}
	systemOn = m.On
	user.WriteMsg(&msg.S2C_SetSystemOn{
		Error: msg.S2C_SetSystemOn_OK,
		On:    systemOn,
	})
	if !systemOn {
		clearToken() // 清除Token
	}
	if systemOn {
		log.Debug("userID: %v 系统开", user.baseData.userData.UserID)
	} else {
		log.Debug("userID: %v 系统关", user.baseData.userData.UserID)
	}
}

func handleHeartbeat(args []interface{}) {
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.user == nil {
		return
	}
	agentInfo.user.heartbeatStop = false
}

func handleSetUsernamePassword(args []interface{}) {
	m := args[0].(*msg.C2S_SetUsernamePassword)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if strings.TrimSpace(m.Username) == "" || strings.TrimSpace(m.Password) == "" {
		// 用户名或密码不能为空
		return
	}
	switch user.baseData.userData.Username {
	case "", m.Username:
		user.setUsernamePassword(m.Username, m.Password)
	default:
		log.Debug("userID: %v 用户名无需更改", user.baseData.userData.UserID)
	}
}

func handleSetLandlordConfig(args []interface{}) {
	m := args[0].(*msg.C2S_SetLandlordConfig)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if user.baseData.userData.Role < roleAdmin {
		log.Debug("userID: %v 没有权限", user.baseData.userData.UserID)
		user.WriteMsg(&msg.S2C_SetLandlordConfig{Error: msg.S2C_SetLandlordConfig_PermissionDenied})
		return
	}
	if m.AndroidVersion > 0 {
		user.setLandlordAndroidVersion(m.AndroidVersion)
	}
	if len(m.AndroidDownloadUrl) > 0 {
		user.setLandlordAndroidDownloadUrl(m.AndroidDownloadUrl)
	}
	if m.IOSVersion > 0 {
		user.setLandlordIOSVersion(m.IOSVersion)
	}
	if len(m.IOSDownloadUrl) > 0 {
		user.setLandlordIOSDownloadUrl(m.IOSDownloadUrl)
	}
	if len(m.Notice) > 0 {
		user.setLandlordNotice(m.Notice)
	}
	if len(m.Radio) > 0 {
		user.setLandlordRadio(m.Radio)
	}
	if len(m.WeChatNumber) > 0 {
		user.setLandlordWeChatNumber(m.WeChatNumber)
	}
}

func handleTransferChips(args []interface{}) {
	m := args[0].(*msg.C2S_TransferChips)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if !systemOn {
		user.Close()
		return
	}
	if m.AccountID == 0 {
		user.WriteMsg(&msg.S2C_TransferChips{Error: msg.S2C_TransferChips_AccountIDInvalid})
		return
	}
	if user.baseData.userData.AccountID == m.AccountID {
		user.WriteMsg(&msg.S2C_TransferChips{Error: msg.S2C_TransferChips_NotYourself})
		return
	}
	if m.Chips < 1 || m.Chips > user.baseData.userData.Chips {
		user.WriteMsg(&msg.S2C_TransferChips{Error: msg.S2C_TransferChips_ChipsInvalid})
		return
	}
	if user.baseData.userData.Role < roleAgent {
		user.WriteMsg(&msg.S2C_TransferChips{Error: msg.S2C_TransferChips_PermissionDenied})
		return
	}
	user.transferChips(m.AccountID, m.Chips)
}

func handleSetUserRole(args []interface{}) {
	m := args[0].(*msg.C2S_SetUserRole)
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.user == nil {
		return
	}
	agentInfo.user.setRole(m.AccountID, m.Role)
}

func handleCreateLandlordRoom(args []interface{}) {
	// 收到的 C2S_CreateLandlordRoom 消息
	m := args[0].(*msg.C2S_CreateLandlordRoom)
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if !systemOn {
		user.Close()
		return
	}
	if _, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_InOtherRoom})
		return
	}
	switch m.RoomType {
	//case roomVIPPrivate:
	//	user.createVIPPrivateRoom(m.MaxPlayers)
	//	return
	//case roomBaseScorePrivate:
	//	user.createBasePrivateRoom(m.BaseScore)
	//	return
	case roomRedPacketPrivate:
		user.createRedPacketPrivateRoom(m.RedPacketType)
		return
	default:
		user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_RuleError})
	}
}

func handleEnterRoom(args []interface{}) {
	// 收到的 C2S_EnterRoom 消息
	m := args[0].(*msg.C2S_EnterRoom)
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if !systemOn {
		user.Close()
		return
	}
	if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.enterRoom(r)
		return
	}
	if strings.TrimSpace(m.RoomNumber) == "" {
		user.WriteMsg(&msg.S2C_EnterRoom{Error: msg.S2C_EnterRoom_Unknown})
		return
	}
	if r, ok := roomNumberRooms[m.RoomNumber]; ok {
		user.enterRoom(r)
	} else {
		user.WriteMsg(&msg.S2C_EnterRoom{
			Error:      msg.S2C_EnterRoom_NotCreated,
			RoomNumber: m.RoomNumber,
		})
	}
}

func handleGetUserChips(args []interface{}) {
	m := args[0].(*msg.C2S_GetUserChips)
	_ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.WriteMsg(&msg.S2C_UpdateUserChips{
		Chips: user.baseData.userData.Chips,
	})
	user.sendRedPacketMatchOnlineNumber()
	user.sendUntakenRedPacketMatchPrizeNumber()
}

func handleGetAllPlayers(args []interface{}) {
	// 收到的 C2S_GetAllPlayers 消息
	_ = args[0].(*msg.C2S_GetAllPlayers)
	// 消息的发送者
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.user == nil {
		return
	}
	user := agentInfo.user
	if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.getAllPlayers(r)
	}
}

func handleExitRoom(args []interface{}) {
	// 收到的 C2S_ExitRoom 消息
	_ = args[0].(*msg.C2S_ExitRoom)
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.exitRoom(r, true)
	} else {
		user.WriteMsg(&msg.S2C_ExitRoom{
			Error:    msg.S2C_ExitRoom_OK,
			Position: -1,
		})
	}
}

func handleLandlordPrepare(args []interface{}) {
	m := args[0].(*msg.C2S_LandlordPrepare)
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.user == nil {
		return
	}
	user := agentInfo.user
	if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.doPrepare(r, m.ShowCards)
	}
}

func handleSetVIPRoomChips(args []interface{}) {
	m := args[0].(*msg.C2S_SetVIPRoomChips)
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if m.Chips < 10000 || m.Chips%10000 > 0 {
		user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_RuleError})
		return
	}
	if m.Chips > user.baseData.userData.Chips {
		user.WriteMsg(&msg.S2C_SetVIPRoomChips{Error: msg.S2C_SetVIPChips_LackOfChips})
		return
	}
	if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.setVIPRoomChips(r, m.Chips)
	}
}

func handleLandlordMatching(args []interface{}) {
	m := args[0].(*msg.C2S_LandlordMatching)
	// 消息的发送者
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if !systemOn {
		user.Close()
		return
	}
	if _, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		//user.enterRoom(r)
		user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_InOtherRoom})
		return
	}
	switch m.RoomType {
	case roomPractice:
		user.createOrEnterPracticeRoom()
		return
	case roomBaseScoreMatching:
		user.createOrEnterBaseScoreMatchingRoom(m.BaseScore)
		return
	case roomRedPacketMatching:
		user.createOrEnterRedPacketMatchingRoom(m.RedPacketType)
		return
	default:
		user.WriteMsg(&msg.S2C_CreateRoom{Error: msg.S2C_CreateRoom_RuleError})
	}
}

func handleLandlordBid(args []interface{}) {
	m := args[0].(*msg.C2S_LandlordBid)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.doBid(r, m.Bid)
	}
}

func handleLandlordGrab(args []interface{}) {
	m := args[0].(*msg.C2S_LandlordGrab)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.doGrab(r, m.Grab)
	}
}

func handleLandlordDouble(args []interface{}) {
	m := args[0].(*msg.C2S_LandlordDouble)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.doDouble(r, m.Double)
	}
}

func handleLandlordShowCards(args []interface{}) {
	m := args[0].(*msg.C2S_LandlordShowCards)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.doShowCards(r, m.ShowCards)
	}
}

func handleLandlordDiscard(args []interface{}) {
	m := args[0].(*msg.C2S_LandlordDiscard)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.doDiscard(r, m.Cards)
	}
}

func handleMonthChipsRank(args []interface{}) {
	m := args[0].(*msg.C2S_GetMonthChipsRank)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.getMonthChipsRank(m.PageNum)
}

func handleMonthChipsRankPos(args []interface{}) {
	_ = args[0].(*msg.C2S_GetMonthChipsRankPos)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.getMonthChipsRankPos()
}

func handleMonthWinsRank(args []interface{}) {
	m := args[0].(*msg.C2S_GetMonthWinsRank)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.getMonthWinsRank(m.PageNum)
}

func handleMonthWinsRankPos(args []interface{}) {
	m := args[0].(*msg.C2S_GetMonthWinsRankPos)
	_ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.getMonthWinsRankPos()
}

func handleCleanMonthRanks(args []interface{}) {
	m := args[0].(*msg.C2S_CleanMonthRanks)
	_ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if user.baseData.userData.Role < roleRoot {
		user.Close()
		return
	}
	user.dropMonthRank()
}

func handleSystemHost(args []interface{}) {
	m := args[0].(*msg.C2S_SystemHost)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.doSystemHost(r, m.Host)
	}
}

func handleChangeTable(args []interface{}) {
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.user == nil {
		return
	}
	user := agentInfo.user
	if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		user.doChangeTable(r)
	}
}

func handleGetRedPacketMatchRecord(args []interface{}) {
	m := args[0].(*msg.C2S_GetRedPacketMatchRecord)
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.user == nil {
		return
	}
	if m.PageNumber > 0 && m.PageSize == 10 {
		agentInfo.user.sendRedPacketMatchRecord(m.PageNumber, m.PageSize)
	}
}

func handleTakeRedPacketMatchPrize(args []interface{}) {
	m := args[0].(*msg.C2S_TakeRedPacketMatchPrize)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.takeRedPacketMatchPrize(m.ID)
}

func handleDoTask(args []interface{}) {
	m := args[0].(*msg.C2S_DoTask)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.doTask(m.TaskID)
}

func handleChangeTask(args []interface{}) {
	m := args[0].(*msg.C2S_ChangeTask)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.changeRedPacketTask(m)
}

func handleFreeChangeCountDown(args []interface{}) {
	//m := args[0].(*msg.C2S_FreeChangeCountDown)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.freeChangeCountDown()
}

func handleTakeTaskPrize(args []interface{}) {
	m := args[0].(*msg.C2S_TakeTaskPrize)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	if _, ok := user.baseData.taskIDTaskDatas[m.TaskID]; ok {
		user.takeTaskPrize(m.TaskID)
	} else {
		user.WriteMsg(&msg.S2C_TakeTaskPrize{Error: msg.S2C_TakeTaskPrize_TaskIDInvalid})
	}
}

func handleFakeWXPay(args []interface{}) {
	m := args[0].(*msg.C2S_FakeWXPay)
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.user == nil {
		return
	}
	agentInfo.user.FakeWXPay(m.TotalFee)
}

func handleFakeAliPay(args []interface{}) {
	m := args[0].(*msg.C2S_FakeAliPay)
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.user == nil {
		return
	}
	agentInfo.user.FakeAliPay(m.TotalAmount)
}

func handleGetCircleLoginCode(args []interface{}) {
	m := args[0].(*msg.C2S_GetCircleLoginCode)
	_ = m
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	userID := user.baseData.userData.UserID
	user.requestCircleLoginCode(func(loginCode string) {
		if theUser, ok := userIDUsers[userID]; ok {
			theUser.WriteMsg(&msg.S2C_UpdateCircleLoginCode{
				Error:     msg.S2C_UpdateCircleLoginCode_OK,
				LoginCode: loginCode,
			})
		}
	}, func() {
		if theUser, ok := userIDUsers[userID]; ok {
			if theUser == user {
				theUser.WriteMsg(&msg.S2C_UpdateCircleLoginCode{
					Error: msg.S2C_UpdateCircleLoginCode_Error,
				})
			}
		}
	})
}

func handleSetRobotData(args []interface{}) {
	m := args[0].(*msg.C2S_SetRobotData)
	a := args[1].(gate.Agent)

	agentInfo := a.UserData().(*AgentInfo)
	if agentInfo == nil || agentInfo.user == nil {
		return
	}
	user := agentInfo.user
	if user.isRobot() {
		if m.Chips > 0 {
			user.setRobotChips(m.Chips)
		}
	} else {
		user.baseData.userData.Chips = rand.Int63n(14000) + 6000
		user.baseData.userData.Role = roleRobot
		user.baseData.userData.LoginIP = m.LoginIP
	}
}
func handleGetCardMa(args []interface{}) {
	_ = args[0].(*msg.C2S_GetCardMa)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		a.Close()
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		a.Close()
		return
	}
	if user.baseData.userData.PlayTimes >= conf.GetCfgCard().PlayTimes {
		user.baseData.userData.CardCode = common.GetTodayCode(5)
		user.baseData.userData.Taken = false
		updateUserData(user.baseData.userData.UserID, bson.M{"$set": bson.M{
			"cardcode":  user.baseData.userData.CardCode,
			"taken":     user.baseData.userData.Taken,
			"playtimes": user.baseData.userData.PlayTimes,
		},
		})
		user.WriteMsg(&msg.S2C_CardMa{
			Code: user.baseData.userData.CardCode,
		})
	}
}
func handleTaskState(args []interface{}) {
	m := args[0].(*msg.C2S_TakeTaskState)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		a.Close()
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		a.Close()
		return
	}
	resultData := new(RedPacketTaskPrizeData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		err := db.DB(DB).C("redpackettaskprize").FindId(m.ID).One(&resultData)
		if err != nil {
			return
		}
	}, func() {
		if resultData.UserID != user.baseData.userData.UserID || resultData.RedPacket <= 0 {
			return
		}
		if resultData.Taken {
			return
		}
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		db.DB(DB).C("redpackettaskprize").UpdateId(m.ID, bson.M{"$set": bson.M{"taken": true}})
		user.redpacketTaskRecord()
	})
}

func handleShareTasks(args []interface{}) {
	_ = args[0].(*msg.C2S_GetShareTaskInfo)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		a.Close()
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		a.Close()
		return
	}
	user.shareTasksInfo()
}

func handleRedpacktTaskCode(args []interface{}) {
	m := args[0].(*msg.C2S_GetRedpacketTaskCode)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		a.Close()
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		a.Close()
		return
	}
	task := user.getRedpacketTask(m.TaskId)
	label := 0
	skeleton.Go(func() {
		if task.State == 2 && task.PlayTimes == task.Total {
			//请求圈圈获取红包码
			temp := &struct {
				Code string
				Data string
			}{}
			r := new(circle.RedPacketCodeInfo)
			r.Sum = task.Fee
			if !task.Real {
				switch task.ID / 1000 {
				case 1:
					r.Sum = primaryRedpacket[rand.Intn(len(primaryRedpacket))]
				case 2:
					r.Sum = middleRedpacket[rand.Intn(len(middleRedpacket))]
				case 3:
					r.Sum = highRedpacket[rand.Intn(len(highRedpacket))]
				}
			}
			param, _ := json.Marshal(r)
			json.Unmarshal(circle.DoRequestRepacketCode(string(param)), temp)
			if temp.Code != "0" {
				user.WriteMsg(&msg.S2C_RedpacketTaskCode{
					Error: msg.S2C_RedpacketTaskInValid,
				})
			}
			if temp.Code == "0" {
				WriteRedPacketGrantRecord(user.baseData.userData, 1, task.Desc, task.Fee)
				user.WriteMsg(&msg.S2C_RedpacketTaskCode{
					Code:  temp.Data,
					Error: 0,
					Desc:  fmt.Sprintf("%.2f元红包", r.Sum),
				})
				//红包任务记录存储
				record := new(msg.RedpacketTaskRecord)
				record.Createdat = time.Now().Unix()
				record.Desc = task.Desc
				record.Fee = r.Sum
				record.ID = task.ID
				record.Real = task.Real
				record.Type = task.Type
				record.RedPacketCode = temp.Data
				user.saveRedPacketTaskRecord(record)
				//刷新任务列表
				for key, value := range user.baseData.redPacketTaskList {
					if value.ID == m.TaskId {
						label = key
						break
					}
				}
				user.baseData.redPacketTaskList = append(user.baseData.redPacketTaskList[:label], user.baseData.redPacketTaskList[label+1:]...)
				user.saveRedPacketTask(user.baseData.redPacketTaskList)
				user.WriteMsg(&msg.S2C_RedpacketTask{
					Tasks:          user.baseData.redPacketTaskList,
					Chips:          ChangeChips[user.baseData.userData.Level],
					FreeChangeTime: user.baseData.userData.FreeChangedAt,
				})
				//刷新红包任务记录
				user.redpacketTaskRecord()
			}
		}
	}, nil)
}

func handleSubsidyChip(args []interface{}) {
	m := args[0].(*msg.C2S_SubsidyChip)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		a.Close()
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		a.Close()
		return
	}

	user.TakenSubsidyChip(m.Reply)
}

func handleIsExistSubsidy(args []interface{}) {
	_ = args[0].(*msg.C2S_IsExistSubsidy)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		a.Close()
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		a.Close()
		return
	}
	user.AskSubsidyChip()
}

func handleShareInfo(args []interface{}) {
	_ = args[0].(*msg.C2S_ShareInfo)
	a := args[1].(gate.Agent)
	if a.UserData() == nil {
		a.Close()
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		a.Close()
		return
	}
	if user.baseData.userData.ParentId == 0 {
		return
	}
}

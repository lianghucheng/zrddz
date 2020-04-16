package internal

import (
	"common"
	"conf"
	"msg"
	"time"

	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
	"gopkg.in/mgo.v2/bson"
)

// 用户状态
const (
	userLogin  = iota
	userLogout // 1
)

const (
	roleRobot  = -2 // 机器人
	roleBlack  = -1 // 黑名单
	rolePlayer = 1  // 玩家
	roleAgent  = 2  // 代理
	roleAdmin  = 3  // 管理员
	roleRoot   = 4  // 最高管理员
)

var (
	userIDUsers = make(map[int]*User)

	userIDRooms     = make(map[int]interface{})
	roomNumberRooms = make(map[string]interface{})

	systemOn = true // 系统开关

	accountIDs         []int
	accountIDCounter   = 0
	reservedAccountIDs = []int{6666666, 8888888, 9999999}

	// 红包类型分别是1、10、100、999
	redPacketMatchOnlineNumber = []int{0, 0, 0, 0} // 红包比赛在线人数

	redPacketCounter = 1
)

type User struct {
	gate.Agent
	state          int
	baseData       *BaseData
	heartbeatTimer *timer.Timer
	heartbeatStop  bool
	LastTaskId     int //该任务在上一任务完成后重新生成(所以对该任务不需要加一处理)
}

type BaseData struct {
	userData               *UserData
	ownerUserID            int               // 所在房间的房主
	taskIDTaskDatas        map[int]*TaskData // 任务列表(包含金币、红包任务)
	togetherUserIDs        map[int]bool      // 记录对局过的玩家
	TaskId                 int
	redPacketTaskList      []msg.RedPacketTask
	redPacketTaskRecord    []msg.RedPacketTask
	NoReceiveRedpacketTask []msg.RedPacketTask
	TaskCount              int //领取的红包任务数量
}

func init() {
	result := new(UserData)

	m := make(map[int]bool)
	for _, v := range reservedAccountIDs {
		m[v] = true
	}

	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		iter := db.DB(DB).C("users").Find(nil).Iter()
		for iter.Next(&result) {
			m[result.AccountID] = true
		}
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
	}, func() {
		for i := 1000000; i < 10000000; i++ {
			if !m[i] {
				accountIDs = append(accountIDs, i)
			}
		}
		accountIDs = common.Shuffle(accountIDs)
		// log.Debug("%v %v", len(m), len(accountIDs))
	})
}

// 生成7位数的账号ID
func getAccountID() int {
	log.Debug("账号ID计数器: %v", accountIDCounter)
	accountID := accountIDs[accountIDCounter]
	accountIDCounter++
	return accountID
}

func newUser(a gate.Agent) *User {
	user := new(User)
	user.Agent = a
	user.state = userLogin
	user.baseData = new(BaseData)
	user.baseData.userData = new(UserData)
	user.baseData.taskIDTaskDatas = make(map[int]*TaskData)
	user.baseData.togetherUserIDs = make(map[int]bool)
	return user
}

func (user *User) autoHeartbeat() {
	if user.heartbeatStop {
		log.Debug("userID: %v 心跳停止", user.baseData.userData.UserID)
		user.Close()
		return
	}
	user.heartbeatStop = true
	user.WriteMsg(&msg.S2C_Heartbeat{})
	// 服务端发送心跳包间隔120秒
	user.heartbeatTimer = skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().HeartTimeout)*time.Second, func() {
		user.autoHeartbeat()
	})
}

func (user *User) transferChips(accountID int, chips int64) {
	otherUserData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		// load
		db.DB(DB).C("users").
			Find(bson.M{"accountid": accountID}).One(otherUserData)
	}, func() {
		if user.state == userLogout {
			return
		}
		if otherUserData.UserID < 1 {
			user.WriteMsg(&msg.S2C_TransferChips{Error: msg.S2C_TransferChips_AccountIDInvalid})
			return
		}
		if otherUser, ok := userIDUsers[otherUserData.UserID]; ok {
			otherUser.baseData.userData.Chips += chips
			//updateUserData(otherUserData.UserID, bson.M{"$set": bson.M{"chips": otherUser.baseData.Chips}})
			otherUser.WriteMsg(&msg.S2C_UpdateUserChips{
				Chips: otherUser.baseData.userData.Chips,
			})
		} else {
			otherUserData.Chips += chips
			updateUserData(otherUserData.UserID, bson.M{"$set": bson.M{"chips": otherUserData.Chips}})
		}
		user.baseData.userData.Chips -= chips
		updateUserData(user.baseData.userData.UserID, bson.M{"$set": bson.M{"chips": user.baseData.userData.Chips}})
		user.WriteMsg(&msg.S2C_UpdateUserChips{
			Chips: user.baseData.userData.Chips,
		})
		user.WriteMsg(&msg.S2C_TransferChips{
			Error: msg.S2C_TransferChips_OK,
			Chips: chips,
		})
		log.Debug("userID %v 给账号ID: %v 转了 %v筹码", user.baseData.userData.UserID, accountID, chips)
	})
}

func (user *User) getAllPlayers(r interface{}) {
	landlordRoom := r.(*LandlordRoom)
	landlordRoom.GetAllPlayers(user)
}

func (user *User) FakeWXPay(totalFee int) {
	if common.InArray([]int{100, 600, 1200, 5000, 10000}, totalFee) || user.isRobot() {
		outTradeNo := common.GetOutTradeNo()
		startWXPayOrder(outTradeNo, user.baseData.userData.AccountID, totalFee, func() {
			finishWXPayOrder(outTradeNo, totalFee, false)
		})
	}
}

func (user *User) FakeAliPay(totalAmount float64) {
	if common.InArray([]int{1, 6, 12, 50, 100}, int(totalAmount)) || user.isRobot() {
		outTradeNo := common.GetOutTradeNo()
		startAliPayOrder(outTradeNo, user.baseData.userData.AccountID, totalAmount, func() {
			finishAliPayOrder(outTradeNo, totalAmount, false)
		})
	}
}

func (user *User) isRobot() bool {
	return user.baseData.userData.Role == roleRobot
}

func (user *User) setRobotChips(chips int64) {
	robotData := new(UserData)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		iter := db.DB(DB).C("users").Find(bson.M{"role": roleRobot}).Iter()
		if err := iter.Close(); err != nil {
			log.Error("iter close error: %v", err)
		}
		for iter.Next(&robotData) {
			if robot, ok := userIDUsers[robotData.UserID]; ok {
				robot.baseData.userData.Chips += chips
			} else {
				updateUserData(robotData.UserID, bson.M{"$inc": bson.M{"chips": chips}})
			}
		}
	}, nil)
}

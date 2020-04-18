package internal

import (
	"fmt"
	"msg"
	"time"

	"conf"

	"github.com/name5566/leaf/log"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	primaryRedpacket = []float64{0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 1.9, 2.0}
	middleRedpacket  = []float64{1.0, 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 1.9, 2.0, 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.9, 3.0}
	highRedpacket    = []float64{1.5, 1.6, 1.7, 1.8, 1.9, 2.0, 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.9, 3.0, 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9, 4.0}
)

type RedpacketTaskList struct {
	UserID int // 用户ID
	Tasks  []msg.RedPacketTask
}

func (user *User) sendRedpacketTask(level int) {
	if user.isRobot() {
		user.delRedPacketTask()
		return

	}
	data := make([]msg.RedPacketTask, 0)
	switch level {
	case 1:
		//已经完成的任务
		//当前的任务
		//玩家没有任务,分配第一个任务
		for _, value := range conf.GetCfgPrimaryTask() {
			task := msg.RedPacketTask{
				ID:    value.ID,
				Real:  value.Real,
				Total: value.Total,
				Fee:   value.Fee,
				Desc:  value.Desc,
				Jump:  value.Jump,
				Type:  value.Type,
			}
			data = append(data, task)
		}
	case 2:
		for _, value := range conf.GetCfgMiddleTask() {
			task := msg.RedPacketTask{
				ID:    value.ID,
				Real:  value.Real,
				Total: value.Total,
				Fee:   value.Fee,
				Desc:  value.Desc,
				Jump:  value.Jump,
				Type:  value.Type,
			}
			data = append(data, task)
		}
	case 3:
		for _, value := range conf.GetCfgHighTask() {
			task := msg.RedPacketTask{
				ID:    value.ID,
				Real:  value.Real,
				Total: value.Total,
				Fee:   value.Fee,
				Desc:  value.Desc,
				Jump:  value.Jump,
				Type:  value.Type,
			}
			data = append(data, task)
		}
	}
	/*
	  查找当自己的任务类型
	*/
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	tasks := new(RedpacketTaskList)
	err := db.DB(DB).C("userDoRedpakcetTask").Find(bson.M{"userid": user.baseData.userData.UserID}).One(tasks)
	if err != nil && err != mgo.ErrNotFound {
		log.Release("查看用户的红包任务报错:%v", err)
		return
	}
	if err == mgo.ErrNotFound {
		data[0].StartTime = time.Now().Unix()
		data[0].State = 1
		data[0].PlayTimes = 0

	}
	if err == nil {
		data = tasks.Tasks
		index := data[len(data)-1].ID
		//判断是否添加了新的红包任务,添加新的红包任务
		if user.baseData.userData.Level == 1 {
			for key, value := range tasks.Tasks {
				for _, v := range conf.GetCfgPrimaryTask() {
					if value.ID == v.ID && value.Desc != v.Desc {
						tasks.Tasks[key].Desc = v.Desc
						break
					}
				}
			}
			if index < conf.GetCfgPrimaryTask()[len(conf.GetCfgPrimaryTask())-1].ID {
				for _, value := range conf.GetCfgPrimaryTask() {
					if index < value.ID {
						task := msg.RedPacketTask{
							ID:    value.ID,
							Real:  value.Real,
							Total: value.Total,
							Fee:   value.Fee,
							Desc:  value.Desc,
							Jump:  value.Jump,
							Type:  value.Type,
						}
						data = append(data, task)
					}
				}
			}
		}

		if user.baseData.userData.Level == 2 {
			for key, value := range tasks.Tasks {
				for _, v := range conf.GetCfgMiddleTask() {
					if value.ID == v.ID && value.Desc != v.Desc {
						tasks.Tasks[key].Desc = v.Desc
						break
					}
				}
			}
			if index < conf.GetCfgMiddleTask()[len(conf.GetCfgMiddleTask())-1].ID {
				for _, value := range conf.GetCfgMiddleTask() {
					if index < value.ID {
						task := msg.RedPacketTask{
							ID:    value.ID,
							Real:  value.Real,
							Total: value.Total,
							Fee:   value.Fee,
							Desc:  value.Desc,
							Jump:  value.Jump,
							Type:  value.Type,
						}
						data = append(data, task)
					}
				}
			}
		}
		if user.baseData.userData.Level == 3 {
			for key, value := range tasks.Tasks {
				for _, v := range conf.GetCfgHighTask() {
					if value.ID == v.ID && value.Desc != v.Desc {
						tasks.Tasks[key].Desc = v.Desc
						break
					}
				}
			}
			if index < conf.GetCfgHighTask()[len(conf.GetCfgHighTask())-1].ID {
				for _, value := range conf.GetCfgHighTask() {
					if index < value.ID {
						task := msg.RedPacketTask{
							ID:    value.ID,
							Real:  value.Real,
							Total: value.Total,
							Fee:   value.Fee,
							Desc:  value.Desc,
							Jump:  value.Jump,
							Type:  value.Type,
						}
						data = append(data, task)
					}
				}
			}
		}
	}
	if user.baseData.NoReceiveRedpacketTask == nil {
		user.baseData.NoReceiveRedpacketTask = make([]msg.RedPacketTask, 0)
	}
	data = append(user.baseData.NoReceiveRedpacketTask, data[:]...)
	user.baseData.NoReceiveRedpacketTask = make([]msg.RedPacketTask, 0)
	user.baseData.redPacketTaskList = data

	user.baseData.TaskId = user.getPlayingTask()

	user.WriteMsg(&msg.S2C_RedpacketTask{
		Tasks:          data,
		Chips:          ChangeChips[user.baseData.userData.Level],
		FreeChangeTime: user.baseData.userData.FreeChangedAt,
	})
	user.saveRedPacketTask(data)
}

//存储当前的任务
func (user *User) saveRedPacketTask(list []msg.RedPacketTask) {
	if user.isRobot() {
		return

	}
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	update := &struct {
		Tasks     []msg.RedPacketTask
		UpdatedAt int64
	}{
		Tasks:     list,
		UpdatedAt: time.Now().Unix(),
	}
	_, err := db.DB(DB).C("userDoRedpakcetTask").Upsert(bson.M{"userid": user.baseData.userData.UserID}, bson.M{"$set": update})
	if err != nil {
		log.Release("update userID: %v usertasklist error: %v", user.baseData.userData.UserID, err)
	}
}

//删除当前的任务
func (user *User) delRedPacketTask() {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	err := db.DB(DB).C("userDoRedpakcetTask").Remove(bson.M{"userid": user.baseData.userData.UserID})
	if err != nil {
		log.Release("remove userID: %v usertasklist error: %v", user.baseData.userData.UserID, err)
	}
}

//更新当前任务

func (user *User) updateRedPacketTask(taskId int) {
	if v, ok := userIDRooms[user.baseData.userData.UserID]; ok {
		r := v.(*LandlordRoom)
		if r.rule.RoomType != roomBaseScoreMatching {
			return
		}
	}
	if user.isRobot() {
		return

	}
	fmt.Println("******************当前玩家正在执行的任务:", user.getPlayingTask())
	fmt.Println("**************************taskid:", taskId)
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	if user.getPlayingTask() == taskId && user.LastTaskId == 0 {
		err := db.DB(DB).C("userDoRedpakcetTask").Update(bson.M{"userid": user.baseData.userData.UserID, "tasks.id": taskId}, bson.M{"$inc": bson.M{"tasks.$.playtimes": 1}})
		if err != nil {
			log.Release("update userID: %v usertasklist error: %v", user.baseData.userData.UserID, err)
		}
		//刷新当前任务列表
		lable := user.getPlayingTaskIndex()
		user.baseData.redPacketTaskList[lable].PlayTimes++
		//当前任务已经完成,重置下一条任务
		if user.baseData.redPacketTaskList[lable].Total == user.baseData.redPacketTaskList[lable].PlayTimes {
			user.baseData.redPacketTaskList[lable].State = 2 //状态变成领取
			user.baseData.TaskCount++
			//完成10个任务,玩家自动升级等级
			if user.baseData.TaskCount == conf.Server.Level {
				user.baseData.userData.Level++
				/*
										删除玩家原来的红包任务，但是不可以删除未被领取的红包任务
					                    当前等级完成的任务清零
										保留未被领取的任务红包
				*/
				for _, value := range user.baseData.redPacketTaskList {
					if value.State == 2 && value.Total == value.PlayTimes {
						user.baseData.NoReceiveRedpacketTask = append(user.baseData.NoReceiveRedpacketTask, value)
					}
				}
				user.delRedPacketTask()
				user.baseData.TaskCount = 0
				user.sendRedpacketTask(user.baseData.userData.Level)
				return
			}
			user.LastTaskId = user.baseData.redPacketTaskList[lable].ID
			if lable != len(user.baseData.redPacketTaskList)-1 {
				user.baseData.redPacketTaskList[lable+1].StartTime = time.Now().Unix()
				user.baseData.redPacketTaskList[lable+1].State = 1
				user.baseData.redPacketTaskList[lable+1].PlayTimes = 0
				user.baseData.TaskId = user.baseData.redPacketTaskList[lable+1].ID
			}
		}
		user.saveRedPacketTask(user.baseData.redPacketTaskList)
		user.WriteMsg(&msg.S2C_RedpacketTask{
			Tasks:          user.baseData.redPacketTaskList,
			Chips:          ChangeChips[user.baseData.userData.Level],
			FreeChangeTime: user.baseData.userData.FreeChangedAt,
		})
	}

}

//更新连续完成的任务

func (user *User) clearRedPacketTask(taskId int) {
	if user.isRobot() {
		return

	}
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	if user.getPlayingTask() == taskId {
		_, err := db.DB(DB).C("userDoRedpakcetTask").Upsert(bson.M{"userid": user.baseData.userData.UserID}, bson.M{"$set": bson.M{"playtimes": 0}})
		if err != nil {
			log.Release("update userID: %v usertasklist error: %v", user.baseData.userData.UserID, err)
		}
		//刷新当前任务列表
		user.baseData.redPacketTaskList[user.getPlayingTaskIndex()].PlayTimes = 0
		user.WriteMsg(&msg.S2C_RedpacketTask{
			Tasks:          user.baseData.redPacketTaskList,
			Chips:          ChangeChips[user.baseData.userData.Level],
			FreeChangeTime: user.baseData.userData.FreeChangedAt,
		})
		user.saveRedPacketTask(user.baseData.redPacketTaskList)
	}
}

//获取玩家当前执行的任务
func (user *User) getPlayingTask() int {
	for _, value := range user.baseData.redPacketTaskList {
		if value.State == 1 {
			return value.ID
		}
	}
	return 0
}

//获取当前执行任务的下标(通过下标修改玩家任务进度)
func (user *User) getPlayingTaskIndex() int {
	for key, value := range user.baseData.redPacketTaskList {
		if value.State == 1 {
			return key
		}
	}
	return 0
}

//获取某个特定任务的数据
func (user *User) getRedpacketTask(taskId int) msg.RedPacketTask {
	for _, value := range user.baseData.redPacketTaskList {
		if value.ID == taskId {
			return value
		}
	}
	return msg.RedPacketTask{}
}

var (
	ChangeChips = map[int]int64{
		1: 5000,
		2: 10000,
		3: 30000,
	}
)

//更换红包任务
func (user *User) changeRedPacketTask(m *msg.C2S_ChangeTask) {
	//新人任务无法进行更换
	if user.getRedpacketTask(user.getPlayingTask()).Type < 3 {
		user.WriteMsg(&msg.S2C_ChangeTask{
			Error: msg.S2C_NewPlayer_NotChange,
		})
		return
	}
	//24小时后免费更换
	Free := false
	if m.Free {
		log.Release("*******************:%v", user.baseData.userData.FreeChangedAt)
		if user.baseData.userData.FreeChangedAt > time.Now().Unix() {
			user.WriteMsg(&msg.S2C_ChangeTask{
				Error: msg.S2C_ChangeTask_NotReachTime,
			})
			return
		}
		Free = true
		user.baseData.userData.FreeChangedAt = time.Now().Unix() + conf.Server.TaskFreeChange*60*60
	}
	//消耗金币更换 初 中 高 对应的金币比例是5000 10000 30000金币
	if !Free {
		if user.baseData.userData.Chips < ChangeChips[user.baseData.userData.Level] {
			user.WriteMsg(&msg.S2C_ChangeTask{
				Error: msg.S2C_ChipsLack,
			})
			return
		}
		/*
			if r, ok := userIDRooms[user.baseData.userData.UserID]; ok {
				Room := r.(*Room)
				if Room.rule.RoomType == roomMatch {
					if user.baseData.userData.Chips < ChangeChips[user.baseData.userData.Level]+int64(Room.rule) {
						user.WriteMsg(&msg.S2C_ChangeTask{
							Error: msg.S2C_ChipsLack,
						})
						return
					}

				}
			}
		*/

	}
	if len(user.baseData.redPacketTaskList) == 1 {
		user.WriteMsg(&msg.S2C_ChangeTask{
			Error: msg.S2C_NoTaskChange,
		})
		return
	}
	if !Free {
		user.baseData.userData.Chips -= ChangeChips[user.baseData.userData.Level]
		user.WriteMsg(&msg.S2C_UpdateUserChips{
			Chips: user.baseData.userData.Chips,
		})
	}
	tasks := make([]msg.RedPacketTask, 0)
	label := 0
	for key, value := range user.baseData.redPacketTaskList {
		if value.ID == user.getPlayingTask() {

			if key == len(user.baseData.redPacketTaskList)-1 {
				user.WriteMsg(&msg.S2C_ChangeTask{
					Error: msg.S2C_NoTaskChange,
				})
				return
			}
			user.baseData.redPacketTaskList[key].StartTime = 0
			user.baseData.redPacketTaskList[key].State = 0
			user.baseData.redPacketTaskList[key].PlayTimes = 0
			label = key
			break
		}
	}
	//先整理领取状态的任务
	for _, value := range user.baseData.redPacketTaskList {
		if value.State == 2 {
			tasks = append(tasks, value)
		}
	}
	//整理更换后排序的任务
	user.baseData.redPacketTaskList[label+1].StartTime = time.Now().Unix()
	user.baseData.redPacketTaskList[label+1].State = 1
	tasks = append(tasks, user.baseData.redPacketTaskList[label+1])
	for key, value := range user.baseData.redPacketTaskList {
		if key == label {
			continue
		}
		if value.State == 0 {
			tasks = append(tasks, value)
		}
	}
	tasks = append(tasks, user.baseData.redPacketTaskList[label])
	user.baseData.TaskId = user.baseData.redPacketTaskList[label+1].ID
	user.baseData.redPacketTaskList = tasks
	user.WriteMsg(&msg.S2C_ChangeTask{})
	user.WriteMsg(&msg.S2C_RedpacketTask{
		Tasks:          user.baseData.redPacketTaskList,
		Chips:          ChangeChips[user.baseData.userData.Level],
		FreeChangeTime: user.baseData.userData.FreeChangedAt,
	})
	user.saveRedPacketTask(user.baseData.redPacketTaskList)
}

//红包领取记录
type TasksRecord struct {
	UserID int // 用户ID
	Tasks  []msg.RedpacketTaskRecord
}

func (user *User) redpacketTaskRecord() {
	if user.isRobot() {
		return

	}
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	record := new(TasksRecord)
	err := db.DB(DB).C("redpacketTaskRecord").Find(bson.M{"_id": user.baseData.userData.UserID}).One(&record)
	if err != nil && err != mgo.ErrNotFound {
		log.Release("查看用户的红包任务记录报错:%v", err)
		return
	}
	if err == mgo.ErrNotFound {
		record.Tasks = make([]msg.RedpacketTaskRecord, 0)
	}
	if len(record.Tasks) > 10 {
		record.Tasks = record.Tasks[:10]
	}
	user.WriteMsg(&msg.S2C_RedPacketTaskRecord{
		TaskRecords: record.Tasks,
	})
	user.countRedpacketTaskRecord()
}

func (user *User) saveRedPacketTaskRecord(r *msg.RedpacketTaskRecord) {
	selector := user.baseData.userData.UserID
	update := bson.M{"$push": bson.M{"tasks": bson.M{"$each": []msg.RedpacketTaskRecord{*r}, "$sort": bson.M{"createdat": -1}}}}
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	_, err := db.DB(DB).C("redpacketTaskRecord").UpsertId(selector, update)
	if err != nil && err != mgo.ErrNotFound {
		log.Release("******************:%v", err)
	}
}

//获取用户当前等级的
func (user *User) countRedpacketTaskRecord() {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	record := new(TasksRecord)
	err := db.DB(DB).C("redpacketTaskRecord").Find(bson.M{"_id": user.baseData.userData.UserID}).One(&record)
	if err != nil && err != mgo.ErrNotFound {
		return
	}
	if err == mgo.ErrNotFound {
		record.Tasks = make([]msg.RedpacketTaskRecord, 0)
	}
	//已领取的任务红包
	for _, value := range record.Tasks {

		if value.ID/1000 == user.baseData.userData.Level && value.Type > 2 {
			user.baseData.TaskCount++
		}
	}
	//任务已经完成,还没有领取的红包任务
	for _, value := range user.baseData.redPacketTaskList {

		if user.baseData.userData.Level == value.ID/1000 && value.State == 2 && value.Type > 2 {
			user.baseData.TaskCount++
		}
	}
	log.Release("**************************完成的当前级别红包任务次数:%v", user.baseData.TaskCount)
}

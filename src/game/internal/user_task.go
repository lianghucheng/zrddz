package internal

import (
	"common"
	"game/circle"
	"math/rand"
	"msg"
	"time"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

func (user *User) sendTaskList(firstLogin bool, cb func()) {
	var redPacketTaskList []TaskData
	var chipTaskList []TaskData

	skeleton.Go(func() {
		if len(user.baseData.taskIDTaskDatas) > 0 {
			redPacketTaskList = toTaskList(user.baseData.taskIDTaskDatas, taskRedPacket)
			chipTaskList = toTaskList(user.baseData.taskIDTaskDatas, taskChip)
			return
		}

		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		data := new(UserTaskListData)
		err := db.DB(DB).C("usertasklist").
			Find(bson.M{"userid": user.baseData.userData.UserID}).One(&data)
		if err == nil {
			redPacketTaskList = data.RedPacketTaskList
			redPacketTaskList = updateRedPacketTaskList(redPacketTaskList)

			chipTaskList = data.ChipTaskList
			chipTaskList = updateChipTaskList(chipTaskList)
		} else {
			var taskIDs []int
			if firstLogin {
				taskIDs = getRandomTaskIDs(1, taskRedPacket, FirstLoginRedPacketIDs)
			} else {
				if user.baseData.userData.Chips >= 50000 {
					taskIDs = getRandomTaskIDs(1, taskRedPacket, RedPacketIDs2)
				} else {
					taskIDs = getRandomTaskIDs(1, taskRedPacket, RedPacketIDs)
				}
			}
			redPacketTaskList = buildTaskList(taskIDs)

			taskIDs = []int{1000}
			// taskIDs = append(taskIDs, getRandomTaskIDs(3, taskChip)...)
			chipTaskList = buildTaskList(taskIDs)
		}
		// 活动任务更新和移除
		chipTaskList = updateActivityTaskList(chipTaskList)

		user.baseData.taskIDTaskDatas = loadTaskList(redPacketTaskList, user.baseData.taskIDTaskDatas)
		user.baseData.taskIDTaskDatas = loadTaskList(chipTaskList, user.baseData.taskIDTaskDatas)
	}, func() {
		user.WriteMsg(&msg.S2C_UpdateRedPacketTaskList{
			Items: buildTaskItems(redPacketTaskList),
		})
		user.WriteMsg(&msg.S2C_UpdateChipTaskList{
			Items: buildTaskItems(chipTaskList),
		})
		if cb != nil {
			cb()
		}
	})
}

func (user *User) saveTaskList() {
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		update := &struct {
			RedPacketTaskList []TaskData
			ChipTaskList      []TaskData
			UpdatedAt         int64
		}{
			RedPacketTaskList: toTaskList(user.baseData.taskIDTaskDatas, taskRedPacket),
			ChipTaskList:      toTaskList(user.baseData.taskIDTaskDatas, taskChip),
			UpdatedAt:         time.Now().Unix(),
		}
		_, err := db.DB(DB).C("usertasklist").
			Upsert(bson.M{"userid": user.baseData.userData.UserID}, bson.M{"$set": update})
		if err != nil {
			log.Debug("update userID: %v usertasklist error: %v", user.baseData.userData.UserID, err)
		}
	}, nil)
}

func (user *User) doTask(taskID int) {
	if task, ok := user.baseData.taskIDTaskDatas[taskID]; ok {
		if task.Progress < TaskList[taskID].Total {
			task.Progress++
			if TaskList[taskID].Type == taskRedPacket && task.Progress == TaskList[taskID].Total {
				upsertTaskTicket(bson.M{"userid": user.baseData.userData.UserID, "finish": false, "taskid": task.TaskID},
					bson.M{"$set": bson.M{"finish": true, "updatedat": time.Now().Unix()}})
			}
			user.WriteMsg(&msg.S2C_UpdateTaskProgress{
				TaskID:   task.TaskID,
				Progress: task.Progress,
			})
		}
	}
}

func (user *User) takeTaskPrize(taskID int) {
	task := user.baseData.taskIDTaskDatas[taskID]
	if task.Taken {
		user.WriteMsg(&msg.S2C_TakeTaskPrize{Error: msg.S2C_TakeTaskPrize_Repeated})
		return
	}
	if task.Progress < TaskList[taskID].Total {
		user.WriteMsg(&msg.S2C_TakeTaskPrize{Error: msg.S2C_TakeTaskPrize_NotDone})
		return
	}
	switch TaskList[taskID].Type {
	case taskRedPacket:
		if task.Handling {
			return
		}
		task.Handling = true
		redPacket := rand.Float64()/2 + 1
		if redPacketCounter%200 == 0 {
			redPacket += float64(rand.Intn(3)) + 1
		}
		redPacket = common.Decimal(redPacket)
		userID := user.baseData.userData.UserID

		go func() {
			exchangeCode, err := circle.GetRedPacketCode(redPacket)
			if err != nil {
				takeRedPacketTaskPrizeFail(userID, taskID)
				return
			}
			takeRedPacketTaskPrizeSuccess(userID, taskID, redPacket, exchangeCode)
			WriteRedPacketGrantRecord(user.baseData.userData, 1, TaskList[taskID].Desc, redPacket)
			redPacketCounter++
		}()
	case taskChip:
		task.Taken = true
		task.TakenAt = time.Now().Unix()

		chips := TaskList[taskID].Chips
		user.baseData.userData.Chips += chips
		user.WriteMsg(&msg.S2C_UpdateUserChips{
			Chips: user.baseData.userData.Chips,
		})
		user.WriteMsg(&msg.S2C_TakeTaskPrize{
			Error:  msg.S2C_TakeTaskPrize_TakeChipPrizeOK,
			TaskID: taskID,
			Chips:  chips,
		})
		saveChipTaskPrize(user.baseData.userData.UserID, taskID)
	}
}

func takeRedPacketTaskPrizeSuccess(userID, taskID int, redPacket float64, exchangeCode string) {
	saveRedPacketTaskPrize(userID, taskID, redPacket, exchangeCode)
	if user, ok := userIDUsers[userID]; ok {
		if task, ok := user.baseData.taskIDTaskDatas[taskID]; ok {
			task.Taken = true
			task.TakenAt = time.Now().Unix()
			task.Handling = false
			user.WriteMsg(&msg.S2C_TakeTaskPrize{
				Error:        msg.S2C_TakeTaskPrize_TakeRedPacketPrizeOK,
				TaskID:       taskID,
				RedPacket:    common.Round(redPacket, 2),
				ExchangeCode: exchangeCode,
			})
			user.updateRedPacketTaskList()
		}
		user.redpacketTaskRecord()
	} else {
		task := &TaskData{
			TaskID:   taskID,
			Progress: TaskList[taskID].Total,
			Taken:    true,
			TakenAt:  time.Now().Unix(),
			Handling: false,
		}
		updateUserTask(userID, task)
	}
}

func takeRedPacketTaskPrizeFail(userID, taskID int) {
	if user, ok := userIDUsers[userID]; ok {
		if task, ok := user.baseData.taskIDTaskDatas[taskID]; ok {
			task.Handling = false
			user.WriteMsg(&msg.S2C_TakeTaskPrize{
				Error:  msg.S2C_TakeTaskPrize_Error,
				TaskID: taskID,
			})
			user.WriteMsg(&msg.S2C_UpdateTaskProgress{
				TaskID:   taskID,
				Progress: task.Progress,
			})
		}
	} else {
		task := &TaskData{
			TaskID:   taskID,
			Progress: TaskList[taskID].Total,
			Handling: false,
		}
		updateUserTask(userID, task)
	}
}

func saveChipTaskPrize(userID, taskID int) {
	now := time.Now()
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		insert := &ChipTaskPrizeData{
			UserID:    userID,
			TaskID:    taskID,
			Chips:     TaskList[taskID].Chips,
			CreatedAt: now.Unix(),
			UpdatedAt: now.Unix(),
		}
		err := db.DB(DB).C("chiptaskprize").Insert(insert)
		if err != nil {
			log.Debug("insert userid: %v chiptaskprize error: %v", userID, err)
		}
	}, nil)
}

func saveRedPacketTaskPrize(userID, taskID int, redPacket float64, exchangeCode string) {
	now := time.Now()
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	insert := &RedPacketTaskPrizeData{
		UserID:       userID,
		TaskID:       taskID,
		RedPacket:    redPacket,
		ExchangeCode: exchangeCode,
		Desc:         TaskList[taskID].Desc,
		CreatedAt:    now.Unix(),
		UpdatedAt:    now.Unix(),
		Taken:        false,
	}
	err := db.DB(DB).C("redpackettaskprize").Insert(insert)
	if err != nil {
		log.Debug("insert userid: %v redpackettaskprize error: %v", userID, err)
	}
}

func (user *User) updateRedPacketTaskList() {
	n := 0
	for _, task := range user.baseData.taskIDTaskDatas {
		switch TaskList[task.TaskID].Type {
		case taskRedPacket:
			if task.Taken {
				delete(user.baseData.taskIDTaskDatas, task.TaskID)
			} else {
				n++
			}
		}
	}
	if n == 0 {
		var taskIDs []int
		if user.baseData.userData.Chips >= 50000 {
			taskIDs = getRandomTaskIDs(1, taskRedPacket, RedPacketIDs2)
		} else {
			taskIDs = getRandomTaskIDs(1, taskRedPacket, RedPacketIDs)
		}
		redPacketTaskList := buildTaskList(taskIDs)
		user.baseData.taskIDTaskDatas = loadTaskList(redPacketTaskList, user.baseData.taskIDTaskDatas)
		user.WriteMsg(&msg.S2C_UpdateRedPacketTaskList{
			Items: buildTaskItems(redPacketTaskList),
		})
	}
}

func (user *User) changeRedPacketTaskList(free bool) {
}

func saveRedPacketTaskChange(userID int, oldTaskID, newTaskID []int, free bool) {
	now, fare := time.Now().Unix(), int64(0)
	if !free {
		fare = 5000
	}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		err := db.DB(DB).C("redpackettaskchange").Insert(struct {
			UserID    int
			OldTaskID []int
			NewTaskID []int
			Fare      int64
			CreatedAt int64
		}{
			UserID:    userID,
			OldTaskID: oldTaskID,
			NewTaskID: newTaskID,
			Fare:      fare,
			CreatedAt: now,
		})
		if err != nil {
			log.Debug("insert userid: %v redpackettaskprize error: %v", userID, err)
		}
	}, nil)
}

func (user *User) freeChangeCountDown() {
	countDown := time.Now().Unix() - user.baseData.userData.FreeChangedAt - 24*60*60
	if countDown < 0 {
		countDown = -countDown
	} else {
		countDown = 0
	}
	user.WriteMsg(&msg.S2C_FreeChangeCountDown{
		Second: countDown,
	})
}

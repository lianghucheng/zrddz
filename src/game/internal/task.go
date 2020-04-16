package internal

import (
	"common"
	"msg"
	"sort"
	"time"

	"github.com/name5566/leaf/log"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func getRandomTaskIDs(n, taskType int, fromTaskIDs []int) []int {
	var taskIDs []int
	switch taskType {
	case taskRedPacket:
		taskIDs = common.Shuffle(fromTaskIDs)
		if checkRedPacketMatchingTime() {
			taskIDs = common.RemoveAll(taskIDs, 28)
		}
	case taskChip:
		taskIDs = common.Shuffle(ChipTaskIDs[2:])
	default:
		return []int{}
	}
	temp := append([]int{}, taskIDs[:n]...)
	sort.Ints(temp)
	return temp
}

func loadTaskList(taskList []TaskData, m map[int]*TaskData) map[int]*TaskData {
	for _, taskData := range taskList {
		task := &TaskData{
			TaskID:   taskData.TaskID,
			Progress: taskData.Progress,
			Taken:    taskData.Taken,
			TakenAt:  taskData.TakenAt,
			Handling: taskData.Handling,
		}
		m[taskData.TaskID] = task
	}
	return m
}

func buildTaskList(taskIDs []int) []TaskData {
	var taskList []TaskData
	for _, taskID := range taskIDs {
		taskList = append(taskList, TaskData{
			TaskID: taskID,
		})
	}
	return taskList
}

func addTaskItem(items []msg.TaskItem, taskID, progress int, taken bool) []msg.TaskItem {
	items = append(items, msg.TaskItem{
		TaskID:   taskID,
		Progress: progress,
		Taken:    taken,
		Total:    TaskList[taskID].Total,
		Desc:     TaskList[taskID].Desc,
		Chips:    TaskList[taskID].Chips,
		Jump:     TaskList[taskID].Jump,
	})
	return items
}

func buildTaskItems(taskList []TaskData) []msg.TaskItem {
	var taskItems []msg.TaskItem
	m := make(map[int]bool)
	for _, taskData := range taskList {
		// 可以领取奖励的任务
		if !m[taskData.TaskID] && taskData.Progress >= TaskList[taskData.TaskID].Total && !taskData.Taken {
			taskItems = addTaskItem(taskItems, taskData.TaskID, TaskList[taskData.TaskID].Total, taskData.Taken)
			m[taskData.TaskID] = true
		}
	}
	for _, taskData := range taskList {
		// 进行中的任务
		if !m[taskData.TaskID] && taskData.Progress < TaskList[taskData.TaskID].Total {
			taskItems = addTaskItem(taskItems, taskData.TaskID, taskData.Progress, taskData.Taken)
			m[taskData.TaskID] = true
		}
	}
	for _, taskData := range taskList {
		// 已领取奖励的任务
		if !m[taskData.TaskID] && taskData.Progress >= TaskList[taskData.TaskID].Total && taskData.Taken {
			taskItems = addTaskItem(taskItems, taskData.TaskID, TaskList[taskData.TaskID].Total, taskData.Taken)
			m[taskData.TaskID] = true
		}
	}
	return taskItems
}

func getOrderKeys(m map[int]*TaskData) []int {
	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func toTaskList(m map[int]*TaskData, taskType int) []TaskData {
	var taskList []TaskData
	keys := getOrderKeys(m)
	for _, k := range keys {
		taskData := m[k]
		if TaskList[taskData.TaskID].Type == taskType {
			taskList = append(taskList, TaskData{
				TaskID:   taskData.TaskID,
				Progress: taskData.Progress,
				Taken:    taskData.Taken,
				TakenAt:  taskData.TakenAt,
				Handling: taskData.Handling,
			})
		}
	}
	return taskList
}

// 更新奖励被领取的红包任务
func updateRedPacketTaskList(taskList []TaskData) []TaskData {
	var newTaskList []TaskData
	var existTaskIDs []int
	n := 0
	for _, taskData := range taskList {
		// 更新不可用任务
		if _, ok := TaskList[taskData.TaskID]; !ok || taskData.Taken {
			n++
		} else {
			newTaskList = append(newTaskList, taskData)
			existTaskIDs = append(existTaskIDs, taskData.TaskID)
		}
	}
	if n > 0 {
		taskIDs := common.Remove(RedPacketIDs, existTaskIDs)
		taskIDs = common.Shuffle(taskIDs)
		newTaskList = append(newTaskList, buildTaskList(taskIDs[:n])...)
	}
	return newTaskList
}

// 更新奖励被领取的金币任务
func updateChipTaskList(taskList []TaskData) []TaskData {
	var newTaskList []TaskData
	var existTaskIDs []int
	n := 0
	for _, taskData := range taskList {
		// 活动任务直接追加进去
		if _, ok := ActivityTimeList[taskData.TaskID]; ok {
			newTaskList = append(newTaskList, taskData)
			existTaskIDs = append(existTaskIDs, taskData.TaskID)
			continue
		}
		// 任务转换，如果删表可以删除
		switch taskData.TaskID {
		case 12:
			taskData.TaskID = 1000
		}
		// 更新不可用任务
		if _, ok := TaskList[taskData.TaskID]; !ok {
			//n++
			continue
		}
		if taskData.Taken {
			takenTime := time.Unix(taskData.TakenAt, 0)
			takenMidday := time.Date(takenTime.Year(), takenTime.Month(), takenTime.Day(), 12, 0, 0, 0, time.Local)
			if takenMidday.Unix() > taskData.TakenAt {
				takenMidday = takenMidday.Add(-24 * time.Hour)
			}
			if time.Now().Sub(takenMidday).Hours() >= 24 {
				switch taskData.TaskID {
				case 1000, 1001:
					taskData.Progress = 0
					taskData.Taken = false
					newTaskList = append(newTaskList, taskData)
					existTaskIDs = append(existTaskIDs, taskData.TaskID)
				default:
					n++
				}
			} else {
				newTaskList = append(newTaskList, taskData)
				existTaskIDs = append(existTaskIDs, taskData.TaskID)
			}
		} else {
			newTaskList = append(newTaskList, taskData)
			existTaskIDs = append(existTaskIDs, taskData.TaskID)
		}
	}
	if n > 0 {
		taskIDs := common.Remove(ChipTaskIDs[2:], existTaskIDs)
		taskIDs = common.Shuffle(taskIDs)
		newTaskList = append(newTaskList, buildTaskList(taskIDs[:n])...)
	}
	return newTaskList
}

func updateActivityTaskList(taskList []TaskData) []TaskData {
	now := time.Now().Unix()
	var newTaskList []TaskData
	// 非活动任务添加
	for _, taskData := range taskList {
		if _, ok := ActivityTimeList[taskData.TaskID]; !ok {
			newTaskList = append(newTaskList, taskData)
		}
	}
	for id, activityTime := range ActivityTimeList {
		// 检查活动任务时效性
		if now < activityTime.Start || now > activityTime.Deadline {
			continue // 移除任务
		}
		have := false // 是否有这个任务
		for _, taskData := range taskList {
			if id == taskData.TaskID {
				have = true
				newTaskList = append(newTaskList, taskData)
				break
			}
		}
		if !have { // 没有就添加任务
			if now < activityTime.End {
				newTaskList = append(newTaskList, TaskData{
					TaskID: id,
				})
			}
		}
	}
	return newTaskList
}

func addTaskProgress(userID, taskID int) {
	task := TaskList[taskID]
	if task == nil {
		log.Debug("taskID: %v 不存在", taskID)
		return
	}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		var err error
		switch task.Type {
		case taskRedPacket:
			err = db.DB(DB).C("usertasklist").
				Update(bson.M{"userid": userID, "redpackettasklist.taskid": taskID, "redpackettasklist.progress": bson.M{"$lt": task.Total}},
					bson.M{"$inc": bson.M{"redpackettasklist.$.progress": 1}})
			if err == nil {
				count, _ := db.DB(DB).C("usertasklist").
					Find(bson.M{"userid": userID, "redpackettasklist.taskid": taskID, "redpackettasklist.progress": bson.M{"$eq": task.Total}}).
					Count()
				if count == 1 {
					upsertTaskTicket(bson.M{"userid": userID, "finish": false, "taskid": taskID},
						bson.M{"$set": bson.M{"finish": true, "updatedat": time.Now().Unix()}})
				}
			}
		case taskChip:
			err = db.DB(DB).C("usertasklist").
				Update(bson.M{"userid": userID, "chiptasklist.taskid": taskID, "chiptasklist.progress": bson.M{"$lt": task.Total}},
					bson.M{"$inc": bson.M{"chiptasklist.$.progress": 1}})
		}
		if err != nil && err != mgo.ErrNotFound {
			log.Debug("update userID: %v usertasklist error: %v", userID, err)
		}
	}, nil)
}

func resetTaskProgress(userID, taskID int) {
	task := TaskList[taskID]
	if task == nil {
		log.Error("taskID: %v 不存在", taskID)
		return
	}
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		var err error
		switch task.Type {
		case taskRedPacket:
			err = db.DB(DB).C("usertasklist").
				Update(bson.M{"userid": userID, "redpackettasklist.taskid": taskID, "redpackettasklist.progress": bson.M{"$lt": task.Total}},
					bson.M{"$set": bson.M{"redpackettasklist.$.progress": 0}})
		case taskChip:
			err = db.DB(DB).C("usertasklist").
				Update(bson.M{"userid": userID, "chiptasklist.taskid": taskID, "chiptasklist.progress": bson.M{"$lt": task.Total}},
					bson.M{"$set": bson.M{"chiptasklist.$.progress": 0}})
		}
		if err != nil && err != mgo.ErrNotFound {
			log.Debug("update userID: %v usertasklist error: %v", userID, err)
		}
	}, nil)
}

func updateUserTask(userID int, task *TaskData) {
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		var err error
		switch TaskList[task.TaskID].Type {
		case taskRedPacket:
			err = db.DB(DB).C("usertasklist").
				Update(bson.M{"userid": userID, "redpackettasklist.taskid": task.TaskID},
					bson.M{"$set": bson.M{"redpackettasklist.$": task}})
		case taskChip:
			err = db.DB(DB).C("usertasklist").
				Update(bson.M{"userid": userID, "chiptasklist.taskid": task.TaskID},
					bson.M{"$set": bson.M{"chiptasklist.$": task}})
		}
		if err != nil {
			log.Debug("update userID: %v usertasklist error: %v", userID, err)
		}
	}, nil)
}

// 做活动任务
func doActivityTask(userID int, taskID int) {
	activityTime, ok := ActivityTimeList[taskID]
	if !ok {
		return
	}
	now := time.Now().Unix()
	if now >= activityTime.Start && now < activityTime.End {
		if user, ok := userIDUsers[userID]; ok {
			user.doTask(taskID)
		} else {
			addTaskProgress(userID, taskID)
		}
	}
}

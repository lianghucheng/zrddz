package msg

import "gopkg.in/mgo.v2/bson"

type TaskItem struct {
	TaskID   int    // 任务ID
	Progress int    // 进度
	Taken    bool   // 是否被领奖
	Total    int    // 总进度
	Desc     string // 描述
	Chips    int64  // 金币奖励
	Jump     int    // 跳转
}

type S2C_UpdateRedPacketTaskList struct {
	Items []TaskItem
}

type S2C_UpdateChipTaskList struct {
	Items []TaskItem
}

type C2S_DoTask struct {
	TaskID int
}

type S2C_UpdateTaskProgress struct {
	TaskID   int
	Progress int
}

type C2S_TakeTaskPrize struct {
	TaskID int
}

const (
	S2C_TakeTaskPrize_TakeChipPrizeOK      = 0 // 恭喜获得 S2C_TakeTaskPrize.Chips金币奖励
	S2C_TakeTaskPrize_TakeRedPacketPrizeOK = 1 // 恭喜获得 S2C_TakeTaskPrize.RedPacket元红包奖励，打开圈圈可领取您的红包
	S2C_TakeTaskPrize_TaskIDInvalid        = 2 // 任务ID无效
	S2C_TakeTaskPrize_NotDone              = 3 // 离领奖还差一点点，请继续努力吧
	S2C_TakeTaskPrize_Repeated             = 4 // 奖励已被领取，请勿重复操作
	S2C_TakeTaskPrize_Error                = 5 // 领取出错，请稍后重试
)

type S2C_TakeTaskPrize struct {
	Error        int
	TaskID       int
	RedPacket    float64
	Chips        int64
	ExchangeCode string
}

type C2S_FreeChangeCountDown struct {
}

type S2C_FreeChangeCountDown struct {
	Second int64 // 倒计时
}

// 领取任务红包比赛奖励
type C2S_TakeTaskState struct {
	ID bson.ObjectId
}

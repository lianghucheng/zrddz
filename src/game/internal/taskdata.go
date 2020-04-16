package internal

import (
	"conf"
	"time"
)

// 任务类型
const (
	_             = iota
	taskRedPacket // 1 红包任务
	taskChip      // 2 金币任务
)

// 跳转
const (
	_                 = iota
	basescoreMatching // 1 底分匹配
	redpacketMatch    // 2 红包比赛
	wechatShare       // 3 微信分享
	shop              // 4 商店
	wechatInvite      // 5 邀请
)

type TaskMetaData struct {
	ID        int
	Type      int     // 类型
	Total     int     // 总数
	RedPacket float64 // 红包奖励
	Chips     int64   // 金币奖励
	Desc      string  // 描述
	Jump      int     // 跳转
}

// 金币任务奖励数据(存数据库)
type ChipTaskPrizeData struct {
	UserID    int
	TaskID    int
	Chips     int64
	CreatedAt int64
	UpdatedAt int64
}

// 红包任务奖励数据(存数据库)
type RedPacketTaskPrizeData struct {
	UserID       int
	TaskID       int
	RedPacket    float64 // 红包奖励
	ExchangeCode string
	Desc         string
	CreatedAt    int64
	UpdatedAt    int64
	Taken        bool //  true点击过复制按钮 false 没有点击过复制按钮
}

// (存数据库)
type TaskData struct {
	TaskID   int  // 任务ID
	Progress int  // 任务进度
	Taken    bool // 奖励是否被领取
	TakenAt  int64
	Handling bool // 处理中
}

// 用户任务列表数据(存数据库)
type UserTaskListData struct {
	UserID            int        // 用户ID
	RedPacketTaskList []TaskData // 红包任务列表
	ChipTaskList      []TaskData // 金币任务列表
	UpdatedAt         int64
}

// 活动任务时间表
type ActivityTaskSchedule struct {
	Start    int64 // 开始时间
	End      int64 // 结束时间
	Deadline int64 // 截至时间
}

var (
	TaskList         = make(map[int]*TaskMetaData)
	ActivityTimeList = make(map[int]*ActivityTaskSchedule) // 活动时间列表
	// 小于五万
	RedPacketIDs = []int{28, 28, 28, 28, 28, 11, 1, 1, 1, 1, 1, 33, 33, 34, 34, 34, 46, 46, 31, 31, 47, 47, 47, 47, 47, 48, 48, 48, 49, 49, 49, 49, 49, 50, 50, 61, 62, 32, 32, 32, 32, 53, 53, 54, 54, 56, 56, 56,
		38, 38, 38, 38, 40, 40, 40, 40, 40, 63, 63, 63, 59, 59, 59, 59, 43, 43, 45, 45, 45}
	// 大于5万
	RedPacketIDs2 = []int{28, 28, 28, 28, 28, 11, 1, 1, 1, 1, 33, 33, 31, 31, 31, 31, 48, 48, 48, 48, 48, 49, 49, 50, 53, 53, 54, 54, 37, 37, 38, 38, 38, 55, 55, 56, 56, 56, 56, 56, 40, 57, 57, 57, 63, 63,
		58, 58, 58, 58, 41, 41, 41, 59, 60, 60, 61, 61, 43, 44, 44, 45}
	ChipTaskIDs            = []int{1000, 1001, 1002, 1003, 1004, 1005, 1006, 1008, 1009, 1010, 1011, 1012, 1013, 1016}
	FirstLoginRedPacketIDs = []int{1, 1, 1, 33, 33, 34, 34, 34, 46, 46, 31, 31, 47, 47, 47, 48, 48, 48, 49, 49, 49, 50, 50, 61, 62, 32, 32, 53, 53, 54, 54, 38, 40, 63, 63, 59, 43, 45}
)

func init() {
	//红包任务
	TaskList[0] = &TaskMetaData{ID: 0, Type: taskRedPacket, Total: 10, Desc: "胜利10局", Jump: basescoreMatching}
	TaskList[1] = &TaskMetaData{ID: 1, Type: taskRedPacket, Total: 3, Desc: "打出3个飞机", Jump: basescoreMatching}
	TaskList[2] = &TaskMetaData{ID: 2, Type: taskRedPacket, Total: 5, Desc: "打出5个连对", Jump: basescoreMatching}
	TaskList[3] = &TaskMetaData{ID: 3, Type: taskRedPacket, Total: 3, Desc: "打出3个炸弹", Jump: basescoreMatching}
	TaskList[4] = &TaskMetaData{ID: 4, Type: taskRedPacket, Total: 3, Desc: "打出3个王炸", Jump: basescoreMatching}
	TaskList[5] = &TaskMetaData{ID: 5, Type: taskRedPacket, Total: 10, Desc: "对局10局", Jump: basescoreMatching}
	TaskList[6] = &TaskMetaData{ID: 6, Type: taskRedPacket, Total: 2, Desc: "连胜2局", Jump: basescoreMatching}
	TaskList[7] = &TaskMetaData{ID: 7, Type: taskRedPacket, Total: 10, Desc: "当地主10次", Jump: basescoreMatching}
	TaskList[8] = &TaskMetaData{ID: 8, Type: taskRedPacket, Total: 10, Desc: "抢地主10次", Jump: basescoreMatching}
	TaskList[9] = &TaskMetaData{ID: 9, Type: taskRedPacket, Total: 10, Desc: "明牌开始10次", Jump: basescoreMatching}
	TaskList[10] = &TaskMetaData{ID: 10, Type: taskRedPacket, Total: 10, Desc: "加倍10次", Jump: basescoreMatching}
	TaskList[11] = &TaskMetaData{ID: 11, Type: taskRedPacket, Total: 1, Desc: "购买任意数量的金币", Jump: shop}

	TaskList[28] = &TaskMetaData{ID: 28, Type: taskRedPacket, Total: 1, Desc: "参加一次红包比赛场", Jump: redpacketMatch}

	TaskList[30] = &TaskMetaData{ID: 30, Type: taskRedPacket, Total: 4, Desc: "打出4个炸弹", Jump: basescoreMatching}
	TaskList[31] = &TaskMetaData{ID: 31, Type: taskRedPacket, Total: 4, Desc: "打出4个王炸", Jump: basescoreMatching}
	TaskList[32] = &TaskMetaData{ID: 32, Type: taskRedPacket, Total: 2, Desc: "单局打出2个炸弹", Jump: basescoreMatching}
	TaskList[33] = &TaskMetaData{ID: 33, Type: taskRedPacket, Total: 6, Desc: "打出6个连对", Jump: basescoreMatching} // 打出6个连对
	TaskList[34] = &TaskMetaData{ID: 34, Type: taskRedPacket, Total: 5, Desc: "打出5个炸弹", Jump: basescoreMatching} // 打出5个炸弹

	TaskList[36] = &TaskMetaData{ID: 36, Type: taskRedPacket, Total: 10, Desc: "打出10个顺子", Jump: basescoreMatching}    // 打出10个顺子
	TaskList[37] = &TaskMetaData{ID: 37, Type: taskRedPacket, Total: 10, Desc: "普通场累计对局10次", Jump: basescoreMatching} // 普通场累计对局10次
	TaskList[38] = &TaskMetaData{ID: 38, Type: taskRedPacket, Total: 1, Desc: "普通场打出1个飞机", Jump: basescoreMatching}   // 普通场打出1个飞机
	TaskList[39] = &TaskMetaData{ID: 39, Type: taskRedPacket, Total: 2, Desc: "普通场打出2个连对", Jump: basescoreMatching}   // 普通场打出2个连对
	TaskList[40] = &TaskMetaData{ID: 40, Type: taskRedPacket, Total: 2, Desc: "普通场打出2个炸弹", Jump: basescoreMatching}   // 普通场打出2个炸弹
	TaskList[41] = &TaskMetaData{ID: 41, Type: taskRedPacket, Total: 2, Desc: "普通场打出2个王炸", Jump: basescoreMatching}   // 普通场打出2个王炸
	TaskList[42] = &TaskMetaData{ID: 42, Type: taskRedPacket, Total: 6, Desc: "普通场打出6个顺子", Jump: basescoreMatching}   // 普通场打出6个顺子
	TaskList[43] = &TaskMetaData{ID: 43, Type: taskRedPacket, Total: 8, Desc: "普通场加倍底分8次", Jump: basescoreMatching}   // 普通场加倍底分8次
	TaskList[44] = &TaskMetaData{ID: 44, Type: taskRedPacket, Total: 8, Desc: "普通场当地主8次", Jump: basescoreMatching}    // 普通场当地主8次
	TaskList[45] = &TaskMetaData{ID: 45, Type: taskRedPacket, Total: 6, Desc: "普通场叫地主6次", Jump: basescoreMatching}    // 普通场叫地主6次
	TaskList[46] = &TaskMetaData{ID: 46, Type: taskRedPacket, Total: 6, Desc: "打出6个炸弹", Jump: basescoreMatching}      // 打出6个炸弹
	TaskList[47] = &TaskMetaData{ID: 47, Type: taskRedPacket, Total: 5, Desc: "打出5个王炸", Jump: basescoreMatching}      // 打出5个王炸
	TaskList[48] = &TaskMetaData{ID: 48, Type: taskRedPacket, Total: 10, Desc: "以地主身份获胜10局", Jump: basescoreMatching} // 以地主身份获胜10局
	TaskList[49] = &TaskMetaData{ID: 49, Type: taskRedPacket, Total: 12, Desc: "以农民身份获胜12局", Jump: basescoreMatching} // 以农民身份获胜12局
	TaskList[50] = &TaskMetaData{ID: 50, Type: taskRedPacket, Total: 5, Desc: "连胜5局", Jump: basescoreMatching}        // 连胜5局
	TaskList[51] = &TaskMetaData{ID: 51, Type: taskRedPacket, Total: 3, Desc: "单局打出2个顺子3次", Jump: basescoreMatching}  // 单局打出2个顺子3次
	TaskList[52] = &TaskMetaData{ID: 52, Type: taskRedPacket, Total: 8, Desc: "打出8次三带二", Jump: basescoreMatching}     // 打出8次三带二
	TaskList[53] = &TaskMetaData{ID: 53, Type: taskRedPacket, Total: 3, Desc: "打出3次四带二", Jump: basescoreMatching}     // 打出3次四带二
	TaskList[54] = &TaskMetaData{ID: 54, Type: taskRedPacket, Total: 15, Desc: "当地主15次", Jump: basescoreMatching}     // 当地主15次
	TaskList[55] = &TaskMetaData{ID: 55, Type: taskRedPacket, Total: 2, Desc: "普通场打出2个飞机", Jump: basescoreMatching}   // 普通场打出2个飞机
	TaskList[56] = &TaskMetaData{ID: 56, Type: taskRedPacket, Total: 4, Desc: "普通场打出4个连对", Jump: basescoreMatching}   // 普通场打出4个连对
	TaskList[57] = &TaskMetaData{ID: 57, Type: taskRedPacket, Total: 3, Desc: "普通场打出3个炸弹", Jump: basescoreMatching}   // 普通场打出3个炸弹
	TaskList[58] = &TaskMetaData{ID: 58, Type: taskRedPacket, Total: 10, Desc: "普通场打出10个顺子", Jump: basescoreMatching} // 普通场打出10个顺子
	TaskList[59] = &TaskMetaData{ID: 59, Type: taskRedPacket, Total: 6, Desc: "普通场地主身份获胜6局", Jump: basescoreMatching} // 普通场地主身份获胜6局
	TaskList[60] = &TaskMetaData{ID: 60, Type: taskRedPacket, Total: 8, Desc: "普通场农民身份获胜8局", Jump: basescoreMatching} // 通场农民身份获胜8局
	TaskList[61] = &TaskMetaData{ID: 61, Type: taskRedPacket, Total: 4, Desc: "单局打出2个顺子4次", Jump: basescoreMatching}  // 单局打出2个顺子4次
	TaskList[62] = &TaskMetaData{ID: 62, Type: taskRedPacket, Total: 5, Desc: "单局打出2个顺子5次", Jump: basescoreMatching}  // 单局打出2个顺子5次
	TaskList[63] = &TaskMetaData{ID: 63, Type: taskRedPacket, Total: 8, Desc: "普通场打出8次三带二", Jump: basescoreMatching}  // 普通场打出8次三带二

	// 金币任务
	TaskList[1000] = &TaskMetaData{ID: 1000, Type: taskChip, Total: 1, Chips: 100, Desc: "登录游戏"}                             // 成功登录游戏，奖励100金币
	TaskList[1001] = &TaskMetaData{ID: 1001, Type: taskChip, Total: 1, Chips: 2000, Desc: "微信分享游戏", Jump: wechatShare}       // 微信分享游戏，奖励2000金币
	TaskList[1002] = &TaskMetaData{ID: 1002, Type: taskChip, Total: 3, Chips: 2000, Desc: "胜利3局", Jump: basescoreMatching}   // 累计胜利3局，奖励2000金币
	TaskList[1003] = &TaskMetaData{ID: 1003, Type: taskChip, Total: 5, Chips: 3000, Desc: "胜利5局", Jump: basescoreMatching}   // 累计胜利5局，奖励3000金币
	TaskList[1004] = &TaskMetaData{ID: 1004, Type: taskChip, Total: 6, Chips: 3000, Desc: "胜利6局", Jump: basescoreMatching}   // 累计胜利6局，奖励3000金币
	TaskList[1005] = &TaskMetaData{ID: 1005, Type: taskChip, Total: 5, Chips: 2000, Desc: "对局5局", Jump: basescoreMatching}   // 累计对局5局，奖励2000金币
	TaskList[1006] = &TaskMetaData{ID: 1006, Type: taskChip, Total: 10, Chips: 3000, Desc: "对局10局", Jump: basescoreMatching} // 累计对局10局，奖励3000金币
	TaskList[1007] = &TaskMetaData{ID: 1007, Type: taskChip, Total: 15, Chips: 5000, Desc: "对局15局", Jump: basescoreMatching} // 累计对局15局，奖励5000金币
	TaskList[1008] = &TaskMetaData{ID: 1008, Type: taskChip, Total: 2, Chips: 3000, Desc: "连胜2局", Jump: basescoreMatching}   // 任意场连胜2局，奖励3000金币
	TaskList[1009] = &TaskMetaData{ID: 1009, Type: taskChip, Total: 3, Chips: 5000, Desc: "连胜3局", Jump: basescoreMatching}   // 任意场连胜3局，奖励5000金币
	TaskList[1010] = &TaskMetaData{ID: 1010, Type: taskChip, Total: 1, Chips: 2000, Desc: "购买任意数量的金币", Jump: shop}           // 购买任意数量金币，奖励2000金币
	TaskList[1011] = &TaskMetaData{ID: 1011, Type: taskChip, Total: 3, Chips: 3000, Desc: "打出3个飞机", Jump: basescoreMatching} // 累计打出3个飞机，奖励3000金币
	TaskList[1012] = &TaskMetaData{ID: 1012, Type: taskChip, Total: 2, Chips: 3000, Desc: "打出2个王炸", Jump: basescoreMatching} // 累计打出2个王炸，奖励3000金币
	TaskList[1013] = &TaskMetaData{ID: 1013, Type: taskChip, Total: 3, Chips: 3000, Desc: "打出3个炸弹", Jump: basescoreMatching} // 累计打出3个炸弹，奖励3000金币
	TaskList[1014] = &TaskMetaData{ID: 1014, Type: taskChip, Total: 2, Chips: 2000, Desc: "明牌开始2次", Jump: basescoreMatching} // 累计明牌开始2次，奖励2000金币
	TaskList[1015] = &TaskMetaData{ID: 1015, Type: taskChip, Total: 3, Chips: 3000, Desc: "明牌开始3次", Jump: basescoreMatching} // 累计明牌开始3次，奖励3000金币
	TaskList[1016] = &TaskMetaData{ID: 1016, Type: taskChip, Total: 1, Chips: 5000, Desc: "参加一次红包比赛场", Jump: redpacketMatch} // 参加一次红包比赛，奖励5000金币

	// 活动任务（必须有活动时间表）
	TaskList[1017] = &TaskMetaData{ID: 1017, Type: taskChip, Total: 2, Chips: 18888, Desc: "活动期间成功邀请2位好友", Jump: wechatInvite} // 成功邀请2位好友，奖励9999金币
	ActivityTimeList[1017] = &ActivityTaskSchedule{
		Start:    time.Date(2018, 5, 9, 0, 0, 0, 0, time.Local).Unix(),
		End:      time.Date(2018, 5, 18, 0, 0, 0, 0, time.Local).Unix(),
		Deadline: time.Date(2018, 5, 19, 20, 0, 0, 0, time.Local).Unix(),
	}

	TaskList[1018] = &TaskMetaData{ID: 1018, Type: taskChip, Total: 18, Chips: 1888, Desc: "活动期间完成18次红包比赛", Jump: redpacketMatch} // 活动期间完成18次红包比赛，奖励999金币
	ActivityTimeList[1018] = &ActivityTaskSchedule{
		Start:    time.Date(2018, 4, 28, 0, 0, 0, 0, time.Local).Unix(),
		End:      time.Date(2018, 5, 6, 0, 0, 0, 0, time.Local).Unix(),
		Deadline: time.Date(2018, 5, 7, 20, 0, 0, 0, time.Local).Unix(),
	}
	for _, schedule := range conf.GetCfgActivityTimes() {
		ActivityTimeList[schedule.TaskID] = &ActivityTaskSchedule{
			Start:    schedule.Start,
			End:      schedule.End,
			Deadline: schedule.Deadline,
		}
	}
}

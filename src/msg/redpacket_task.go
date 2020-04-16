package msg

func init() {
	Processor.Register(&S2C_RedpacketTask{})
	Processor.Register(&C2S_GetRedpacketTaskCode{})
	Processor.Register(&S2C_RedpacketTaskCode{})
}

type S2C_RedpacketTask struct {
	Tasks          []RedPacketTask
	Chips          int64
	FreeChangeTime int64
}
type RedPacketTask struct {
	ID            int     //任务Id
	Real          bool    `json:"-"` //true表示指定金额红包 false表示幸运红包
	Total         int     //总次数
	Fee           float64 //指定红包金额的数值
	Desc          string  //红包任务描述内容
	PlayTimes     int     //当前任务执行的次数
	StartTime     int64   //任务的开始时间
	RedPacketCode string  //红包码
	State         int     //状态 0 表示未进行 1 表示进行中 2 领取
	Jump          int     // 0 调到当前游戏的金币场  10 调到当前游戏的创建房间
	Type          int     //红包类型 1:新人红包 2:指定金额 3:幸运红包
}

type C2S_GetRedPacketTaskRecord struct {
}
type RedpacketTaskRecord struct {
	ID            int     //任务Id
	Real          bool    `json"-"` //true表示指定金额红包 false表示幸运红包
	Fee           float64 //指定红包金额的数值
	Desc          string  //红包任务描述内容
	Createdat     int64   //红包领取的时间
	RedPacketCode string  //红包码
	Type          int     //红包类型 1:新人红包 2:指定金额 3:幸运红包
}

type S2C_RedPacketTaskRecord struct {
	TaskRecords []RedpacketTaskRecord
}
type C2S_ChangeTask struct {
	Free bool
}
type S2C_ChangeTask struct {
	Error int
}

const (
	S2C_ChangeTaskSuccess       = 0 //更换任务成功
	S2C_ChipsLack               = 1 //金币不足已更换任务
	S2C_NoTaskChange            = 2 //已无任务可以更换
	S2C_ChangeTask_NotReachTime = 3 //24小时后才可更换
	S2C_NewPlayer_NotChange     = 4 //新人任务必须完成
)

type C2S_GetRedpacketTaskCode struct {
	TaskId int
}

const (
	S2C_GetRedpacketTaskCodeSuccess = 0 //成功
	S2C_RedpacketTaskInValid        = 1 //获取红包码失败,稍后再试
)

type S2C_RedpacketTaskCode struct {
	Error int
	Code  string
	Desc  string
}

package msg

func init() {
	Processor.Register(&S2C_ShareTasksInfo{})
	Processor.Register(&C2S_GetShareTaskInfo{})
	Processor.Register(&C2S_BindSharer{})
	Processor.Register(&S2C_BindSharer{})
	Processor.Register(&C2S_ShareRecord{})
	Processor.Register(&S2C_ShareRecord{})
	Processor.Register(&C2S_CopyExchangeCode{})
	Processor.Register(&S2C_CopyExchangeCode{})
	Processor.Register(&C2S_Achievement{})
	Processor.Register(&S2C_Achievement{})
	Processor.Register(&C2S_AbleProfit{})
	Processor.Register(&S2C_AbleProfit{})
	Processor.Register(&C2S_AgentNumbersProfit{})
	Processor.Register(&S2C_AgentNumbersProfit{})
	Processor.Register(&C2S_ReceiveShareProfit{})
	Processor.Register(&S2C_ReceiveShareProfit{})
	Processor.Register(&C2S_TakenProfit{})
	Processor.Register(&S2C_TakenProfit{})
}

//任务列表
type C2S_GetShareTaskInfo struct {
}
type S2C_ShareTasksInfo struct {
	ShareTasks []*ShareTasks //分享任务列表
}
type ShareTasks struct {
	FeeDes string //奖励
	Desc   string //任务描述
}

//绑定邀请人
type C2S_BindSharer struct {
	AccountID string //绑定玩家的邀请码
}

const (
	BindSharerOK       = 0 //成功
	BindSharerAbnormal = 1 //格式错误，请重新绑定
	BindShareLate      = 2 //邀请人注册时间较晚，请绑定其他推荐码
	BindNotToLevel     = 3 //下级不能绑定到上级
	BindRobot          = 4 //绑定目标是机器人或者自己是机器人
	BindDuplicate      = 5 //已绑定
	BindSelf           = 6 //绑定自己
)

type S2C_BindSharer struct {
	Error    int   //详见错误码描述
	ParentId int64 //（绑定成功后)更新玩家的绑定码
}

//获取领取记录
type C2S_ShareRecord struct {
	Page int //页码
	Per  int //条数
}

type S2C_ShareRecord struct {
	Page              int                 //页码
	Per               int                 //条数
	Total             int                 //总数
	ShareTakenRecords []*ShareTakenRecord //数据
}

type ShareTakenRecord struct {
	ID           int     //唯一标识
	Fee          float64 //奖励金额
	Level        int     //来源代理层级
	DirID        int     `json:"-"` //目标人
	CreatedAt    int64   //创建时间
	IsCopy       bool    //是否已复制
	ExchangeCode string  //兑换码
}

//累计总收入
type C2S_TakenProfit struct {
}

type S2C_TakenProfit struct {
	TakenProfit float64 //累计总收入
}

//复制兑换码
type C2S_CopyExchangeCode struct {
	ShareRecordID int //要复制的记录id
}

const (
	CopyOK   = 0 //复制操作成功
	CopyFail = 1 //复制操作失败
)

type S2C_CopyExchangeCode struct {
	Error         int //错误码
	ShareRecordID int //要复制的记录id
}

//业绩详情列表
type C2S_Achievement struct {
	Level int //代理级别
	Page  int //页码
	Per   int //条数
}

type Achievement struct {
	Createdat int64   `json:"-"` //创建时间
	AccountId int64   //用户Id
	Recharge  float64 //累计充值金额
	AllProfit float64 //贡献总收益
	Profit    float64 //可领取收益
	Updatedat int64   `json:"-"` //更新时间
}
type S2C_Achievement struct {
	Page         int           //页码
	Per          int           //条数
	Total        int           //总数
	Achievements []Achievement //业绩数据
}

//可领取收益
type C2S_AbleProfit struct {
	Level int //代理级别
}

type S2C_AbleProfit struct {
	AbleProfit float64 //可领取收益
}

//代理人数和对应的总收益
type C2S_AgentNumbersProfit struct {
}

type S2C_AgentNumbersProfit struct {
	NumbersProfits []NumbersProfit //代理人数和对应的总收益数据
}

type NumbersProfit struct {
	Level  int     //代理级别
	Number int     //数量
	Profit float64 //可领取收益
}

//领取总收益
type C2S_ReceiveShareProfit struct {
	Level int //代理级别
}

const (
	ReceiveOK   = 0 //成功
	ReceiveFail = 1 //失败
)

type S2C_ReceiveShareProfit struct {
	Error        int     //错误码
	Profit       float64 //领取成功的收益
	ExchangeCode string  //兑换码
}

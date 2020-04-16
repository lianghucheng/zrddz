package msg

const (
	S2C_CreateRoom_InnerError    = 1 // 创建房间出错，请稍后重试
	S2C_CreateRoom_InOtherRoom   = 3 // 正在其他房间对局，是否回去？
	S2C_CreateRoom_RuleError     = 5 // 规则错误，请稍后重试
	S2C_CreateRoom_LackOfChips   = 6 // 需要 S2C_CreateRoom.MinChips 筹码才能游戏，请先购买筹码
	S2C_CreateRoom_MaxChipsLimit = 7 //
)

type S2C_CreateRoom struct {
	Error    int
	MinChips int64 // 最小筹码
	MaxChips int64 // 最大筹码
}

type C2S_EnterRoom struct {
	RoomNumber string
}

const (
	S2C_EnterRoom_OK            = 0
	S2C_EnterRoom_NotCreated    = 1 // "房间: " + S2C_EnterRoom.RoomNumber + " 未创建"
	S2C_EnterRoom_Full          = 2 // "房间: " + S2C_EnterRoom.RoomNumber + " 玩家人数已满"
	S2C_EnterRoom_Unknown       = 4 // 进入房间出错，请稍后重试
	S2C_EnterRoom_IPConflict    = 5 // IP重复，无法进入
	S2C_EnterRoom_LackOfChips   = 6 // 需要 + S2C_EnterRoom.MinChips + 筹码才能进入，请先购买筹码
	S2C_EnterRoom_NotRightNow   = 7 // 比赛暂未开始，请到时再来
	S2C_EnterRoom_MaxChipsLimit = 8 //
)

type S2C_EnterRoom struct {
	Error         int
	RoomType      int // 房间类型: 0 练习、1 底分匹配、2 底分私人、3 VIP私人、4 红包匹配、5 红包私人
	RoomNumber    string
	Position      int
	BaseScore     int
	RedPacketType int // 红包种类(元): 1、10、100、999
	RoomDesc      string
	MaxPlayers    int   // 最大玩家数
	MinChips      int64 // 最小筹码
	MaxChips      int64 // 最大筹码
	Tickects      int64 // 门票金额
	GamePlaying   bool  // 游戏是否进行中
}

type C2S_GetAllPlayers struct{}

type S2C_SitDown struct {
	Position   int
	AccountID  int
	LoginIP    string
	Nickname   string
	Headimgurl string
	Sex        int
	Owner      bool
	Ready      bool
	Chips      int64
}

type S2C_StandUp struct {
	Position int
}

type C2S_ExitRoom struct{}

const (
	S2C_ExitRoom_OK          = 0
	S2C_ExitRoom_GamePlaying = 1 // 游戏进行中，不能退出房间
)

type S2C_ExitRoom struct {
	Error    int
	Position int
}

// 更新座位上的玩家筹码
type S2C_UpdatePlayerChips struct {
	Position int
	Chips    int64
}

// 系统托管
type C2S_SystemHost struct {
	Host bool
}

type S2C_SystemHost struct {
	Position int
	Host     bool
}

type C2S_SetVIPRoomChips struct {
	Chips int64
}

const (
	S2C_SetVIPChips_OK          = 0
	S2C_SetVIPChips_ChipsUnset  = 1 // (前端显示输入筹码界面)
	S2C_SetVIPChips_LackOfChips = 2 // 筹码不足，请先购买筹码
)

type S2C_SetVIPRoomChips struct {
	Error    int
	Position int
	Chips    int64
}

type C2S_ChangeTable struct{}

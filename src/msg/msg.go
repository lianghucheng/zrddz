package msg

import (
	"github.com/name5566/leaf/network/json"
	"gopkg.in/mgo.v2/bson"
)

var Processor = json.NewProcessor()

func init() {
	Processor.Register(&C2S_SetSystemOn{})
	Processor.Register(&C2S_SetLandlordConfig{})
	Processor.Register(&C2S_SetUsernamePassword{})
	Processor.Register(&C2S_SetUserRole{})
	Processor.Register(&C2S_TransferChips{})
	Processor.Register(&C2S_WeChatLogin{})
	Processor.Register(&C2S_TokenLogin{})
	Processor.Register(&C2S_UsernamePasswordLogin{})
	Processor.Register(&C2S_Heartbeat{})
	Processor.Register(&C2S_CreateLandlordRoom{})
	Processor.Register(&C2S_EnterRoom{})
	Processor.Register(&C2S_GetUserChips{})
	Processor.Register(&C2S_SetVIPRoomChips{})
	Processor.Register(&C2S_GetCheckInDetail{})
	Processor.Register(&C2S_CheckIn{})
	Processor.Register(&C2S_GetAllPlayers{})
	Processor.Register(&C2S_ExitRoom{})
	Processor.Register(&C2S_LandlordPrepare{}) //准备游戏
	Processor.Register(&C2S_LandlordMatching{})
	Processor.Register(&C2S_LandlordBid{})
	Processor.Register(&C2S_LandlordGrab{})
	Processor.Register(&C2S_LandlordDouble{})
	Processor.Register(&C2S_LandlordShowCards{})
	Processor.Register(&C2S_LandlordDiscard{})
	Processor.Register(&C2S_GetMonthChipsRank{})
	Processor.Register(&C2S_GetMonthChipsRankPos{})
	Processor.Register(&C2S_GetMonthWinsRank{})
	Processor.Register(&C2S_GetMonthWinsRankPos{})
	Processor.Register(&C2S_CleanMonthRanks{})
	Processor.Register(&C2S_SystemHost{})
	Processor.Register(&C2S_ChangeTable{})
	Processor.Register(&C2S_GetRedPacketMatchRecord{})
	Processor.Register(&C2S_TakeRedPacketMatchPrize{})
	Processor.Register(&C2S_DoTask{})
	Processor.Register(&C2S_ChangeTask{})
	Processor.Register(&C2S_FreeChangeCountDown{})
	Processor.Register(&C2S_TakeTaskPrize{})
	Processor.Register(&C2S_FakeWXPay{})
	Processor.Register(&C2S_FakeAliPay{})
	Processor.Register(&C2S_GetCircleLoginCode{})

	Processor.Register(&S2C_SetSystemOn{})
	Processor.Register(&S2C_SetLandlordConfig{})
	Processor.Register(&S2C_SetUserRole{})
	Processor.Register(&S2C_TransferChips{})
	Processor.Register(&S2C_UpdateUserChips{})
	Processor.Register(&S2C_SetVIPRoomChips{})
	Processor.Register(&S2C_CheckInDetail{})
	Processor.Register(&S2C_UpdateCheckInDetail{})
	Processor.Register(&S2C_UpdateNotice{})
	Processor.Register(&S2C_UpdateRadio{})
	Processor.Register(&S2C_Close{})
	Processor.Register(&S2C_Login{})
	Processor.Register(&S2C_Heartbeat{})
	Processor.Register(&S2C_CreateRoom{})
	Processor.Register(&S2C_EnterRoom{})
	Processor.Register(&S2C_SitDown{})
	Processor.Register(&S2C_StandUp{})
	Processor.Register(&S2C_ExitRoom{})
	Processor.Register(&S2C_Prepare{})
	Processor.Register(&S2C_GameStart{})
	Processor.Register(&S2C_UpdatePokerHands{})
	Processor.Register(&S2C_ActionLandlordBid{})
	Processor.Register(&S2C_LandlordBid{})
	Processor.Register(&S2C_ActionLandlordGrab{})
	Processor.Register(&S2C_LandlordGrab{})
	Processor.Register(&S2C_DecideLandlord{})
	Processor.Register(&S2C_UpdateLandlordLastThree{})
	Processor.Register(&S2C_ActionLandlordDouble{})
	Processor.Register(&S2C_LandlordDouble{})
	Processor.Register(&S2C_ActionLandlordShowCards{})
	Processor.Register(&S2C_LandlordShowCards{})
	Processor.Register(&S2C_UpdateLandlordMultiple{})
	Processor.Register(&S2C_ActionLandlordDiscard{})
	Processor.Register(&S2C_LandlordDiscard{})
	Processor.Register(&S2C_ClearAction{})
	Processor.Register(&S2C_LandlordSpring{})
	Processor.Register(&S2C_LandlordRoundResult{})
	Processor.Register(&S2C_UpdateMonthChipsRankPos{})
	Processor.Register(&S2C_UpdateMonthChipsRanks{})
	Processor.Register(&S2C_UpdateMonthWinsRankPos{})
	Processor.Register(&S2C_UpdateMonthWinsRanks{})
	Processor.Register(&S2C_CleanMonthRanks{})
	Processor.Register(&S2C_UpdatePlayerChips{})
	Processor.Register(&S2C_SystemHost{})
	Processor.Register(&S2C_RedPacketMatchRecord{})
	Processor.Register(&S2C_TakeRedPacketMatchPrize{})
	Processor.Register(&S2C_UpdateUntakenRedPacketMatchPrizeNumber{})
	Processor.Register(&S2C_UpdateRedPacketMatchOnlineNumber{})
	Processor.Register(&S2C_UpdateRedPacketTaskList{})
	Processor.Register(&S2C_UpdateChipTaskList{})
	Processor.Register(&S2C_UpdateTaskProgress{})
	Processor.Register(&S2C_TakeTaskPrize{})
	Processor.Register(&S2C_ChangeTask{})
	Processor.Register(&S2C_FreeChangeCountDown{})
	Processor.Register(&S2C_PayOK{})
	Processor.Register(&S2C_UpdateCircleLoginCode{})

	Processor.Register(&C2S_SetRobotData{})

	Processor.Register(&S2C_OfferSubsidy{})
	Processor.Register(&C2S_GetCardMa{})
	Processor.Register(&S2C_CardMa{})
	Processor.Register(&S2C_RedPacketTaskRecord{})
	Processor.Register(&C2S_CardCodeState{})
	Processor.Register(&C2S_TakeTaskState{})
	Processor.Register(&S2C_CircleLink{})
	Processor.Register(&C2S_SubsidyChip{})
	Processor.Register(&S2C_SubsidyChip{})
}

type C2S_Heartbeat struct{}

type S2C_Heartbeat struct{}

type C2S_TransferChips struct {
	AccountID int
	Chips     int64
}

const (
	S2C_TransferChips_OK               = 0
	S2C_TransferChips_AccountIDInvalid = 1 // 账号ID无效
	S2C_TransferChips_NotYourself      = 2 // 不能转给自己
	S2C_TransferChips_ChipsInvalid     = 3 // 筹码数量 + C2S_TransferChip.Chips + 无效
	S2C_TransferChips_PermissionDenied = 4 // 没有权限
)

type S2C_TransferChips struct {
	Error int
	Chips int64
}

type C2S_GetCheckInDetail struct {
}

type C2S_CheckIn struct {
}

type S2C_CheckInDetail struct {
	CheckIn    bool
	CheckInSum int
}

type S2C_UpdateCheckInDetail struct {
	CheckIn    bool
	CheckInSum int
}

type S2C_Prepare struct {
	Position int
	Ready    bool
}

type S2C_GameStart struct{}

type S2C_UpdatePokerHands struct {
	Position      int
	Hands         []int // 手牌
	NumberOfHands int   // 手牌数量
	ShowCards     bool  // 明牌
}

// 获取红包比赛记录
type C2S_GetRedPacketMatchRecord struct {
	PageNumber int // 页码数
	PageSize   int // 一页显示的条数
}

type RedPacketMatchRecordItem struct {
	ID            bson.ObjectId
	RedPacketType int
	RedPacket     float64
	Taken         bool
	Date          string
	CardCode      string //红包码
}

type S2C_RedPacketMatchRecord struct {
	Items      []RedPacketMatchRecordItem
	Total      int // 总数
	PageNumber int // 页码数
	PageSize   int // 一页显示的条数
}

// 领取红包比赛奖励
type C2S_TakeRedPacketMatchPrize struct {
	ID bson.ObjectId
}

const (
	S2C_TakeRedPacketMatchPrize_OK              = 0 // 恭喜领取 S2C_TakeRedPacketMatchPrize.RedPacket元红包奖励，请至“圈圈”查看
	S2C_TakeRedPacketMatchPrize_IDInvalid       = 1 // 比赛记录ID无效
	S2C_TakeRedPacketMatchPrize_NotYetWon       = 2 // 离获奖还差一点点，请继续努力吧
	S2C_TakeRedPacketMatchPrize_TakeRepeated    = 3 // S2C_TakeRedPacketMatchPrize.RedPacket元红包奖励已被领取，请勿重复操作
	S2C_TakeRedPacketMatchPrize_CircleIDInvalid = 4 // 圈圈ID无效
	S2C_TakeRedPacketMatchPrize_Error           = 5 // 领取出错，请稍后重试
)

// 领取红包比赛奖励
type S2C_TakeRedPacketMatchPrize struct {
	Error     int
	ID        bson.ObjectId
	RedPacket float64
}

// 更新未领取的红包比赛奖励数量
type S2C_UpdateUntakenRedPacketMatchPrizeNumber struct {
	Number int
}

// 更新红包比赛在线人数
type S2C_UpdateRedPacketMatchOnlineNumber struct {
	Numbers []int
}

type C2S_FakeWXPay struct {
	TotalFee int
}

type C2S_FakeAliPay struct {
	TotalAmount float64
}

// 购买S2C_PayOK.Chips金币成功
type S2C_PayOK struct {
	Chips int64
}

type C2S_GetCircleLoginCode struct{}

const (
	S2C_UpdateCircleLoginCode_OK    = 0
	S2C_UpdateCircleLoginCode_Error = 1 // 圈圈授权出错，请稍后重试
)

type S2C_UpdateCircleLoginCode struct {
	Error     int
	LoginCode string
}

// robot
type C2S_SetRobotData struct {
	LoginIP string
	Chips   int64
}
type C2S_CardCodeState struct {
}

const (
	SubsidyOK = 0
	SubsidyMore = 1
	SubsidyNotLack = 3
)
type S2C_SubsidyChip struct {
	Error   int
	Chip 	int
}

type C2S_SubsidyChip struct {
	Reply 	bool
}
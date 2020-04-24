package msg

import (
	"game/poker"
)

type C2S_SetLandlordConfig struct {
	AndroidVersion     int    // Android 版本号
	AndroidDownloadUrl string // Android 下载链接
	IOSVersion         int    // iOS 版本号
	IOSDownloadUrl     string // iOS 下载链接
	SougouVersion	   int    // Sougou 版本号
	SougouDownloadUrl  string // Sougou 下载链接
	AndroidGuestLogin  bool   // Android 游客登录
	IOSGuestLogin      bool   // iOS 游客登录
	SougouGuestLogin   bool
	Notice             string // 公告
	Radio              string // 广播
	WeChatNumber       string // 客服微信号
	EnterAddress       bool   // 输入服务器地址
}

const (
	S2C_SetLandlordConfig_OK                  = 0
	S2C_SetLandlordConfig_PermissionDenied    = 1 // 没有权限
	S2C_SetLandlordConfig_VersionInvalid      = 2 // 版本 + S2C_SetLandlordConfig.AndroidVersion + 无效
	S2C_SetLandlordConfig_DownloadUrlInvalid  = 3 // 下载地址 + S2C_SetLandlordConfig.AndroidDownloadUrl + 无效
	S2C_SetLandlordConfig_WeChatNumberInvalid = 4 // 客服微信号 + S2C_SetLandlordConfig.WeChatNumberOfCustomerService + 无效
)

type S2C_SetLandlordConfig struct {
	Error              int
	AndroidVersion     int    // Android 版本号
	AndroidDownloadUrl string // Android 下载链接
	IOSVersion         int    // iOS 版本号
	IOSDownloadUrl     string // iOS 下载链接
	Notice             string // 公告
	Radio              string // 广播
	WeChatNumber       string // 客服微信号
}

type C2S_CreateLandlordRoom struct {
	poker.LandlordRule
}

type C2S_LandlordPrepare struct {
	ShowCards bool
}

type C2S_LandlordMatching struct {
	RoomType      int // 房间类型: 0 练习、1 底分匹配、4 红包匹配
	BaseScore     int // 底分: 100、5000、1万
	RedPacketType int // 红包种类(元): 1、10、100、999
}

// 叫地主动作
type S2C_ActionLandlordBid struct {
	Position  int
	Countdown int // 倒计时
}

type C2S_LandlordBid struct {
	Bid bool
}

type S2C_LandlordBid struct {
	Position int
	Bid      bool
}

// 抢地主动作
type S2C_ActionLandlordGrab struct {
	Position  int
	Countdown int // 倒计时
}

type C2S_LandlordGrab struct {
	Grab bool
}

type S2C_LandlordGrab struct {
	Position int
	Grab     bool
	Again    bool
}

// 加倍动作（只发给自己）
type S2C_ActionLandlordDouble struct {
	Countdown int // 倒计时
}

type C2S_LandlordDouble struct {
	Double bool
}

type S2C_LandlordDouble struct {
	Position int
	Double   bool
}

// 明牌动作（只发给自己）
type S2C_ActionLandlordShowCards struct {
	Countdown int // 倒计时
}

type C2S_LandlordShowCards struct {
	ShowCards bool
}

type S2C_LandlordShowCards struct {
	Position int
}

type S2C_DecideLandlord struct {
	Position int
}

type S2C_UpdateLandlordLastThree struct {
	Cards []int
}

type S2C_UpdateLandlordMultiple struct {
	Multiple int
}

// 出牌动作
type S2C_ActionLandlordDiscard struct {
	ActionDiscardType int // 出牌动作类型
	Position          int
	Countdown         int     // 倒计时
	PrevDiscards      []int   // 上一次出的牌
	Hint              [][]int // 出牌提示
}

type C2S_LandlordDiscard struct {
	Cards []int
}

type S2C_LandlordDiscard struct {
	Position int
	Cards    []int
}

type S2C_ClearAction struct{} // 清除动作

type S2C_LandlordSpring struct{}

// 单局成绩
type S2C_LandlordRoundResult struct {
	Result       int // 0 失败、1 胜利
	RoomDesc     string
	Spring       bool
	RoundResults []poker.LandlordPlayerRoundResult
	ContinueGame bool // 是否继续游戏
}

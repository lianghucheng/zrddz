package msg

type C2S_WeChatLogin struct {
	Nickname   string
	Headimgurl string
	Sex        int    // 1为男性，2为女性
	Serial     string // 安卓设备硬件序列号,例如:a1113028
	Model      string // 安卓手机型号,例如:MI NOTE Pro
	UnionID    string // 微信unionid
	Channel    int    //渠道号。0：圈圈   1：搜狗   2:IOS  3:GooglePlay
}

type C2S_TokenLogin struct {
	Token string
}

type C2S_UsernamePasswordLogin struct {
	Username string
	Password string
}

// Close
const (
	S2C_Close_LoginRepeated   = 1 // 您的账号在其他设备上线，非本人操作请注意修改密码
	S2C_Close_InnerError      = 2 // 登录出错，请重新登录
	S2C_Close_TokenInvalid    = 3 // 登录状态失效，请重新登录
	S2C_Close_UnionIDInvalid  = 4 // 登录出错，微信ID无效
	S2C_Close_UsernameInvalid = 5 // 登录出错，用户名无效
	S2C_Close_SystemOff       = 6 // 系统升级维护中，请稍后重试
	S2C_Close_RoleBlack       = 7 // 账号已冻结，请联系客服微信 S2C_Close.WeChatNumber
	S2C_Close_IPChanged       = 8 // 登录IP发生变化，非本人操作请注意修改密码
)

type S2C_Close struct {
	Error        int
	WeChatNumber string
}

type S2C_Login struct {
	AccountID       int
	Nickname        string
	Headimgurl      string
	Sex             int // 1 男、2 女
	Role            int // 1 玩家、2 代理、3 管理员、4 超管
	Token           string
	AnotherLogin    bool   // 其他设备登录
	AnotherRoom     bool   // 在其他房间
	FirstLogin      bool   // 首次登录
	Radio           string // 广播
	Parentid        int64  // 上级ID
	WeChatNumber    string // 客服微信号
	CardCode        string // 取牌码
	Taken           bool   // 已领取或未领取
	CardCodeDesc    string // 取牌码的描述
	PlayTimes       int    // 当天游戏次数
	Total           int    // 取牌码任务总次数
	GivenChips      int    // 绑定赠送的金币数量
	FirstLoginChips int    // 首次登录送的金币
}

type C2S_SetUsernamePassword struct {
	Username string
	Password string
}

type C2S_GetUserChips struct{}

type S2C_UpdateUserChips struct {
	Chips int64
}

type (
	S2C_OfferSubsidy struct { // 发放补助
		Chips int64
	}
)
type C2S_GetCardMa struct {
}
type S2C_CardMa struct {
	Code      string //	取牌码
	Total     int    // 总进度
	PlayTimes int    // 已完成进度
	Completed bool   // true表示去复制，显示取牌码
}

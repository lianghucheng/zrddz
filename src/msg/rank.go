package msg

// 金币排行榜
type MonthChipsRank struct {
	UserID     int
	Nickname   string
	Headimgurl string
	Chips      int64
}

type TempMCR struct {
	Nickname   string
	Headimgurl string
	Chips      int64
}

// 胜场排行榜
type MonthWinsRank struct {
	UserID     int
	Nickname   string
	Headimgurl string
	Wins       int
}

type TempMWR struct {
	Nickname   string
	Headimgurl string
	Wins       int
}

type C2S_GetMonthChipsRank struct {
	PageNum int
}

type C2S_GetMonthChipsRankPos struct {
}

type C2S_GetMonthWinsRank struct {
	PageNum int
}

type C2S_GetMonthWinsRankPos struct {
}

type C2S_CleanMonthRanks struct {
}

type S2C_UpdateMonthChipsRankPos struct {
	Pos   int
	Chips int64
}

type S2C_UpdateMonthChipsRanks struct {
	PageNum    int // 当前页数
	PageSum    int // 总页数
	ChipsRanks []TempMCR
}

type S2C_UpdateMonthWinsRankPos struct {
	Pos  int
	Wins int
}

type S2C_UpdateMonthWinsRanks struct {
	PageNum   int // 当前页数
	PageSum   int // 总页数
	WinsRanks []TempMWR
}

type S2C_CleanMonthRanks struct {
}

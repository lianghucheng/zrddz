package internal

import (
	"github.com/name5566/leaf/log"
	"time"
)

type GameRecord struct {
	AccountId      int          //用户Id
	Desc           string       //房间描述
	RoomNumber     string       //房间号
	Profit         int64        //资金变化
	Amount         int64        //当前余额
	StartTimestamp int64        // 开始时间
	EndTimestamp   int64        // 结束时间
	Results        []ResultData //对战详情
	Nickname       string       // 昵称
	IsSpring		bool        //这一局是否为春天
	LastThree	   []string     //三张底牌
}

type ResultData struct {
	AccountId  int    //用户Id
	Nickname   string // 昵称
	Desc       string
	Dealer     bool       //是否是庄家
	Hands      []string   //手牌
	Chips      int64      //金币
	Headimgurl string     //头像地址
}

func saveGameRecord(temp *GameRecord) {
	log.Release("保存战绩")
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		err := db.DB(DB).C("gamerecord").Insert(temp)
		if err != nil {
			log.Error("insert redpacketmatchresult data error: %v", err)
		}
	}, nil)
}

type RedPacketGrantRecord struct {
	CreatedAt	int64		//用户充值时间
	Nickname	string		//昵称
	AccountID	int			//用户ID
	GrantType	int			//发放类型，1 为红包任务    2 为分享赚钱
	Desc		string 		//原因
	Value		float64		//金额
}

func (ctx *RedPacketGrantRecord)Save() error {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	return db.DB(DB).C("red_packet_grant_record").Insert(ctx)
}

func WriteRedPacketGrantRecord(userData *UserData, grantType int, desc string, fee float64) {
	redPacketGrantRecord := &RedPacketGrantRecord{
		CreatedAt:time.Now().Unix(),
		Nickname:userData.Nickname,
		AccountID:userData.AccountID,
		GrantType:grantType,
		Desc:desc,
		Value:fee,
	}

	go func() {
		if err := redPacketGrantRecord.Save();err != nil {
			log.Error(err.Error())
		}
	}()
}

type RechargeRecord struct {
	CreatedAt	int64		//完成时间
	AccountID	int 		//用户id
	NickName	string 		//用户昵称
	Desc 		string 		//购买商品类型：金币、钻石
	Value 		float64		//金额
	Channel		int			//充值渠道	1. wxpay    2. alipay    3. applepay
}

func (ctx *RechargeRecord) Save () error {
	se := mongoDB.Ref()
	defer mongoDB.UnRef(se)
	return se.DB(DB).C("recharge_record").Insert(ctx)
}

func WriteRechageRecord (userData *UserData, createdAt int64, desc string, value float64, channel int) {
	rechargeRedcord := &RechargeRecord{
		CreatedAt: createdAt,
		AccountID:userData.AccountID,
		Desc:desc,
		Value:value,
		Channel:channel,
	}

	go func() {
		if err := rechargeRedcord.Save(); err != nil {
			log.Error(err.Error())
		}
	}()
}
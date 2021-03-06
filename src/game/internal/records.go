package internal

import (
	"common"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
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
	IsSpring	   bool         //这一局是否为春天
	LastThree	   []string     //三张底牌
	Channel 	   int          //渠道号
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
	Channel     int 		//渠道号
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
		Channel:userData.Channel,
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
	DownChannel int			//渠道号
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
		NickName:userData.Nickname,
		Desc:desc,
		Value:value,
		Channel:channel,
		DownChannel:userData.Channel,
	}

	go func() {
		if err := rechargeRedcord.Save(); err != nil {
			log.Error(err.Error())
		}
	}()
}

//搜狗活跃人数统计
type SougouActivityRecord struct {
	Num		  	int
	CreatedAt 	int64
}

func (ctx *SougouActivityRecord) Save () error {
	se := mongoDB.Ref()
	defer mongoDB.UnRef(se)
	_,err :=  se.DB(DB).C("sougou_activity_record").Upsert(
		bson.M{"createdat" : common.OneDay0ClockTimestamp(time.Now())},
		bson.M{"$inc" : bson.M{"num" : 1}})
	return err
}

func WriteSougouActivityRecord() {
	sougouActivityRecord := &SougouActivityRecord{}
	go func() {
		if err := sougouActivityRecord.Save(); err != nil {
			log.Error(err.Error())
		}
	}()
}

type ChipsRecord struct {
	AccountID	int
	CreatedAt   int64
	ActionType  int
	NickName    string
	AddChips 	int64
	Chips 		int64
	DownChannel int   //渠道号
}

func (ctx *ChipsRecord) Save() error {
	se := mongoDB.Ref()
	defer mongoDB.UnRef(se)
	return se.DB(DB).C("chip_record").Insert(ctx)
}

const (
	rechargeChip = 1
	subsidyChip = 2
	SignInChip = 3
)
func WriteChipsRecord(userData *UserData, addChip int64, actionType int) {
	chipsRecord := &ChipsRecord{
		AccountID	:userData.AccountID,
		CreatedAt   :time.Now().Unix(),
		ActionType  :actionType,
		NickName    :userData.Nickname,
		AddChips 	:addChip,
		Chips 		:userData.Chips,
		DownChannel :userData.Channel,
	}

	go func() {
		if err := chipsRecord.Save(); err != nil {
			log.Error(err.Error())
		}
	}()
}
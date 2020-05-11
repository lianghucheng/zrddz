package internal

import (
	"github.com/name5566/leaf/log"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type AgentProfit struct {
	Accountid       int
	Chips           int64
	FirstLevelChips int64
	OtherLevelChips int64
	CreateDat       int64
	Str             string
}

func (a *AgentProfit) insert() {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	selector := bson.M{"accountid": a.Accountid, "createdat": a.CreateDat}
	b := new(AgentProfit)
	err := db.DB(DB).C("agent_profit").Find(selector).One(b)
	if err != nil && err != mgo.ErrNotFound {
		log.Release("查找玩家当日收益失败")
		return
	}
	if err == mgo.ErrNotFound {
		db.DB(DB).C("agent_profit").Insert(a)
		return
	}
	update := bson.M{
		"$inc": bson.M{
			"firstlevelchips": a.FirstLevelChips,
			"otherlevelchips": a.OtherLevelChips,
			"chips":           a.Chips,
		},
	}
	db.DB(DB).C("agent_profit").Update(selector, update)
}

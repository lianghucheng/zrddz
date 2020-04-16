package internal

import (
	"time"

	"github.com/name5566/leaf/log"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Agent struct {
	Level int
	Datas []Data
}
type Data struct {
	Createdat int64   //创建时间
	AccountId int64   //用户Id
	Recharge  float64 //累计充值金额
	AllProfit float64 //贡献总收益
	Profit    float64 //可领取收益
	Updatedat int64   //更新时间
}

type UserAgent struct {
	Agents    []Agent
	AccountId int64
	ParentId  int64
}

func (a *Data) InsertAgent(level int, Accountid int64, ParentId int64) {
	selector := bson.M{"parentid": ParentId, "accountid": Accountid, "agents.level": level}
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	u := new(UserAgent)
	err := db.DB(DB).C("userAgents").Find(selector).One(u)
	if err != nil && err != mgo.ErrNotFound {
		log.Release("查找玩家是否有%v级下家失败:%v", level, err)
	}
	update := bson.M{"$push": bson.M{"agents.$.datas": a}}
	if err == mgo.ErrNotFound {
		a1 := new(Agent)
		a1.Level = level
		a1.Datas = append(a1.Datas, *a)
		selector = bson.M{"parentid": ParentId, "accountid": Accountid}
		update = bson.M{"$push": bson.M{"agents": a1}}
	}
	_, err = db.DB(DB).C("userAgents").Upsert(selector, update)
	if err != nil {
		log.Release("insert 账号:%v 的一级代理出现问题%v", a.AccountId, err)
	}
}

//更新某个玩家对他的上级产生的可领取盈利(返利获取的)
func (a *Data) updateAgent(level int, parentid int64, params map[string]float64) {
	selector := bson.M{"accountid": parentid, "agents.level": level,
		"agents": bson.M{"$elemMatch": bson.M{"level": level}}}
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	d := new(UserAgent)
	err := db.DB(DB).C("userAgents").Find(selector).One(d)
	if err != nil && err != mgo.ErrNotFound {
		log.Release("查看:%v第%v级代理%v的收益情况:%v", parentid, level, a.AccountId, err)
		return
	}
	label1 := 0
	label2 := 0
	for key, value := range d.Agents {
		if value.Level != level {
			continue
		}
		for key1, value1 := range value.Datas {
			label1 = key
			if a.AccountId == value1.AccountId {
				label2 = key1
			}
		}
	}

	for key, value := range params {
		switch key {

		case "profit":
			d.Agents[label1].Datas[label2].Profit = value
		case "allprofit":
			d.Agents[label1].Datas[label2].AllProfit = value
		case "recharge":
			d.Agents[label1].Datas[label2].Recharge = value
		}
	}
	err = db.DB(DB).C("userAgents").Update(selector, d)
	if err != nil {
		log.Release("更新账号:%v第%v级代理%v的收益情况:%v", parentid, level, a.AccountId, err)
	}
}

//更新某个玩家对他的上级产生的可领取盈利(返利获取的)这个是在原来的基础上添加
func (a *Data) updateIncAgent(level int, parentid int64, params map[string]float64) {
	selector := bson.M{"accountid": parentid, "agents.level": level,
		"agents": bson.M{"$elemMatch": bson.M{"level": level}}}
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	d := new(UserAgent)
	err := db.DB(DB).C("userAgents").Find(selector).One(d)
	if err != nil && err != mgo.ErrNotFound {
		log.Release("查看:%v第%v级代理%v的收益情况:%v", parentid, level, a.AccountId, err)
		return
	}
	label1 := 0
	label2 := 0
	for key, value := range d.Agents {
		if value.Level != level {
			continue
		}
		for key1, value1 := range value.Datas {
			label1 = key
			if a.AccountId == value1.AccountId {
				label2 = key1
			}
		}
	}

	for key, value := range params {
		switch key {

		case "profit":
			d.Agents[label1].Datas[label2].Profit += value
		case "allprofit":
			d.Agents[label1].Datas[label2].AllProfit += value
		case "recharge":
			d.Agents[label1].Datas[label2].Recharge += value
		}
	}
	d.Agents[label1].Datas[label2].Updatedat = time.Now().Unix()
	err = db.DB(DB).C("userAgents").Update(selector, d)
	if err != nil {
		log.Release("(在原有的基础上添加)更新账号:%v第%v级代理%v的收益情况:%v", parentid, level, a.AccountId, err)
	}
}

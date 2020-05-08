package internal

import (
	"time"

	"github.com/name5566/leaf/log"

	"msg"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*


	我的ID：	自己账户的ID号
	推荐人ID：	就是上级ID号
	直推总人数：	直属下级总用户人数
	昨日直推新增：	昨天新增的直属下级数量
	团队总人数：	团队所有人数
	昨日团队新增：	除直属下级外新增的人数
	昨日直推佣金奖励：	昨天直属下级所带来的收益
	昨天团队佣金奖励：	昨天自己团队所有人创造的佣金奖励（自己+直属+直属下级）



*/
func (user *User) ShareInfo() {

	data := new(msg.S2C_ShareInfo)
	data.AccountId = int64(user.baseData.userData.AccountID)
	data.ParentId = user.baseData.userData.ParentId
	now := time.Now()
	last := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	//获取直推总人数
	pipe := []bson.M{
		{

			"$project": bson.M{
				"_id":       1,
				"agents":    1,
				"accountid": 1,
			},
		},

		{

			"$unwind": "$agents",
		},
		{

			"$project": bson.M{
				"_id":       1,
				"agents":    1,
				"accountid": 1,
			},
		},

		{

			"$unwind": "$agents.datas",
		},
		{
			"$match": bson.M{
				"agents.level": 1,
				"accountid":    user.baseData.userData.AccountID,
			},
		},

		{

			"$group": bson.M{
				"_id": nil,
				"sum": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	data.FirstLevelNumber = getUserAgents(pipe)
	//获取直推昨日新增人数

	pipe = []bson.M{
		{

			"$project": bson.M{
				"_id":       1,
				"agents":    1,
				"accountid": 1,
			},
		},

		{

			"$unwind": "$agents",
		},
		{

			"$project": bson.M{
				"_id":       1,
				"agents":    1,
				"accountid": 1,
			},
		},

		{

			"$unwind": "$agents.datas",
		},
		{
			"$match": bson.M{
				"agents.level": 1,
				"accountid":    user.baseData.userData.AccountID,
				"agents.datas.createdat": bson.M{
					"$lt":  last.Unix(),
					"$gte": last.Unix() - 24*60*60,
				},
			},
		},

		{

			"$group": bson.M{
				"_id": nil,
				"sum": bson.M{
					"$sum": 1,
				},
			},
		},
	}

	data.FirstLevelAdd = getUserAgents(pipe)

	//获取团队总人数
	pipe = []bson.M{
		{

			"$project": bson.M{
				"_id":       1,
				"agents":    1,
				"accountid": 1,
			},
		},

		{

			"$unwind": "$agents",
		},
		{

			"$project": bson.M{
				"_id":       1,
				"agents":    1,
				"accountid": 1,
			},
		},

		{

			"$unwind": "$agents.datas",
		},

		{

			"$match": bson.M{
				"accountid": data.AccountId,
			},
		},
		{

			"$group": bson.M{
				"_id": nil,
				"sum": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	data.TeamNumber = getUserAgents(pipe)

	//获取昨日团队新增人数、、除直属下级外新增的人数
	pipe = []bson.M{
		{

			"$project": bson.M{
				"_id":       1,
				"agents":    1,
				"accountid": 1,
			},
		},

		{

			"$unwind": "$agents",
		},
		{

			"$project": bson.M{
				"_id":       1,
				"agents":    1,
				"accountid": 1,
			},
		},

		{

			"$unwind": "$agents.datas",
		},
		{
			"$match": bson.M{
				"agents.level": bson.M{
					"$gt": 1,
				},
				"accountid": user.baseData.userData.AccountID,
				"agents.datas.createdat": bson.M{
					"$lt":  last.Unix(),
					"$gte": last.Unix() - 24*60*60,
				},
			},
		},

		{

			"$group": bson.M{
				"_id": nil,
				"sum": bson.M{
					"$sum": 1,
				},
			},
		},
	}

	data.TeamAdd = getUserAgents(pipe)
	user.WriteMsg(data)
}

func getUserAgents(pipe interface{}) int {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	data := make(map[string]int)
	err := db.DB(DB).C("userAgents").Pipe(pipe).One(&data)
	if err != nil && err != mgo.ErrNotFound {
		log.Error("获取数据失败:%v", err)
		return 0
	}
	if err == mgo.ErrNotFound {
		return 0
	}
	return data["sum"]
}

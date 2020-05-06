package internal

/*
import (
	"time"

	"github.com/name5566/leaf/log"

	"msg"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

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
				"id":        1,
				"agents":    1,
				"accountid": 1,
			},
		},

		{

			"$unwind": "$agents",
		},

		{
			"$match": bson.M{
				"agents.level": 1,
				"accountid":    user.baseData.userData.AccountID,
			},
		},

		{

			"$group": bson.M{
				"id": nil,
				"sum": bson.M{
					"$sum": 1,
				},
			},
		},
	}

	//获取直推昨日新增人数

	pipe = []bson.M{
		{

			"$project": bson.M{
				"id":        1,
				"agents":    1,
				"accountid": 1,
			},
		},

		{

			"$unwind": "$agents",
		},

		{
			"$match": bson.M{
				"agents.level": 1,
				"accountid":    user.baseData.userData.AccountID,
				"agents.creatdat": bson.M{
					"$lt":  last.Unix(),
					"$gte": last.Unix() - 24*60*60,
				},
			},
		},

		{

			"$group": bson.M{
				"id": nil,
				"sum": bson.M{
					"$sum": 1,
				},
			},
		},
	}
}

func getUserAgents(pipe interface{}) int {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	data := make(map[string]int)
	err := db.DB(DB).C("userAgents").Pipe(pipe).All(&data)
	if err != nil && err != mgo.ErrNotFound {
		log.Error("获取数据失败")
		return 0
	}
	if err == mgo.ErrNotFound {
		return 0
	}
	return data["sum"]
}
*/

package internal

import (
	"common"
	"conf"
	"msg"
	"time"

	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
)

var (
	monthChipsRank []msg.MonthChipsRank // 金币排行榜
	monthWinsRank  []msg.MonthWinsRank  // 胜场排行榜
)

var (
	showRankLen = conf.GetCfgRank().ShowRankLen // 排行榜展示最大长度，带有头像和昵称的
)

func init() {
	common.HourClock(time.Duration(conf.GetCfgRank().UpdateRankTime)*time.Hour, generateMCR, generateMWR)
}

type TempUserInfo struct {
	Nickname   string
	Headimgurl string
}

// 生成MonthChipsRank
func generateMCR() {
	var result struct {
		UserID     int
		MonthChips int64
	}
	monthChipsRank = []msg.MonthChipsRank{}
	tempInfo := TempUserInfo{}
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	iter := db.DB(DB).C("monthrank").Find(bson.M{"monthchips": bson.M{"$gt": 0}}).Sort("-monthchips").Limit(showRankLen).Iter()
	for iter.Next(&result) {
		err := db.DB(DB).C("users").FindId(result.UserID).One(&tempInfo)
		if err != nil {
			log.Error("month rank monthchips find user info error : %v", err)
			continue
		}
		monthChipsRank = append(monthChipsRank, msg.MonthChipsRank{
			UserID:     result.UserID,
			Nickname:   tempInfo.Nickname,
			Headimgurl: tempInfo.Headimgurl,
			Chips:      result.MonthChips,
		})
	}
	if err := iter.Close(); err != nil {
		log.Error("generateMCR iter.Next error: %v", err)
	}
}

// 生成MonthWinsRank
func generateMWR() {
	var result struct {
		UserID    int
		MonthWins int
	}
	monthWinsRank = []msg.MonthWinsRank{}
	tempInfo := TempUserInfo{}
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	iter := db.DB(DB).C("monthrank").Find(bson.M{"monthwins": bson.M{"$gt": 0}}).Sort("-monthwins").Limit(showRankLen).Iter()
	for iter.Next(&result) {
		err := db.DB(DB).C("users").FindId(result.UserID).One(&tempInfo)
		if err != nil {
			log.Error("month rank monthwins find user info error : %v", err)
			continue
		}
		monthWinsRank = append(monthWinsRank, msg.MonthWinsRank{
			UserID:     result.UserID,
			Nickname:   tempInfo.Nickname,
			Headimgurl: tempInfo.Headimgurl,
			Wins:       result.MonthWins,
		})
	}
	if err := iter.Close(); err != nil {
		log.Error("generateMWR iter.Next error: %v", err)
	}
}

func (user *User) getMonthChipsRank(pageNum int) {
	if pageNum < 1 {
		return
	}
	onePageLen := 20 // 单页长度
	pageSum := 0     // 总页数
	if len(monthChipsRank) < showRankLen {
		pageSum = len(monthChipsRank) / onePageLen
		if len(monthChipsRank)%onePageLen > 0 {
			pageSum++
		}
	} else {
		pageSum = showRankLen / onePageLen
	}
	if pageNum > pageSum && pageSum != 0 {
		return
	}
	tempRanks := []msg.TempMCR{}
	if pageSum != 0 {
		start := onePageLen * (pageNum - 1)
		end := 0
		if start+onePageLen < len(monthChipsRank) {
			end = start + onePageLen
		} else {
			end = len(monthChipsRank)
		}
		for index := start; index < end; index++ {
			tempRanks = append(tempRanks, msg.TempMCR{
				Nickname:   monthChipsRank[index].Nickname,
				Headimgurl: monthChipsRank[index].Headimgurl,
				Chips:      monthChipsRank[index].Chips,
			})
		}
	}
	user.WriteMsg(&msg.S2C_UpdateMonthChipsRanks{
		PageNum:    pageNum,
		PageSum:    pageSum,
		ChipsRanks: tempRanks,
	})
}

func (user *User) getMonthChipsRankPos() {
	for i := 0; i < len(monthChipsRank); i++ {
		if user.baseData.userData.UserID == monthChipsRank[i].UserID {
			user.WriteMsg(&msg.S2C_UpdateMonthChipsRankPos{
				Pos:   i + 1,
				Chips: monthChipsRank[i].Chips,
			})
			return
		}
	}
	user.WriteMsg(&msg.S2C_UpdateMonthChipsRankPos{
		Pos:   0,
		Chips: 0,
	})
}

func (user *User) getMonthWinsRank(pageNum int) {
	if pageNum < 1 {
		return
	}
	onePageLen := 20 // 单页长度
	pageSum := 0     // 总页数
	if len(monthWinsRank) < showRankLen {
		pageSum = len(monthWinsRank) / onePageLen
		if len(monthWinsRank)%onePageLen > 0 {
			pageSum++
		}
	} else {
		pageSum = showRankLen / onePageLen
	}
	if pageNum > pageSum && pageSum != 0 {
		return
	}
	tempRanks := []msg.TempMWR{}
	if pageSum != 0 {
		start := onePageLen * (pageNum - 1)
		end := 0
		if start+onePageLen < len(monthWinsRank) {
			end = start + onePageLen
		} else {
			end = len(monthWinsRank)
		}
		for index := start; index < end; index++ {
			tempRanks = append(tempRanks, msg.TempMWR{
				Nickname:   monthWinsRank[index].Nickname,
				Headimgurl: monthWinsRank[index].Headimgurl,
				Wins:       monthWinsRank[index].Wins,
			})
		}
	}
	user.WriteMsg(&msg.S2C_UpdateMonthWinsRanks{
		PageNum:   pageNum,
		PageSum:   pageSum,
		WinsRanks: tempRanks,
	})
}

func (user *User) getMonthWinsRankPos() {
	for i := 0; i < len(monthWinsRank); i++ {
		if user.baseData.userData.UserID == monthWinsRank[i].UserID {
			user.WriteMsg(&msg.S2C_UpdateMonthWinsRankPos{
				Pos:  i + 1,
				Wins: monthWinsRank[i].Wins,
			})
			return
		}
	}
	user.WriteMsg(&msg.S2C_UpdateMonthWinsRankPos{
		Pos:  0,
		Wins: 0,
	})
}

func upsertMonthChipsRank(userID int, addChips int64) {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	_, err := db.DB(DB).C("monthrank").Upsert(bson.M{"userid": userID}, bson.M{"$inc": bson.M{"monthchips": addChips}})
	if err != nil {
		log.Error("monthrank chips add error: %v", err)
	}
}

func upsertMonthWinsRank(userID, addWins int) {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	_, err := db.DB(DB).C("monthrank").Upsert(bson.M{"userid": userID}, bson.M{"$inc": bson.M{"monthwins": addWins}})
	if err != nil {
		log.Error("monthrank wins add error: %v", err)
	}
}

func (user *User) dropMonthRank() {
	log.Debug("<-- DROP MONTH RANK START -->")
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		now := time.Now().Unix()
		// month chips rank
		mcr := struct {
			UserID     int
			MonthChips int64
			CleanTime  int64
		}{}
		if len(monthChipsRank) > 0 {
			max := 10
			if len(monthChipsRank) < max {
				max = len(monthChipsRank)
			}
			for i := 0; i < max; i++ {
				mcr.UserID = monthChipsRank[i].UserID
				mcr.MonthChips = monthChipsRank[i].Chips
				mcr.CleanTime = now
				err := db.DB(DB).C("monthrankchipsrecord").Insert(mcr)
				if err != nil {
					log.Error("insert month chips rank top 10 error: %v", err)
				}
			}
		}
		log.Debug("<-- STORE MONTH CHIPS RANK TOP 10 OVER -->")
		err := db.DB(DB).C("monthrank").DropCollection()
		if err != nil {
			log.Error("monthrank drop error: %v", err)
		}
		log.Debug("<-- DROP MONTH RANK OVER -->")
		user.WriteMsg(&msg.S2C_CleanMonthRanks{})
	}, nil)
}

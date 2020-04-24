package internal

import (
	"conf"
	"msg"

	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
)

var landlordConfigData *ConfigData

type ConfigData struct {
	ID   int "_id"
	Game string
	msg.C2S_SetLandlordConfig
}


func (data *ConfigData) initLandlord() {
	data.Game = conf.GetCfgDDZ().Gamename

	//db := mongoDB.Ref()
	//defer mongoDB.UnRef(db)
	//err := db.DB(DB).C("configs").Find(bson.M{"game": data.Game}).One(data)
	//if err == nil {
	//	return
	//}
	//if err != mgo.ErrNotFound {
	//	log.Error("init %v config data error: %v", data.Game, err)
	//	return
	//}
	//id, err := mongoDBNextSeq("configs")
	//if err != nil {
	//	log.Error("get next configs id error: %v", err)
	//	return
	//}
	//data.ID = id
	data.AndroidVersion = conf.GetCfgDDZ().AndroidVersion
	data.AndroidDownloadUrl = conf.GetCfgDDZ().DefaultAndroidDownloadUrl
	data.IOSVersion = conf.GetCfgDDZ().IOSVersion
	data.IOSDownloadUrl = conf.GetCfgDDZ().DefaultIOSDownloadUrl
	data.SougouVersion = conf.GetCfgDDZ().SougouVersion
	data.SougouDownloadUrl = conf.GetCfgDDZ().DefaultSougouDownloadUrl
	data.AndroidGuestLogin = conf.GetCfgDDZ().AndroidGuestLogin
	data.IOSGuestLogin = conf.GetCfgDDZ().IOSGuestLogin
	data.SougouGuestLogin = conf.GetCfgDDZ().SougouGuestLogin
	data.Notice = conf.GetCfgDDZ().Notice
	data.Radio = conf.GetCfgDDZ().Radio
	data.WeChatNumber = conf.GetCfgDDZ().WeChatNumber
	data.EnterAddress = conf.GetCfgDDZ().EnterAddress
	//saveConfigData(data)
}

func saveConfigData(configdata *ConfigData) {
	data := util.DeepClone(configdata)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		id := data.(*ConfigData).ID
		_, err := db.DB(DB).C("configs").UpsertId(id, data)
		if err != nil {
			log.Error("save %v config data error: %v", id, err)
		}
	}, func() {

	})
}

package internal

import (
	"conf"
	"fmt"
	"msg"
	"net/url"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/util"
)

type UserData struct {
	UserID          int `bson:"_id"`
	AccountID       int
	Nickname        string
	Headimgurl      string
	Sex             int // 1 男性，2 女性
	UnionID         string
	CircleID        int // 圈圈ID
	Serial          string
	Model           string
	LoginIP         string
	Token           string
	ExpireAt        int64 // token 过期时间
	Role            int   // 1 玩家、2 代理、3 管理员、4 超管
	Username        string
	Password        string
	Chips           int64 // 筹码
	Wins            int   // 胜场
	FreeChangedAt   int64 // 免费重置时间
	SubsidizedAt    int64 // 补助时间
	CreatedAt       int64
	UpdatedAt       int64
	ParentId        int64
	CardCode        string //取牌码
	Taken           bool   //是否领取
	CollectDeadLine int64  //取牌码的过期时间
	PlayTimes       int    //当天对局次数
	Level           int    //用户等级(初,中,高,完成10个任务自动升级）
	Online          bool   //玩家是否在线
	Channel         int    //渠道号。0：圈圈   1：搜狗   2:IOS
	SubsidyDeadLine      int64  //救济过期时间
	SubsidyTimes         int  //救济次数
}

const defaultAvatar = "https://www.shenzhouxing.com/czddz/dl/img/logo.jpg"

func (data *UserData) initValue(channel int) error {
	userID, err := mongoDBNextSeq("users")
	if err != nil {
		return fmt.Errorf("get next users id error: %v", err)
	}
	data.UserID = userID
	data.Role = rolePlayer
	// data.AccountID = common.GetID(4) + strconv.Itoa(data.UserID)
	data.AccountID = getAccountID()
	data.CreatedAt = time.Now().Unix()
	data.Channel = channel
	return nil
}

func saveUserData(userdata *UserData) {
	data := util.DeepClone(userdata)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		id := data.(*UserData).UserID
		_, err := db.DB(DB).C("users").UpsertId(id, data)
		if err != nil {
			log.Error("save user %v data error: %v", id, err)
		}
	}, nil)
}

func updateUserData(id int, update interface{}) {
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		_, err := db.DB(DB).C("users").UpsertId(id, update)
		if err != nil {
			log.Error("update user %v data error: %v", id, err)
		}
	}, nil)
}

func (data *UserData) updateWeChatInfo(info *msg.C2S_WeChatLogin) {
	if data.UnionID == "" {
		//表示新用户
		data.UnionID = info.UnionID
		switch data.UnionID {
		case "o8c-nt6tO8aIBNPoxvXOQTVJUxY0":
			data.Role = roleRoot
			// data.Chips = 99999999
		default:
			data.Role = rolePlayer
			data.Chips = int64(conf.Server.FirstLogin)
		}
	}
	if info.Nickname != "" {
		data.Nickname = info.Nickname
	}

	surl, err := url.Parse(info.Headimgurl)
	if err == nil {
		if surl.Scheme == "" {
			if data.Headimgurl == "" {
				data.Headimgurl = defaultAvatar
			}
		} else {
			if strings.HasSuffix(info.Headimgurl, "/0") {
				data.Headimgurl = info.Headimgurl[:len(info.Headimgurl)-1] + "132"
			} else {
				data.Headimgurl = info.Headimgurl
			}
		}
	} else {
		if data.Headimgurl == "" {
			data.Headimgurl = defaultAvatar
		}
	}
	if info.Sex == 1 {
		data.Sex = info.Sex
	} else {
		data.Sex = 2
	}
	data.Serial = info.Serial
	data.Model = info.Model

	data.UpdatedAt = time.Now().Unix()
}
func (data *UserData) readByAccountID(accountid int64) error {
	db := mongoDB.Ref()
	defer mongoDB.UnRef(db)
	return db.DB(DB).C("users").Find(bson.M{"accountid": accountid}).One(data)
}

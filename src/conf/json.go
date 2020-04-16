package conf

import (
	"common"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/name5566/leaf/log"
)

type Config struct {
	CfgLeafSvr        LeafSvr
	CfgMatchs         []CfgMatch
	CfgTimeout        CfgTimeout
	CfgDDZ            CfgDDZ
	CfgRank           CfgRank
	CfgActivityTimes  []CfgActivityTime
	CfgRedpacketCode  CfgRedpacketCode
	CfgCard           CfgCard
	CfgRedPacketItems map[string]RedPacketItem
	CfgLink           CfgLink
}
type LeafSvr struct {
	LogLevel       string
	LogPath        string
	WSAddr         string
	CertFile       string
	KeyFile        string
	TCPAddr        string
	MaxConnNum     int
	DBUrl          string
	DBMaxConnNum   int
	ConsolePort    int
	ProfilePath    string
	HTTPAddr       string
	DBName         string
	TaskFreeChange int64
	Level          int
	Chips          int
	RebateRate     float64
	FamilyActivity bool
	RoomCard       int
	FirstLogin     int
	PrimaryStart   int
	PrimaryEnd     int
	HighStart      int
	HighEnd        int
}
type CfgDDZ struct {
	DefaultAndroidDownloadUrl string
	DefaultIOSDownloadUrl     string
	Gamename                  string
	AndroidVersion            int
	IOSVersion                int
	AndroidGuestLogin         bool
	IOSGuestLogin             bool
	Notice                    string
	Radio                     string
	WeChatNumber              string
	EnterAddress              bool
	CardCodeDesc              string
}
type CfgRank struct {
	ShowRankLen    int
	UpdateRankTime int
}
type CfgMatch struct {
	BaseScore int
	MinScore  int
	MaxScore  int
	Tickets   int
}

type CfgTimeout struct {
	ConnectTimeout         int
	HeartTimeout           int
	LandlordBid            int
	LandlordGrab           int
	LandlordDouble         int
	LandlordShowCards      int
	LandlordDiscard        int
	LandlordDiscardNothing int
}

type CfgActivityTime struct {
	TaskID   int    // 任务id
	Start    string // 开始时间
	End      string // 结束时间
	Deadline string // 截至时间
}
type ActivityTime struct {
	TaskID   int   // 任务id
	Start    int64 // 开始时间
	End      int64 // 结束时间
	Deadline int64 // 截至时间
}
type CfgRedpacketCode struct {
	Url        string
	PartnerKey string
	SecretKey  string
	Method     string
}
type CfgCard struct {
	PlayTimes int
}
type RedPacketItem struct {
	Chips         int
	Desc          string
	RedPacketType int
	IsPrivate     bool
	Clock         []string
}
type CfgLink struct {
	CircleLink string
}

var Server LeafSvr
var ServerConfig Config

func init() {
	ReadConfigure()
	RedpacketTaskCfgInit()
	ShareCfgInit()
}
func ReadConfigure() {
	cfg := Config{}
	_, err := toml.DecodeFile("conf/ddz-server.toml", &cfg)
	if err != nil {
		log.Error("读取server.toml失败,error:%v", err)
	}
	ServerConfig = cfg
	Server = cfg.CfgLeafSvr
	log.Release("*****************:%v", ServerConfig)

}

func GetCfgTimeout() CfgTimeout {
	return ServerConfig.CfgTimeout
}

func GetCfgMatchs() []CfgMatch {
	return ServerConfig.CfgMatchs
}
func GetCfgDDZ() CfgDDZ {
	return ServerConfig.CfgDDZ
}
func GetCfgRank() CfgRank {
	return ServerConfig.CfgRank
}
func GetCfgActivityTimes() []*ActivityTime {
	activityTime := []*ActivityTime{}
	for _, v := range ServerConfig.CfgActivityTimes {
		start := strings.Split(v.Start, "-")
		end := strings.Split(v.End, "-")
		deadline := strings.Split(v.Deadline, "-")
		activityTime = append(activityTime, &ActivityTime{
			TaskID:   v.TaskID,
			Start:    time.Date(common.Atoi(start[0]), time.Month(common.Atoi(start[1])), common.Atoi(start[2]), common.Atoi(start[3]), common.Atoi(start[4]), common.Atoi(start[5]), 0, time.Local).Unix(),
			End:      time.Date(common.Atoi(end[0]), time.Month(common.Atoi(end[1])), common.Atoi(end[2]), common.Atoi(end[3]), common.Atoi(end[4]), common.Atoi(end[5]), 0, time.Local).Unix(),
			Deadline: time.Date(common.Atoi(deadline[0]), time.Month(common.Atoi(deadline[1])), common.Atoi(deadline[2]), common.Atoi(deadline[3]), common.Atoi(deadline[4]), common.Atoi(deadline[5]), 0, time.Local).Unix(),
		})
	}
	return activityTime
}
func GetCfgRedpacketCode() CfgRedpacketCode {
	return ServerConfig.CfgRedpacketCode
}
func GetCfgCard() CfgCard {
	return ServerConfig.CfgCard
}

func GetCfgRedPacketItems() map[string]RedPacketItem {
	return ServerConfig.CfgRedPacketItems
}

func GetCfgLink() CfgLink {
	return ServerConfig.CfgLink
}

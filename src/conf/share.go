package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/name5566/leaf/log"
)

type Share struct {
	CfgShareTasks 		[]CfgShareTask
	CfgShareCalcMethods	[]CfgShareCalcMethod
}
type CfgShareTask struct {
	ID    int
	Real  bool // true 表示获得 false 表示返利
	Total int
	Fee   float64
	Desc  string
	Info  string
}

type CfgShareCalcMethod struct {
	AchieveScope 		int 	//业绩区间范围（万）
	BackMoney			float64	//返佣金额比例（%）
	DiffMoney			float64	//代理差返佣百分比例
}

var ShareCfg Share

func ShareCfgInit() {
	_, err := toml.DecodeFile("conf/share.toml", &ShareCfg)
	if err != nil {
		log.Error("读取share.toml失败,error:%v", err)
	}
	log.Release("*****************:%v", ShareCfg.CfgShareTasks)
}
func GetCfgShareTask() []CfgShareTask {
	return ShareCfg.CfgShareTasks
}

func GetCfgShareCalcMethods() []CfgShareCalcMethod {
	return ShareCfg.CfgShareCalcMethods
}

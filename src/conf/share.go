package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/name5566/leaf/log"
)

type Share struct {
	CfgShareTasks []CfgShareTask
}
type CfgShareTask struct {
	ID    int
	Real  bool // true 表示获得 false 表示返利
	Total int
	Fee   float64
	Desc  string
	Info  string
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

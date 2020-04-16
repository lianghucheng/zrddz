package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/name5566/leaf/log"
)

type CfgPrimaryTask struct {
	ID    int
	Real  bool
	Total int
	Fee   float64
	Desc  string
	Jump  int
	Type  int //  红包类型 1:新人指定金额红包 2:新人幸运红包 3:指定金额红包 4:幸运红包
}

type CfgMiddleTask struct {
	ID    int
	Real  bool
	Total int
	Fee   float64
	Desc  string
	Jump  int
	Type  int //  红包类型 1:新人指定金额红包 2:新人幸运红包 3:指定金额红包 4:幸运红包

}
type CfgHighTask struct {
	ID    int
	Real  bool
	Total int
	Fee   float64
	Desc  string
	Jump  int
	Type  int //  红包类型 1:新人指定金额红包 2:新人幸运红包 3:指定金额红包 4:幸运红包

}

type RedpacketTaskList struct {
	CfgPrimaryTasks []CfgPrimaryTask
	CfgMiddleTasks  []CfgMiddleTask
	CfgHighTasks    []CfgHighTask
}

var RedpacketTaskCfg RedpacketTaskList

func RedpacketTaskCfgInit() {
	_, err := toml.DecodeFile("conf/redpacket-task.toml", &RedpacketTaskCfg)
	if err != nil {
		log.Error("读取redpacket-task.toml失败,error:%v", err)
	}
	log.Release("*****************:%v", RedpacketTaskCfg.CfgPrimaryTasks)
	log.Release("*****************:%v", RedpacketTaskCfg.CfgMiddleTasks)
	log.Release("*****************:%v", RedpacketTaskCfg.CfgHighTasks)
}
func GetCfgMiddleTask() []CfgMiddleTask {
	return RedpacketTaskCfg.CfgMiddleTasks
}

func GetCfgPrimaryTask() []CfgPrimaryTask {
	return RedpacketTaskCfg.CfgPrimaryTasks
}

func GetCfgHighTask() []CfgHighTask {
	return RedpacketTaskCfg.CfgHighTasks
}

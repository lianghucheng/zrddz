package msg

func init() {
	Processor.Register(&C2S_ShareInfo{})
	Processor.Register(&S2C_ShareInfo{})
}

type C2S_ShareInfo struct {
}

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
type S2C_ShareInfo struct {
	AccountId        int64   //用户Id
	ParentId         int64   //推荐人ID
	FirstLevelNumber int     //直推总人数
	FirstLevelAdd    int     //昨日直推新增人数
	TeamNumber       int     //团队总人数
	TeamAdd          int     //昨日团队新增人数
	FirstLevelProfit float64 //昨日直推佣金奖励
	TeamProfit       float64 //昨日团队佣金奖励
}

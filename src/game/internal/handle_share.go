package internal

import (
	"msg"

	"github.com/name5566/leaf/gate"
)

func init() {
	handler(&msg.C2S_BindSharer{}, handleBindSharer)
	handler(&msg.C2S_ShareRecord{}, handleShareRecord)
	handler(&msg.C2S_CopyExchangeCode{}, handleCopyExchangeCode)
	handler(&msg.C2S_Achievement{}, handleAchievement)
	handler(&msg.C2S_AbleProfit{}, handleAbleProfit)
	handler(&msg.C2S_AgentNumbersProfit{}, handleAgentNumbersProfit)
	handler(&msg.C2S_ReceiveShareProfit{}, handleReceiveShareProfit)
	handler(&msg.C2S_TakenProfit{}, handleTakenProfit)
}

func handleBindSharer(args []interface{}) {
	m := args[0].(*msg.C2S_BindSharer)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	user.BindSharer(m.AccountID)
}

func handleShareRecord(args []interface{}) {
	m := args[0].(*msg.C2S_ShareRecord)
	a := args[1].(gate.Agent)
	_ = m
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}
	user.shareAwardRecord(m.Page, m.Per)
}

func handleCopyExchangeCode(args []interface{}) {
	m := args[0].(*msg.C2S_CopyExchangeCode)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	user.copyExchangeCode(m.ShareRecordID)
}

func handleAchievement(args []interface{}) {
	m := args[0].(*msg.C2S_Achievement)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	user.achievement(m.Level, m.Page, m.Per)
}

func handleAbleProfit(args []interface{}) {
	m := args[0].(*msg.C2S_AbleProfit)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	user.ableProfit(m.Level)
}

func handleAgentNumbersProfit(args []interface{}) {
	m := args[0].(*msg.C2S_AgentNumbersProfit)
	a := args[1].(gate.Agent)
	_ = m
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	user.agentNumbersProfit()
}

func handleReceiveShareProfit(args []interface{}) {
	m := args[0].(*msg.C2S_ReceiveShareProfit)
	a := args[1].(gate.Agent)

	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	skeleton.Go(func() {
		user.receiveProfit(m.Level)
	}, nil)
}

func handleTakenProfit(args []interface{}) {
	m := args[0].(*msg.C2S_TakenProfit)
	a := args[1].(gate.Agent)
	_ = m
	if a.UserData() == nil {
		return
	}
	user := a.UserData().(*AgentInfo).user
	if user == nil {
		return
	}

	user.takenProfit()
}

package internal

import (
	"conf"
	"msg"
	"strings"
	"time"

	"github.com/name5566/leaf/gate"
)

type AgentInfo struct {
	user *User
}

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
	skeleton.RegisterChanRPC("WeChatLogin", rpcWeChatLogin)
	skeleton.RegisterChanRPC("TokenLogin", rpcTokenLogin)
	skeleton.RegisterChanRPC("UsernamePasswordLogin", rpcUsernamePasswordLogin)
}

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)

	a.SetUserData(new(AgentInfo))
	skeleton.AfterFunc(time.Duration(conf.GetCfgTimeout().ConnectTimeout)*time.Second, func() {
		if a.UserData() != nil {
			agentInfo := a.UserData().(*AgentInfo)
			if agentInfo != nil && agentInfo.user == nil {
				a.Close()
			}
		}
	})
}

func rpcWeChatLogin(args []interface{}) {
	a := args[0].(gate.Agent)
	m := args[1].(*msg.C2S_WeChatLogin)

	agentInfo := a.UserData().(*AgentInfo)
	// network closed
	if agentInfo == nil || agentInfo.user != nil {
		return
	}
	if strings.TrimSpace(m.UnionID) == "" {
		a.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_UnionIDInvalid})
		a.Close()
		return
	}
	if !systemOn && m.UnionID != "o8c-nt6tO8aIBNPoxvXOQTVJUxY0" {
		a.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_SystemOff})
		a.Close()
		return
	}
	newUser := newUser(a)
	a.UserData().(*AgentInfo).user = newUser
	newUser.wechatLogin(m)
}

func rpcTokenLogin(args []interface{}) {
	a := args[0].(gate.Agent)
	m := args[1].(*msg.C2S_TokenLogin)

	agentInfo := a.UserData().(*AgentInfo)
	// network closed
	if agentInfo == nil || agentInfo.user != nil {
		return
	}
	if strings.TrimSpace(m.Token) == "" {
		a.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_TokenInvalid})
		a.Close()
		return
	}
	if !systemOn {
		a.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_SystemOff})
		a.Close()
		return
	}
	newUser := newUser(a)
	a.UserData().(*AgentInfo).user = newUser
	newUser.tokenLogin(m.Token)
}

func rpcUsernamePasswordLogin(args []interface{}) {
	a := args[0].(gate.Agent)
	m := args[1].(*msg.C2S_UsernamePasswordLogin)

	agentInfo := a.UserData().(*AgentInfo)
	// network closed
	if agentInfo == nil || agentInfo.user != nil {
		return
	}
	if strings.TrimSpace(m.Username) == "" {
		a.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_UsernameInvalid})
		a.Close()
		return
	}
	if !systemOn {
		a.WriteMsg(&msg.S2C_Close{Error: msg.S2C_Close_SystemOff})
		a.Close()
		return
	}
	newUser := newUser(a)
	a.UserData().(*AgentInfo).user = newUser
	newUser.usernamePasswordLogin(m.Username, m.Password)
}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)

	user := a.UserData().(*AgentInfo).user
	a.SetUserData(nil)
	if user == nil {
		return
	}
	if user.state == userLogin {
		user.state = userLogout
		user.logout()
	}
}

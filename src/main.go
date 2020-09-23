package main

import (
	"conf"
	"game"
	"gate"
	"github.com/name5566/leaf/log"
	"login"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/name5566/leaf"
	lconf "github.com/name5566/leaf/conf"
)

func main() {
	log.Debug("dididi")
	lconf.LogLevel = conf.Server.LogLevel
	lconf.LogPath = conf.Server.LogPath
	lconf.LogFlag = conf.LogFlag
	lconf.ConsolePort = conf.Server.ConsolePort
	lconf.ProfilePath = conf.Server.ProfilePath
	handleSignal(syscall.SIGINT, handleINT)
	leaf.Run(
		game.Module,
		gate.Module,
		login.Module,
	)
}

//! 注册信号量
func handleSignal(signalType os.Signal, handleFun func(*chan os.Signal)) {
	ch := make(chan os.Signal)
	signal.Notify(ch, signalType)
	go handleFun(&ch)
}
//! ctrl+c
func handleINT(ch *chan os.Signal) {
	for {
		c := <-*ch
		//操作
		//game.BrocastServerDown(&msg.S2C_ServerDown{})
		time.AfterFunc(time.Second*1, func() {
			leaf.C <- c
		})
	}
}

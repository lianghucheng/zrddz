package gate

import (
	"game"
	"login"
	"msg"
)

func init() {
	// login
	msg.Processor.SetRouter(&msg.C2S_WeChatLogin{}, login.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_TokenLogin{}, login.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_UsernamePasswordLogin{}, login.ChanRPC)
	// game
	msg.Processor.SetRouter(&msg.C2S_SetSystemOn{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_SetLandlordConfig{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_SetUsernamePassword{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_TransferChips{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_SetUserRole{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.C2S_Heartbeat{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_CreateLandlordRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_EnterRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetUserChips{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_SetVIPRoomChips{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetCheckInDetail{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_CheckIn{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetAllPlayers{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_ExitRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_LandlordPrepare{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_LandlordMatching{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_LandlordBid{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_LandlordGrab{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_LandlordDouble{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_LandlordShowCards{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_LandlordDiscard{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetMonthChipsRank{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetMonthChipsRankPos{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetMonthWinsRank{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetMonthWinsRankPos{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_CleanMonthRanks{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_SystemHost{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_ChangeTable{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetRedPacketMatchRecord{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_TakeRedPacketMatchPrize{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_DoTask{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_ChangeTask{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_FreeChangeCountDown{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_TakeTaskPrize{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_FakeWXPay{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_FakeAliPay{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetCircleLoginCode{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.C2S_SetRobotData{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_BindSharer{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_ShareRecord{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_CopyExchangeCode{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_Achievement{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_AbleProfit{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_AgentNumbersProfit{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_ReceiveShareProfit{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetShareTaskInfo{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_TakenProfit{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetCardMa{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_TakeTaskState{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_GetRedpacketTaskCode{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_SubsidyChip{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_IsExistSubsidy{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_DailySign{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_ShareInfo{}, game.ChanRPC)
}

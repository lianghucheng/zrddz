package internal

func (user *User) doPrepare(r interface{}, showCards bool) {
	landlordRoom := r.(*LandlordRoom)
	if landlordRoom.state == roomGame {
		landlordRoom.reconnect(user)
	} else {
		landlordRoom.doPrepare(user.baseData.userData.UserID, showCards)
	}
}

func (user *User) doBid(r interface{}, bid bool) {
	landlordRoom := r.(*LandlordRoom)
	if landlordRoom.state == roomGame {
		landlordRoom.doBid(user.baseData.userData.UserID, bid)
	}
}

func (user *User) doGrab(r interface{}, grab bool) {
	landlordRoom := r.(*LandlordRoom)
	if landlordRoom.state == roomGame {
		landlordRoom.doGrab(user.baseData.userData.UserID, grab)
	}
}

func (user *User) doDouble(r interface{}, double bool) {
	landlordRoom := r.(*LandlordRoom)
	if landlordRoom.state == roomGame {
		landlordRoom.doDouble(user.baseData.userData.UserID, double)
	}
}

func (user *User) doShowCards(r interface{}, showCards bool) {
	landlordRoom := r.(*LandlordRoom)
	if landlordRoom.state == roomGame {
		landlordRoom.doShowCards(user.baseData.userData.UserID, showCards)
	}
}

func (user *User) doDiscard(r interface{}, cards []int) {
	landlordRoom := r.(*LandlordRoom)
	if landlordRoom.state == roomGame {
		landlordRoom.doDiscard(user.baseData.userData.UserID, cards)
	}
}

func (user *User) doSystemHost(r interface{}, host bool) {
	landlordRoom := r.(*LandlordRoom)
	if landlordRoom.state == roomGame {
		landlordRoom.doSystemHost(user.baseData.userData.UserID, host)
	}
}

func (user *User) doChangeTable(r interface{}) {
	landlordRoom := r.(*LandlordRoom)
	if landlordRoom.state == roomIdle {
		switch landlordRoom.rule.RoomType {
		case roomPractice, roomBaseScoreMatching, roomRedPacketMatching:
			landlordRoom.changeTable(user)
		}
	}
}

package temp_edy

import (
	"conf"
	"fmt"
	"github.com/name5566/leaf/log"
	"net/http"
	"strconv"
)

func RpcPayOK(accountid, amount int) {
	fmt.Println("远程调用支付成功", conf.Server.GameHttpAddress)
	//123.207.12.67
	resp, err := http.Get(conf.Server.GameHttpAddress+"/temppay?secret=123456&aid="+strconv.Itoa(accountid)+"&fee="+strconv.Itoa(amount))
	if err != nil {
		log.Error(err.Error())
	}

	_ = resp
}


package alipay

type AlipayTradeAppPayRequest struct {
	BizContent string
	Method     string // alipay.trade.app.pay
	NotifyUrl  string
	Version    string // 1.0
}

func NewAlipayTradeAppPayRequest(total_amount, outTradeNo string) *AlipayTradeAppPayRequest {
	bizContent := `{"product_code":"QUICK_MSECURITY_PAY","total_amount":"` + total_amount + `","subject":"车主斗地主游戏充值","out_trade_no":"` + outTradeNo + `"}`
	req := new(AlipayTradeAppPayRequest)
	req.BizContent = bizContent
	req.Method = "alipay.trade.app.pay"
	req.NotifyUrl = notifyUrl
	req.Version = "1.0"
	return req
}

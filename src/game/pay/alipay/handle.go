package alipay

import (
	"common"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/name5566/leaf/log"
)

var (
	rsaPrivateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAngF+Gi0z+WC61Ct1b8psllUSnoLs9iedJygZp2ATNVND7GizhkfWlfydadKf9tK7NUtkOjpYylxMGvwapnjR8rXNu+0ZIszfVHor7/Po9LEZRbIaO/++l8JI03OgBsN/pJdACvxpCngUEfwolcfUUR50YH93MpoRaBFvgNDPzW51W/NhNSYveUBYn55VK6ZFuDltliAvN40TDL4k8UA1q2p6j2h19XRV5R1hZP80b3hmRStRR4t7K8cRQ7awzaqYn/U+LEOb3DBOdMTI1wD7RRuEQmyzjC8WxgB+jVOZHgjOdK6TTx6iQtgSyCeh1IoGnOTAi689OsgJQz7gUzy+9wIDAQABAoIBAFjYeALaFhaMqKEzCqbgiOyDS6Pr9Lh5D+n7p2kxIbvjZRcizIeeD3BpCk59y8rrNa9DBEmlk1W+TmECDy46U7uJNPUN3gtubcm/pMMZQI2Oo6pH+m5wYMhOy8pygrIq7bQsBCvpQFtNp+NxCZUnNyCh4kh8hBblARKmcy9YuvBE4PR4yXrzDWtwkxyopdQPLI7Efeq572zfPftIILOdz7+lMraOXj6WMDpYIOf6e8FK2CIge8tbCwYPZHvl2BQIMvY6C4ocgX7rrAmhL+eEA+RmC0GOObiolTZ0VPSzDWZf2lESBmOrbAdrnkW9HgqDtiFb+hmg59sI9PXBO0VmSTECgYEAyYJWTwDLrxuEDKgjePf1F8IrzFfgy1IqkP0PZkba1TbJkQzgeSE1VyzSkaxenEDQSshrALBHs7oCPPRiwFsKV1i5FL7ueihbrEpG8o7eu40eZ+KANV6zAGGZHxnX1ytHKOJ07aOffxSIa9z0VCjfkVv1yExy54dJ9csVXhvVVK8CgYEAyLuZCtCZcA7xo1tmFeek+fwhpWORPpJbmpZMeyhP1PkAg0f2HHVncOxSbi+hwaZavKJ2LZfyWhXvWry5rTsAEtQP+Wi4MK1QMPNKO1ceaVBXbEC8YHPldXeWW9o2y4vAyZkrccrpa4itxUuHwFo6qDyFfehytjGkdvGnsR4SXDkCgYBqa3YPZRks0jhLwuRw92qt8HLXCTYDytIGHk9qsVLStYuAGi/WaM5VyqsuGb0hgi0+wVeZVn+XkE2sSVh5w9rTRF0Ccs9ZHkVD2Tpc0U0Z+a4sKPeSt/+K3QBT538Q+J8tHWOpOPd70qk1Zcx3QdrIVquX65/nXJCXyXfwanygqwKBgQDEXkEpQ0fXR8c7d342j5Xkt7JyiSTdgW/7mmzXTmhKgAzwYMVysaev4IADKrWjK4o4XvYdRDfhyPOOYHGD9ePsh2fZJYiKlgGM4XQM+PzXKbFcRTgDY11lvMdqs95G4UCH9z944nfWqq7UAz+Z/KrFSe+NbIhLk+TAN0dFDZYIgQKBgE95o0td5uSlGRVgVzP3bcqYgj0ytWyJKfWR6Y/vjcUqOUInjyhCDdSr6ufDIPI1QZHrdGaomVkz+Y3jUIMP+HGTM8KYCr5EGdwGXhHilRbuya6fqydDT9tgXEjelvzK2oJzJX8K7QiT4051+k1BLCqiubfpzBldyeOXS7OcPmoY
-----END RSA PRIVATE KEY-----`)

	alipayRSAPublicKey = []byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAhZ5tUwVnou3dfmginCDmRX8Lfu3HwOOitBEY0buaAb65C6dL9xXtnKJp5QzOuylfRNy4sdXXWXXd6SlkK0Xb8tuBxLcTW67TBz1SmzUstxvWbH18o8e6MED1VhvEt8puNdcBInfgDdeKx+x8X/u8zBldzL4K9Yieb67fMsfvDszWV2rI+BvxaKnNmYYBpC8oyJpq261d43WZosbuFghoT7hYks8NuLfPc+T6+xGRNfl0BrGMsVE4xAvE7E79AvXLCkZh+AV2FPGvy7TB0Dxbnn0mpNt2NrcwvMM7sbIlDL3hPdtCXl2/vY5KIA87qIyBQHpR+w9BTNIW5mkXm36ZQQIDAQAB
-----END PUBLIC KEY-----`)

	partnerID  = "2088811602965802"
	appID      = "2017110809804067"
	gatewayUrl = "https://openapi.alipay.com/gateway.do"
	notifyUrl  = "http://czddz.shenzhouxing.com:8084/alipay"
)

func DoRequest(req *AlipayTradeAppPayRequest) []byte {
	p := url.Values{}
	p.Add("app_id", appID)
	p.Add("biz_content", req.BizContent)
	p.Add("charset", "utf-8")
	p.Add("method", req.Method)
	p.Add("notify_url", req.NotifyUrl)
	p.Add("sign_type", "RSA2")
	p.Add("timestamp", time.Now().Format("2006-01-02 15:04:05"))
	p.Add("version", req.Version)
	p.Add("sign", generateSign(p))

	r, err := http.NewRequest("POST", gatewayUrl, strings.NewReader(p.Encode()))
	if err != nil {
		log.Debug("%v", err)
		return []byte{}
	}
	defer r.Body.Close()
	result, _ := ioutil.ReadAll(r.Body)
	return result
}

func rsaCheck(params url.Values) bool {
	sign := params.Get("sign")
	params.Del("sign")
	params.Del("sign_type")
	return verify([]byte(common.GetSignContent(params)), sign)
}

func Check(params url.Values) bool {
	tradeStatus := params.Get("trade_status")
	if appID == params.Get("app_id") && partnerID == params.Get("seller_id") && (tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED") {
		return rsaCheck(params)
	}
	return false
}

func generateSign(params url.Values) string {
	return sign([]byte(common.GetSignContent(params)))
}

func sign(data []byte) string {
	block, _ := pem.Decode(rsaPrivateKey)
	if block == nil {
		return ""
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return ""
	}
	h := sha256.New()
	h.Write(data)
	sign, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, h.Sum(nil))
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(sign)
}

func verify(data []byte, sign string) bool {
	sig, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}
	block, _ := pem.Decode(alipayRSAPublicKey)
	if block == nil {
		return false
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}
	h := sha256.New()
	h.Write(data)
	err = rsa.VerifyPKCS1v15(pub.(*rsa.PublicKey), crypto.SHA256, h.Sum(nil), sig)
	if err == nil {
		return true
	}
	return false
}

package internal

import (
	"common"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"game/pay/alipay"
	"game/pay/wxpay"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/name5566/leaf/log"
)

func handleAliPay(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		total_amount := r.URL.Query().Get("total_amount")
		account_id := r.URL.Query().Get("account_id")
		if total_amount == "" || account_id == "" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%v", "no total_amount or account_id")
			return
		}
		outTradeNo := common.GetOutTradeNo()
		request := alipay.NewAlipayTradeAppPayRequest(total_amount, outTradeNo)
		data := alipay.DoRequest(request)
		fmt.Fprintf(w, "%s", data)
		totalAmount, _ := strconv.ParseFloat(total_amount, 64)
		accountID, _ := strconv.Atoi(account_id)
		startAliPayOrder(outTradeNo, accountID, totalAmount, nil)
	case "POST":
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		log.Debug("result: %s", result)
		m, err := url.ParseQuery(string(result))
		if err == nil && alipay.Check(m) {
			// 需要验证 out_trade_no 和 total_amount
			fmt.Fprintf(w, "%v", "success")
			totalAmount, _ := strconv.ParseFloat(m.Get("total_amount"), 64)
			finishAliPayOrder(m.Get("out_trade_no"), totalAmount, true)
		} else {
			fmt.Fprintf(w, "%v", "failure")
		}
	}
}

func handleWXPay(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		total_fee := r.URL.Query().Get("total_fee")
		account_id := r.URL.Query().Get("account_id")
		if total_fee == "" || account_id == "" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%v", "no total_fee or account_id")
			return
		}
		totalFee, _ := strconv.Atoi(total_fee)
		accountID, _ := strconv.Atoi(account_id)
		if totalFee < 600 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%v", "充值金额不能低于6元")
			return
		}
		ip := strings.Split(r.RemoteAddr, ":")[0]
		p := wxpay.NewWXTradeAppPayParameter(total_fee, ip)
		data, err := json.Marshal(p)
		if err != nil {
			log.Error("marshal message %v error: %v", reflect.TypeOf(p), err)
			data = []byte{}
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, "%s", data)
		startWXPayOrder(p["out_trade_no"], accountID, totalFee, nil)
	case "POST":
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		log.Debug("result: %s", result)
		payResult := new(wxpay.WXPayResult)
		xml.Unmarshal(result, &payResult)
		if wxpay.VerifyPayResult(payResult) {
			// 需要验证 out_trade_no 和 total_fee
			fmt.Fprintf(w, "%v", wxpay.ReturnWXSuccess)
			finishWXPayOrder(payResult.OutTradeNo, payResult.TotalFee, true)
		} else {
			fmt.Fprintf(w, "%v", wxpay.ReturnWXFail)
		}
	}
}


func handleFakerAliPay(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		total_amount := r.URL.Query().Get("total_amount")
		account_id := r.URL.Query().Get("account_id")
		fmt.Fprintf(w, "%v", total_amount+"    "+account_id)
		if total_amount == "" || account_id == "" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%v", "no total_amount or account_id")
			return
		}
		outTradeNo := common.GetOutTradeNo()
		fmt.Fprintf(w, "faker alipay request    outTradeNo:%v", outTradeNo)
		totalAmount, _ := strconv.ParseFloat(total_amount, 64)
		accountID, _ := strconv.Atoi(account_id)
		startAliPayOrder(outTradeNo, accountID, totalAmount, nil)
	case "POST":
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		log.Debug("result: %s", result)
		// 需要验证 out_trade_no 和 total_amount
		fmt.Fprintf(w, "%v", "faker alipay response")
		totalAmount, _ := strconv.ParseFloat(r.FormValue("total_amount"), 64)
		finishAliPayOrder(r.FormValue("out_trade_no"), totalAmount, true)
	}
}

func handleFakerWXPay(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		total_fee := r.URL.Query().Get("total_fee")
		account_id := r.URL.Query().Get("account_id")
		if total_fee == "" || account_id == "" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%v", "no total_fee or account_id")
			return
		}
		totalFee, _ := strconv.Atoi(total_fee)
		accountID, _ := strconv.Atoi(account_id)
		if totalFee < 600 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%v", "充值金额不能低于6元")
			return
		}
		ip := strings.Split(r.RemoteAddr, ":")[0]
		p := wxpay.NewWXTradeAppPayParameter(total_fee, ip)
		data, err := json.Marshal(p)
		if err != nil {
			log.Error("marshal message %v error: %v", reflect.TypeOf(p), err)
			data = []byte{}
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, "faker wxpay    %s", data)
		startWXPayOrder(p["out_trade_no"], accountID, totalFee, nil)
	case "POST":
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		log.Debug("result: %s", result)
		// 需要验证 out_trade_no 和 total_fee
		fmt.Fprintf(w, "kafer wxpay response    %v    %v", r.FormValue("total_fee"),r.FormValue("out_trade_no"))
		total_fee,_ := strconv.Atoi(r.FormValue("total_fee"))
		finishWXPayOrder(r.FormValue("out_trade_no"), total_fee, true)
	}
}
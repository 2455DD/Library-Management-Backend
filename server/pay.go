package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
	"github.com/smartwalle/alipay/v3"
	. "lms/services"
	"log"
	"net/http"
)

var payAgent PayAgent

func initPayClient(cfg *ini.File) {
	ali, _ := cfg.GetSection("alipay")
	appId, _ := ali.GetKey("appId")
	private, _ := ali.GetKey("private")
	aliPublic, _ := ali.GetKey("aliPublic")

	client, err := alipay.New(appId.String(), private.String(), false)
	if err != nil {
		log.Fatal("alipay init err, ", err)
	}

	_ = client.LoadAppPublicCertFromFile("appCertPublicKey.crt")
	_ = client.LoadAliPayRootCertFromFile("alipayRootCert.crt")        // 加载支付宝根证书
	_ = client.LoadAliPayPublicCertFromFile("alipayCertPublicKey.crt") // 加载支付宝公钥证书
	_ = client.LoadAliPayPublicKey(aliPublic.String())

	payAgent.PayClient = client
}

func SignCheck(context *gin.Context) {
	req := context.Request
	_ = req.ParseForm()
	ok, err := payAgent.PayClient.VerifySign(req.Form)
	fmt.Println(ok, err)
	if !ok {
		context.JSON(http.StatusOK, gin.H{"SignCheck": false})
	} else {
		context.JSON(http.StatusOK, gin.H{"SignCheck": true})
	}
}

// 手机网页支付
// 传参示例http://127.0.0.1/pay/mobile?subject=fine&payId=12340&amount=10
func AliPayMobileHandler(context *gin.Context) {
	var p = alipay.TradeWapPay{}
	p.NotifyURL = "_"
	p.ReturnURL = "http://127.0.0.1/pay/signcheck"
	p.QuitURL = "http://127.0.0.1/pay/signcheck"

	p.Subject = context.Query("subject")
	p.OutTradeNo = context.Query("payId")
	p.TotalAmount = context.Query("amount")
	p.ProductCode = "QUICK_WAP_WAY"

	url, err := payAgent.PayClient.TradeWapPay(p)
	if err != nil {
		fmt.Println("pay client.TradeWapPay error:", err)
		return
	}

	binary, _ := url.MarshalBinary()
	fmt.Println(string(binary))
	data := make(map[string]interface{})
	data["url"] = string(binary)
	context.JSON(http.StatusOK, data)

}

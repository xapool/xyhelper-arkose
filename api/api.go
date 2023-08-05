package api

import (
	"log"
	"strings"
	"xyhelper-arkose/config"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

var (
	challengeUrl = "https://client-api.arkoselabs.com/fc/gt2/public_key/35536E1E-65B4-4D96-9D97-6ADB7EFF8147"
	headers      = map[string]string{
		"Origin":          "http://localhost:3000",
		"Referer":         "http://localhost:3000/v2/1.5.4/enforcement.cd12da708fe6cbe6e068918c38de2ad9.html",
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/116.0",
		"Content-Type":    "application/x-www-form-urlencoded; charset=UTF-8",
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2",
		"Sec-Fetch-Mode":  "cors",
		"Sec-Fetch-Site":  "same-origin",

		"Connection":     "keep-alive",
		"Sec-Fetch-Dest": "empty",
		"Cookie":         "gfsessionid=13yrbr4tjmbp40cujvuhyibfk8100im0; _dd_s=rum=0&expire=1691165643160; _account=1",
	}
)

func GetTokenByPayload(ctx g.Ctx, payload string) (string, error) {
	client := g.Client()
	client.SetHeaderMap(headers)
	client.SetProxy(config.Proxy)
	response, err := client.Post(ctx, challengeUrl, payload)
	if err != nil {
		log.Panic(err)
	}
	defer response.Close()
	if response.StatusCode != 200 {
		return "", gerror.New("获取token失败")
	}
	// response.RawDump()
	token := gjson.New(response.ReadAllString()).Get("token").String()
	if strings.Contains(token, "sup=1|rid=") {
		return token, nil
	}
	return "", gerror.New("获取token失败")

}

func GetPayloadFromCache(ctx g.Ctx) (payload config.Payload, err error) {
	cache := config.Cache.MustGet(ctx, "payload")
	if cache.IsEmpty() {
		return payload, gerror.New("payload is empty")
	}
	err = gconv.Struct(cache, &payload)
	if err != nil {
		return payload, gerror.New("payload format error")
	}
	return payload, nil

}

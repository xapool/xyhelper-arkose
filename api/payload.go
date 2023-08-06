package api

import (
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/encoding/gurl"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/text/gstr"
)

func Payload2BX(payload string) string {
	ctx := gctx.New()
	// form decode payload
	payload, _ = gurl.Decode(payload)
	// 以 \u0026 分割转换为数组
	payloadArray := gstr.Split(payload, `\u0026`)
	g.Dump(payloadArray)
	// 以 = 分割转换为map
	payloadMap := make(map[string]string)
	for _, v := range payloadArray {
		payloadMap[gstr.Split(v, "=")[0]] = gstr.Split(v, "=")[1]
	}
	g.Dump(payloadMap)
	bda := payloadMap["bda"]
	bda, err := gbase64.DecodeToString(bda)
	if err != nil {
		g.Log().Error(ctx, err)
	}
	g.Dump(bda)

	return bda
}

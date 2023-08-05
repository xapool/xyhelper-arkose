package handel

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
	"xyhelper-arkose/config"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

var (
	UpStream = "https://client-api.arkoselabs.com/"
)

func init() {

}

func Proxy(r *ghttp.Request) {
	ctx := r.Context()
	payload := &config.Payload{
		Payload: "",
		Created: time.Now().Unix(),
	}
	u, _ := url.Parse(UpStream)
	proxy := httputil.NewSingleHostReverseProxy(u)
	// g.Dump(config.PROXY(ctx))
	if config.PROXY(ctx).String() != "" {
		proxy.Transport = &http.Transport{
			Proxy: http.ProxyURL(config.PROXY(ctx)),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	proxy.Director = func(req *http.Request) {
		requrl := r.Request.URL.Path
		if requrl == "/fc/gt2/public_key/35536E1E-65B4-4D96-9D97-6ADB7EFF8147" {
			body := r.GetBodyString()
			bodyArray := gstr.Split(body, "&")
			// g.Dump(bodyArray)
			// 遍历数组 当数组元素以 "site=http" 开头时，将其替换为 "site=http%3A%2F%2Flocalhost%3A3000"
			for i, v := range bodyArray {
				if gstr.HasPrefix(v, "site=http") {
					bodyArray[i] = "site=http%3A%2F%2Flocalhost%3A3000"
				}
			}
			body = gstr.Join(bodyArray, "&")

			payload.Payload = body
			req.Body = io.NopCloser(bytes.NewReader(gconv.Bytes(body)))
			req.ContentLength = int64(len(body))
		}

		req.Header = r.Header
		req.Host = u.Host
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		req.URL.Path = requrl
		// req.URL.Path = r.Request.URL.Path
		// req.Header.Set("X-Forwarded-For", "23.254.56.178")
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Referer", "http://localhost:3000/v2/1.5.4/enforcement.cd12da708fe6cbe6e068918c38de2ad9.html")
		req.Header.Del("Cf-Connecting-Ip")
		req.Header.Del("Cf-Ipcountry")
		req.Header.Del("Cf-Ray")
		req.Header.Del("Cf-Request-Id")
		req.Header.Del("Cf-Visitor")
		req.Header.Del("Cf-Warp-Tag-Id")
		req.Header.Del("Cf-Worker")
		req.Header.Del("Cf-Device-Type")
		req.Header.Del("Cf-Request-Id")
		req.Header.Del("X-Forwarded-Host")
		req.Header.Del("X-Forwarded-Proto")
		req.Header.Del("X-Forwarded-For")
		req.Header.Del("X-Forwarded-Port")
		req.Header.Del("X-Forwarded-Server")
		req.Header.Del("X-Real-Ip")
		req.Header.Del("Accept-Encoding")
		g.Dump(req.Header)

	}
	proxy.ModifyResponse = func(resp *http.Response) error {

		// 解码 url
		if resp.StatusCode <= 400 {
			g.Log().Info(r.Context(), resp.StatusCode, resp.Request.URL.Path)
			// 获取返回的body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			g.Log().Info(r.Context(), "body", string(body))
			// 解压缩body

			// g.Dump(string(unzipbody))
			token := gjson.New(body).Get("token").String()
			g.Log().Info(r.Context(), "token", token)
			if strings.Contains(token, "sup=1|rid=") {
				// 获取请求的body
				err := config.Cache.Set(r.Context(), "payload", payload, 0)
				if err != nil {
					return err
				}
				g.Log().Info(r.Context(), "refresh payload cache", payload)

			}
			// 将原始body 返回
			resp.Body = io.NopCloser(bytes.NewReader(body))
		} else {
			g.Log().Warning(r.Context(), resp.StatusCode, resp.Request.URL.Path)

		}
		return nil
	}

	proxy.ServeHTTP(r.Response.RawWriter(), r.Request)

}

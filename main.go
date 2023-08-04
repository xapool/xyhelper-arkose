package main

import (
	// _ "xyhelper-arkose/ja3proxy"

	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"xyhelper-arkose/api"
	"xyhelper-arkose/config"
	"xyhelper-arkose/handel"

	api2captcha "github.com/2captcha/2captcha-go"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/median/captchago"
)

func main() {
	ctx := gctx.New()
	s := g.Server()
	// 每小时清理一次
	if g.Cfg().MustGetWithEnv(ctx, "FORWORD_URL").String() == "" {
		_, err := gcron.AddSingleton(ctx, "0 0 * * * *", func(ctx context.Context) {
			tokenURI := "http://127.0.0.1:" + gconv.String(config.Port) + "/cleantoken"
			g.Log().Print(ctx, "Every hour", tokenURI)
			g.Client().Get(ctx, tokenURI)
			api.RefreshPayloadFromMaster(ctx)

		}, "clean")
		if err != nil {
			panic(err)
		}
	}
	// s.EnableHTTPS("./resource/certs/server.crt", "./resource/certs/server.key")
	// s.SetHTTPSPort(443)
	api.RefreshPayloadFromMaster(ctx)

	s.SetPort(config.Port)
	s.SetServerRoot("resource/public")
	s.BindHandler("/ping", func(r *ghttp.Request) {
		total := config.TokenQueue.Size()
		r.Response.WriteJson(g.Map{
			"code":  1,
			"msg":   "pong",
			"total": total,
		})

	})
	// s.BindHandler("/", api.Index)
	s.BindHandler("/*", handel.Proxy)

	s.BindHandler("/token", func(r *ghttp.Request) {
		ctx := r.Context()

		// get token from local tokenqueue
		result, err := GetTokenFromQueue(ctx)
		if err != nil {
			// get from public arkose url when get from local failed
			fallbackURL := g.Cfg().MustGetWithEnv(ctx, "ARKOSE_TOKEN_FALLBACK_URL").String()
			if fallbackURL != "" {
				resp, err := g.Client().Get(ctx, fallbackURL)
				if err == nil && resp.StatusCode == 200 {
					defer resp.Body.Close()
					responseMap := make(map[string]interface{})
					if err := json.NewDecoder(resp.Body).Decode(&responseMap); err == nil {
						arkoseToken, ok := responseMap["token"]
						if ok && arkoseToken != "" {
							r.Response.WriteJson(responseMap)
							g.Log().Info(ctx, "get token from fallback url", arkoseToken)
							return
						}
					}
				}
			}
			// second get from solver
			arkoseToken, err := GetTokenFromSolver(ctx)
			if err != nil || arkoseToken == "" {
				r.Response.WriteJson(g.Map{
					"code": 0,
					"msg":  err.Error(),
				})
				return
			}
			r.Response.WriteJson(g.Map{
				"token": arkoseToken,
				"created": time.Now().Unix(),
			})
			g.Log().Info(ctx, "get token from solver", arkoseToken)
			return
		}
		r.Response.WriteJson(result)
	})
	s.BindHandler("/payload", func(r *ghttp.Request) {
		ctx := r.Context()
		r.Cookie.Set("uid", gtime.Now().String())

		payload, err := api.GetPayloadFromCache(ctx)
		if err != nil {
			g.Log().Error(ctx, err)
			r.Response.WriteJson(g.Map{
				"code": 0,
				"msg":  err.Error(),
			})
			return
		}
		r.Response.WriteJson(gjson.New(payload))

	})
	s.BindHandler("/pushtoken", func(r *ghttp.Request) {
		// g.Dump(r.Header)
		token := r.Get("token").String()
		if token == "" {
			r.Response.WriteJson(g.Map{
				"code": 0,
				"msg":  "token is empty",
			})
			return
		}
		// if !strings.Contains(token, "sup=1|rid=") {
		// 	g.Log().Error(ctx, "token error", token)
		// 	r.Response.WriteJson(g.Map{
		// 		"code": 0,
		// 		"msg":  "token error",
		// 	})
		// 	return
		// }
		forwordURL := g.Cfg().MustGetWithEnv(ctx, "FORWORD_URL").String()
		g.Log().Info(ctx, "forwordURL", forwordURL)

		if forwordURL != "" {
			result := g.Client().Proxy(config.Proxy).PostVar(ctx, forwordURL, g.Map{
				"token": token,
			})
			g.Log().Info(ctx, getRealIP(r), "forwordURL", forwordURL, result)
			r.Response.WriteJson(g.Map{
				"code":       1,
				"msg":        "success",
				"forwordURL": forwordURL,
			})
			return
		}
		Token := config.Token{
			Token:   token,
			Created: time.Now().Unix(),
		}
		config.TokenQueue.Push(Token)
		g.Log().Info(r.Context(), getRealIP(r), "pushtoken", token)
		r.Response.WriteJson(g.Map{
			"code": 1,
			"msg":  "success",
		})
	})
	s.BindHandler("/cleantoken", func(r *ghttp.Request) {
		ctx := r.Context()

		result, err := GetTokenFromQueue(ctx)
		g.Log().Info(ctx, "clean done，now pool size is", config.TokenQueue.Size())
		if err != nil {
			r.Response.WriteJson(g.Map{
				"code": 0,
				"msg":  err.Error(),
			})
			return
		}
		r.Response.WriteJson(result)
	})
	s.Run()
}

func getRealIP(req *ghttp.Request) string {
	// 优先获取Cf-Connecting-Ip
	if ip := req.Header.Get("Cf-Connecting-Ip"); ip != "" {
		return ip
	}

	// 优先获取X-Real-IP
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// 其次获取X-Forwarded-For
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	// 最后获取RemoteAddr
	ip := req.RemoteAddr
	// 处理端口
	if index := strings.Index(ip, ":"); index != -1 {
		ip = ip[0:index]
	}
	if ip == "[" {
		ip = req.GetClientIp()
	}
	return ip
}

func GetTokenFromQueue(ctx g.Ctx) (map[string]interface{}, error) {
	var token interface{}

	for config.TokenQueue.Size() > 0 {
		token = config.TokenQueue.Pop()
		var tokenStuct config.Token
		gconv.Struct(token, &tokenStuct)
		if time.Now().Unix()-tokenStuct.Created < int64(config.TokenExpire) {
			break
		} else {
			g.Log().Info(ctx, "token is expired,will pop one ", config.TokenQueue.Size(), tokenStuct.Created, config.TokenExpire)
			token = nil
		}
	}

	if token == nil {
		g.Log().Info(ctx, "token is empty, will get one")
		payload, err := api.GetPayloadFromCache(ctx)
		if err != nil {
			g.Log().Error(ctx, err)
			return "", err
		}
		newtoken, err := api.GetTokenByPayload(ctx, payload.Payload, payload.UserAgent)
		if err != nil {
			g.Log().Error(ctx, err)
			return "", err
		}
		token := g.Map{
			"code":    1,
			"token":   newtoken,
			"created": time.Now().Unix(),
		}
	}

	var result map[string]interface{}
	gconv.Struct(token, &result)

	return result, nil
}

func GetTokenFromSolver(ctx g.Ctx) (string, error) {
	captchaSolver := g.Cfg().MustGetWithEnv(ctx, "CAPTCHA_SOLVER").String()
	captchaSolverKey := g.Cfg().MustGetWithEnv(ctx, "CAPTCHA_SOLVER_KEY").String()

	var arkoseToken string
	var err error

	switch captchaSolver {
	case "CapSolver":
		arkoseToken, err = GetTokenFromCapSolver(ctx, captchaSolverKey)
	case "2Captcha":
		arkoseToken, err = GetTokenFrom2Captcha(ctx, captchaSolverKey)
	}

	if err != nil {
		return "", err
	}
	return arkoseToken, nil
}

func GetTokenFrom2Captcha(ctx g.Ctx, api_key string) (string, error) {
	client := api2captcha.NewClient(api_key)
	cap := api2captcha.FunCaptcha {
		SiteKey: "35536E1E-65B4-4D96-9D97-6ADB7EFF8147",
		Url: "https://chat.openai.com",
		Surl: "https://client-api.arkoselabs.com",
		UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
		// Data: map[string]string{"anyKey":"anyValue"},
	}
	req := cap.ToRequest()
	// req.SetProxy("HTTPS", "login:password@IP_address:PORT")
	arkoseToken, err := client.Solve(req)
	if err != nil {
		g.Log().Error(ctx, "get token from 2Capatch failed", err)
		return "", err
	}

	g.Log().Info(ctx, "get token from 2Captcha")
	return arkoseToken, nil
}

func GetTokenFromCapSolver(ctx g.Ctx, api_key string) (string, error) {
	solver, err := captchago.New(captchago.CapSolver, api_key)
	if err != nil {
		return "", err
	}

	sol, err := solver.FunCaptcha(captchago.FunCaptchaOptions{
		PageURL: "https://client-api.arkoselabs.com",
		PublicKey: "35536E1E-65B4-4D96-9D97-6ADB7EFF8147",
		Subdomain: "https://chat.openai.com",
		UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	})

	if err != nil {
		return "", err
	}

	g.Log().Info(ctx, fmt.Sprintf("get token from CapSolver solved in %v ms", sol.Speed))
	return sol.Text, nil
}

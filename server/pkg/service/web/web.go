package web

import (
	"encoding/json"
	"honeypot/server/pkg/logger"
	"honeypot/server/pkg/util"
	"io"
	"net"
	"time"

	"github.com/gin-gonic/gin"
)

func StartWeb(addr string, flag bool) error {
    logger.Log.Warningf("start web service on %v", addr)
    
    // 设置为发布模式
    gin.SetMode(gin.ReleaseMode)

    r := gin.Default()
    r.Use(Flagger(flag))

    // 注册路由
    r.Any("/", IndexHandle)
    
    return r.Run(addr)
}

// Flagger 中间件，用于传递 flag 参数
func Flagger(flag bool) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        ctx.Set("flag", flag)
        ctx.Next()
    }
}

// IndexHandle 处理所有请求
func IndexHandle(ctx *gin.Context) {
    flag, _ := ctx.Get("flag")
    isProxy := flag.(bool)

    // 获取请求信息
    _ = ctx.Request.ParseForm()
    params := ctx.Request.Form
    remoteAddr := ctx.Request.RemoteAddr
    host := ctx.Request.Host
    method := ctx.Request.Method
    uri := ctx.Request.RequestURI
    headers := ctx.Request.Header

    // 读取请求体
    body, _ := io.ReadAll(ctx.Request.Body)

    if isProxy {
        // 获取真实IP
        rawIp := util.GetRawIpByConn(ctx.Request.Context().Value("conn").(net.Conn))

        // 记录访问日志
        accessLog := struct {
            Time       time.Time              `json:"time"`
            IP         string                 `json:"ip"`
            RemoteAddr string                 `json:"remote_addr"`
            Host       string                 `json:"host"`
            Service    string                 `json:"service"`
            Method     string                 `json:"method"`
            URI        string                 `json:"uri"`
            Headers    map[string][]string    `json:"headers"`
            Params     map[string][]string    `json:"params"`
            Body       string                 `json:"body"`
        }{
            Time:       time.Now(),
            IP:         rawIp.String(),
            RemoteAddr: remoteAddr,
            Host:      host,
            Service:   "web",
            Method:    method,
            URI:       uri,
            Headers:   headers,
            Params:    params,
            Body:      string(body),
        }

        // 发送到日志服务器
        if jsonData, err := json.Marshal(accessLog); err == nil {
            logger.LogReport.WithField("api", "/api/web/").Info(string(jsonData))
        }
    }

    // 返回一个简单的响应
    ctx.String(200, "Welcome to Honeypot Web Server\n")
}
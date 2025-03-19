package redisServer

import (
	"encoding/json"
	"github.com/redis-go/redcon"
	"honeypot/server/pkg/logger"
	"honeypot/server/pkg/util"
	"strings"
	"time"
)

func StartRedis(addr string, flag bool) error {
	return redcon.ListenAndServe(addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			// 记录攻击者信息
			if flag {
				rawIp := util.GetRawIpByConn(conn.NetConn())

				// 获取命令内容
				tmpCmd := make([]string, 0)
				for _, c := range cmd.Args {
					tmpCmd = append(tmpCmd, string(c))
				}

				// 记录访问日志
				accessLog := struct {
					Time    time.Time `json:"time"`
					IP      string    `json:"ip"`
					Service string    `json:"service"`
					Command string    `json:"command"`
				}{
					Time:    time.Now(),
					IP:      rawIp.String(),
					Service: "redis",
					Command: strings.Join(tmpCmd, " "),
				}

				// 发送到日志服务器
				if jsonData, err := json.Marshal(accessLog); err == nil {
					logger.LogReport.WithField("api", "/api/redis/").Info(string(jsonData))
				}
			}

			// 处理Redis命令
			switch strings.ToLower(string(cmd.Args[0])) {
			case "ping":
				conn.WriteString("PONG")
			case "quit":
				conn.WriteString("OK")
				err := conn.Close()
				if err != nil {
					return
				}
			case "set":
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for 'set' command")
					return
				}
				conn.WriteString("OK")
			case "get":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for 'get' command")
					return
				}
				conn.WriteNull()
			case "info":
				conn.WriteString("redis_version:6.0.0\r\nredis_mode:standalone\r\nos:Linux")
			default:
				conn.WriteString("OK")
			}
		},
		func(conn redcon.Conn) bool {
			return true
		},
		func(conn redcon.Conn, err error) {
			if err != nil {
				logger.Log.Errorf("Redis connection closed with error: %v", err)
			}
		},
	)
}

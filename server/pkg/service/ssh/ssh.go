package ssh

import (
    "encoding/json"  
    "github.com/gliderlabs/ssh"
    "honeypot/server/pkg/logger"
    "honeypot/server/pkg/util"
    "io"
    "net"
    "time"
)

type SSHService struct {
	server  *ssh.Server
	isProxy bool
}

// StartSSH 启动SSH服务
func StartSSH(addr string, isProxy bool) error {
    sshService := &SSHService{
        isProxy: isProxy,
    }

    server := &ssh.Server{
        Addr:            addr,
        PasswordHandler: sshService.handlePassword,
        Handler:         sshService.handleSession,
    }

    sshService.server = server
    return server.ListenAndServe()
}

// handlePassword 处理密码验证
func (s *SSHService) handlePassword(ctx ssh.Context, password string) bool {
    var attackerIP net.IP

    // 获取攻击者IP
    if s.isProxy {
        // 如果是通过代理来的，从RawIps中获取真实IP
        if conn, ok := ctx.Value("conn").(net.Conn); ok && conn != nil {
            attackerIP = util.GetRawIpByConn(conn)
        }
    } else {
        // 直接连接的情况
        remoteAddr := ctx.RemoteAddr()
        if addr, ok := remoteAddr.(*net.TCPAddr); ok {
            attackerIP = addr.IP
        }
    }

	// 记录认证尝试
	logAuthAttempt(attackerIP, ctx.User(), password)

	// 永远返回认证成功，引导到假的堡垒机
	return true
}

// handleSession 处理SSH会话
func (s *SSHService) handleSession(session ssh.Session) {
	// 模拟一个假的堡垒机界面
	banner := `
Welcome to Bastion Host
This connection is monitored and recorded
Disconnect IMMEDIATELY if you are not an authorized user!
`
	io.WriteString(session, banner)

	// 等待一段时间后断开连接
	time.Sleep(5 * time.Second)
	session.Close()
}

// logAuthAttempt 记录认证尝试
func logAuthAttempt(ip net.IP, username, password string) {
	authLog := struct {
		Time     time.Time `json:"time"`
		IP       string    `json:"ip"`
		Username string    `json:"username"`
		Password string    `json:"password"`
		Service  string    `json:"service"`
	}{
		Time:     time.Now(),
		IP:       ip.String(),
		Username: username,
		Password: password,
		Service:  "ssh",
	}

	// 发送到日志服务器
	if jsonData, err := json.Marshal(authLog); err == nil {
		logger.LogReport.WithField("api", "/api/auth/").Info(string(jsonData))
	}
}

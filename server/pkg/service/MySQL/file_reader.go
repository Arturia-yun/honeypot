package mysql

import (
	"fmt"
	"honeypot/server/pkg/logger"
	"honeypot/server/pkg/util"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// FileReadHandler 处理MySQL文件读取请求
func FileReadHandler(conn net.Conn, isProxy bool) {
	defer conn.Close()
	
	// 获取攻击者IP
	var attackerIP string
	if isProxy {
		ip := util.GetRawIpByConn(conn)
		attackerIP = ip.String()
	} else {
		if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
			attackerIP = addr.IP.String()
		}
	}
	
	// 发送MySQL握手包
	handshakePacket := []byte{
		0x4a, 0x00, 0x00, 0x00, 0x0a, 0x35, 0x2e, 0x35, 0x2e, 0x35, 
		0x33, 0x00, 0x17, 0x00, 0x00, 0x00, 0x6e, 0x7a, 0x3b, 0x54, 
		0x76, 0x73, 0x61, 0x6a, 0x00, 0xff, 0xf7, 0x21, 0x02, 0x00, 
		0x0f, 0x80, 0x15, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 
		0x00, 0x00, 0x00, 0x70, 0x76, 0x21, 0x3d, 0x50, 0x5c, 0x5a, 
		0x32, 0x2a, 0x7a, 0x49, 0x3f, 0x00, 0x6d, 0x79, 0x73, 0x71, 
		0x6c, 0x5f, 0x6e, 0x61, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x70, 
		0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x00,
	}
	conn.Write(handshakePacket)
	
	// 接收客户端认证包
	authBuf := make([]byte, 1024)
	_, err := conn.Read(authBuf)
	if err != nil {
		logger.Log.Errorf("读取认证包失败: %v", err)
		return
	}
	
	// 发送认证成功包
	authOkPacket := []byte{0x07, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00}
	conn.Write(authOkPacket)
	
	// 接收文件读取请求
	fileBuf := make([]byte, 1024)
	n, err := conn.Read(fileBuf)
	if err != nil {
		logger.Log.Errorf("读取文件请求失败: %v", err)
		return
	}
	
	// 解析文件路径 (从第5个字节开始)
	if n < 5 {
		return
	}
	
	filePath := string(fileBuf[5:n])
	
	// 准备响应内容 (可以是真实文件内容或伪造内容)
	var fileContent []byte
	
	// 这里可以根据请求的文件路径返回伪造的内容
	switch {
	case strings.Contains(filePath, "passwd"):
		fileContent = []byte("root:x:0:0:root:/root:/bin/bash\ndaemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin\n")
	case strings.Contains(filePath, "shadow"):
		fileContent = []byte("root:$6$xyz$abcdefghijklmnopqrstuvwxyz:18395:0:99999:7:::\n")
	case strings.Contains(filePath, "hosts"):
		fileContent = []byte("127.0.0.1 localhost\n127.0.1.1 honeypot\n")
	case strings.Contains(filePath, "win.ini"):
		fileContent = []byte("[Mail]\nMAPI=1\n[MCI Extensions.BAK]\n")
	case strings.Contains(filePath, "my.cnf") || strings.Contains(filePath, "my.ini"):
		fileContent = []byte("[mysqld]\nport=3306\nuser=mysql\ndatadir=/var/lib/mysql\n")
	case strings.Contains(filePath, "wp-config.php"):
		fileContent = []byte("<?php\ndefine('DB_NAME', 'wordpress');\ndefine('DB_USER', 'admin');\ndefine('DB_PASSWORD', 'password123');\n?>")
	default:
		fileContent = []byte(fmt.Sprintf("Content of %s\nThis is a honeypot system.\n", filePath))
	}
	
	// 记录攻击日志并发送到日志服务器 - 使用更结构化的格式
	fileReadLog := struct {
		Time      time.Time `json:"time"`
		IP        string    `json:"ip"`
		Service   string    `json:"service"`
		Event     string    `json:"event"`
		FilePath  string    `json:"file_path"`
		Response  string    `json:"response"`
	}{
		Time:      time.Now(),
		IP:        attackerIP,
		Service:   "mysql",
		Event:     "file_read_attempt",
		FilePath:  filePath,
		Response:  string(fileContent),
	}
	
	// 直接发送结构化对象，而不是JSON字符串
	logger.LogReport.WithFields(logrus.Fields{
	    "api": "/api/mysql/",
	    "time": fileReadLog.Time,
	    "ip": fileReadLog.IP,
	    "service": fileReadLog.Service,
	    "event": fileReadLog.Event,
	    "file_path": fileReadLog.FilePath,
	    "response": fileReadLog.Response,
	}).Info("MySQL文件读取尝试")
	
	// 本地记录一下，方便调试
	logger.Log.Infof("MySQL文件读取尝试: IP=%s, 文件=%s", attackerIP, filePath)
	
	// 发送文件内容响应
	conn.Write(fileContent)
}
package mysql

import (
	"encoding/json"
	"fmt"
	"honeypot/server/pkg/logger"
	"honeypot/server/pkg/util"
	"io"
	"net"
	"time"
)

type MySQLProxy struct {
    frontendAddr string
    backendAddr  string
    isProxy      bool
}

func NewMySQLProxy(frontendAddr, backendAddr string, isProxy bool) *MySQLProxy {
    return &MySQLProxy{
        frontendAddr: frontendAddr,
        backendAddr:  backendAddr,
        isProxy:      isProxy,
    }
}

func (p *MySQLProxy) Start() error {
    listener, err := net.Listen("tcp", p.frontendAddr)
    if err != nil {
        return err
    }
    defer listener.Close()

    for {
        clientConn, err := listener.Accept()
        if err != nil {
            logger.Log.Errorf("Accept error: %v", err)
            continue
        }
        go p.handleConnection(clientConn)
    }
}

func (p *MySQLProxy) handleConnection(clientConn net.Conn) {
    defer clientConn.Close()

    // 连接到实际的MySQL服务
    serverConn, err := net.Dial("tcp", p.backendAddr)
    if err != nil {
        logger.Log.Errorf("Cannot connect to backend: %v", err)
        return
    }
    defer serverConn.Close()

    // 记录连接信息
    p.logConnection(clientConn)

    // 双向转发数据
    go p.pipe(clientConn, serverConn)
    p.pipe(serverConn, clientConn)
}

func (p *MySQLProxy) pipe(dst, src net.Conn) {
    buffer := make([]byte, 4096)
    for {
        n, err := src.Read(buffer)
        if err != nil {
            if err != io.EOF {
                logger.Log.Errorf("Read error: %v", err)
            }
            return
        }

        // 记录查询日志
        if p.isProxy {
            p.logQuery(src, buffer[:n])
        }

        _, err = dst.Write(buffer[:n])
        if err != nil {
            logger.Log.Errorf("Write error: %v", err)
            return
        }
    }
}

func (p *MySQLProxy) logConnection(conn net.Conn) {
    var attackerIP net.IP
    if p.isProxy {
        attackerIP = util.GetRawIpByConn(conn)
    } else {
        if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
            attackerIP = addr.IP
        }
    }

    accessLog := struct {
        Time     time.Time `json:"time"`
        IP       string    `json:"ip"`
        Service  string    `json:"service"`
        Event    string    `json:"event"`
    }{
        Time:    time.Now(),
        IP:      attackerIP.String(),
        Service: "mysql",
        Event:   "connection",
    }

    // 直接发送结构化日志
    jsonData, err := json.Marshal(accessLog)
    if err == nil {
        logger.LogReport.WithField("api", "/api/mysql/").Info(string(jsonData))
    }
}

func (p *MySQLProxy) logQuery(conn net.Conn, data []byte) {
    var attackerIP net.IP
    if p.isProxy {
        attackerIP = util.GetRawIpByConn(conn)
    } else {
        if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
            attackerIP = addr.IP
        }
    }

    // 尝试解析MySQL查询命令
    var queryCommand string
    if len(data) > 5 {
        // 简单解析，实际生产环境可能需要更复杂的解析
        queryCommand = string(data[5:])
    } else {
        queryCommand = fmt.Sprintf("未知查询 (数据长度: %d)", len(data))
    }

    queryLog := struct {
        Time     time.Time `json:"time"`
        IP       string    `json:"ip"`
        Service  string    `json:"service"`
        Event    string    `json:"event"`
        Command  string    `json:"command"`
        DataLen  int       `json:"data_length"`
    }{
        Time:     time.Now(),
        IP:       attackerIP.String(),
        Service:  "mysql",
        Event:    "query",
        Command:  queryCommand,
        DataLen:  len(data),
    }

    // 直接发送结构化日志
    jsonData, err := json.Marshal(queryLog)
    if err == nil {
        logger.LogReport.WithField("api", "/api/mysql/").Info(string(jsonData))
    }
}
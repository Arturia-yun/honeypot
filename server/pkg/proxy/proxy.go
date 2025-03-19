package proxy

import (
    "fmt"
    "io"
    "net"
    "time"
    "honeypot/server/pkg/logger"
    "honeypot/server/pkg/util"
    "honeypot/server/pkg/vars"
)

var (
    // 从配置文件获取地址和端口
    sshLocalAddr     = vars.GlobalConfig.Services["ssh"].ListenAddr
    sshBackendAddr   = fmt.Sprintf("127.0.0.1:%d", vars.GlobalConfig.Services["ssh"].BackendPort)
    
    mysqlLocalAddr   = vars.GlobalConfig.Services["mysql"].ListenAddr
    mysqlBackendAddr = fmt.Sprintf("127.0.0.1:%d", vars.GlobalConfig.Services["mysql"].BackendPort)
    
    redisLocalAddr   = vars.GlobalConfig.Services["redis"].ListenAddr
    redisBackendAddr = fmt.Sprintf("127.0.0.1:%d", vars.GlobalConfig.Services["redis"].BackendPort)
    
    webLocalAddr     = vars.GlobalConfig.Services["web"].ListenAddr
    webBackendAddr   = fmt.Sprintf("127.0.0.1:%d", vars.GlobalConfig.Services["web"].BackendPort)
)

// StartProxy 启动所有代理服务
func StartProxy() {
    // 确保地址非空
    if sshLocalAddr != "" {
        go serveProxy(sshLocalAddr, sshBackendAddr)
    }
    if mysqlLocalAddr != "" {
        go serveProxy(mysqlLocalAddr, mysqlBackendAddr)
    }
    if redisLocalAddr != "" {
        go serveProxy(redisLocalAddr, redisBackendAddr)
    }
    if webLocalAddr != "" {
        go serveProxy(webLocalAddr, webBackendAddr)
    }

    // 定期清理过期IP
    go func() {
        for {
            time.Sleep(10 * time.Second)
            util.DelExpireIps(300)
        }
    }()
}

// serveProxy 启动单个代理服务
func serveProxy(localAddr, backendAddr string) {
    listener, err := net.Listen("tcp", localAddr)
    if err != nil {
        logger.Log.Errorf("Failed to start proxy on %s: %v", localAddr, err)
        return
    }
    defer listener.Close()

    logger.Log.Infof("Proxy listening on %s, forwarding to %s", localAddr, backendAddr)

    for {
        clientConn, err := listener.Accept()
        if err != nil {
            logger.Log.Errorf("Failed to accept connection: %v", err)
            continue
        }

        go handleConnection(clientConn, backendAddr)
    }
}

// handleConnection 处理单个连接
func handleConnection(clientConn net.Conn, backendAddr string) {
    defer clientConn.Close()

    // 读取攻击者IP（前4个字节）
    ipBytes := make([]byte, 4)
    if _, err := io.ReadFull(clientConn, ipBytes); err != nil {
        logger.Log.Errorf("Failed to read attacker IP: %v", err)
        return
    }

    // 解析攻击者IP
    attackerIP := util.GetIPFromBytes(ipBytes)
    if attackerIP == nil {
        logger.Log.Error("Invalid attacker IP")
        return
    }

    // 保存攻击者IP
    vars.RawIps.Store(attackerIP.String(), util.IPInfo{
        IP:        attackerIP,
        Timestamp: time.Now(),
    })

    // 连接后端服务
    backendConn, err := net.Dial("tcp", backendAddr)
    if err != nil {
        logger.Log.Errorf("Failed to connect to backend %s: %v", backendAddr, err)
        return
    }
    defer backendConn.Close()

    // 启动双向数据转发
    go handlePipe(clientConn, backendConn)
    handlePipe(backendConn, clientConn)
}

// handlePipe 处理数据转发
func handlePipe(src, dst net.Conn) {
    buffer := make([]byte, 4096)
    for {
        n, err := src.Read(buffer)
        if err != nil {
            return
        }
        if n > 0 {
            if _, err := dst.Write(buffer[:n]); err != nil {
                return
            }
        }
    }
}
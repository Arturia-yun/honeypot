package forward

import (
    "encoding/binary"
    "fmt"
    "net"
    "sync"
    "honeypot/Agent/pkg/logger"
    "honeypot/Agent/pkg/policy"
	"honeypot/Agent/pkg/vars"
    "strings"
)

type ForwardServer struct {
    mu        sync.Mutex
    listeners map[int]net.Listener
}

func NewForwardServer() *ForwardServer {
    return &ForwardServer{
        listeners: make(map[int]net.Listener),
    }
}

func (fs *ForwardServer) Start() error {
    // 获取策略配置
    policy := policy.GetPolicy()
    if policy == nil {
        return fmt.Errorf("no policy loaded")
    }

    // 为每个服务启动转发
    for _, service := range policy.Service {
        if err := fs.startService(service); err != nil {
            return fmt.Errorf("failed to start service %s: %v", service.ServiceName, err)
        }
    }

    return nil
}

// startService 启动单个服务的监听
func (fs *ForwardServer) startService(service vars.BackendService) error {
    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", service.LocalPort))
    if err != nil {
        return err
    }

    fs.mu.Lock()
    fs.listeners[service.LocalPort] = listener
    fs.mu.Unlock()

    go fs.handleConnections(listener, service)
    return nil
}

// handleConnections 处理新的连接
func (fs *ForwardServer) handleConnections(listener net.Listener, service vars.BackendService) error {
    for {
        clientConn, err := listener.Accept()
        if err != nil {
            logger.Log.Errorf("Accept error: %v", err)
            continue
        }

        go fs.handleConnection(clientConn, service)
    }
}

// handleConnection 处理单个连接
func (fs *ForwardServer) handleConnection(clientConn net.Conn, service vars.BackendService) {
    defer clientConn.Close()

    // 连接到后端服务
    backendAddr := formatAddress(service.BackendHost, service.BackendPort)
    backendConn, err := net.Dial("tcp", backendAddr)
    if err != nil {
        logger.Log.Errorf("Failed to connect to backend: %v", err)
        return
    }
    defer backendConn.Close()

    // 获取客户端IP
    clientIP := getIPFromAddr(clientConn.RemoteAddr())
    
    // 将客户端IP转换为4字节数据
    ipBytes := convertIPToBytes(clientIP)
    
    // 发送IP到后端
    if _, err := backendConn.Write(ipBytes); err != nil {
        logger.Log.Errorf("Failed to write IP header: %v", err)
        return
    }

    // 启动双向数据转发
    var wg sync.WaitGroup
    wg.Add(2)

    // 客户端 -> 后端
    go func() {
        defer wg.Done()
        transfer(clientConn, backendConn)
    }()

    // 后端 -> 客户端
    go func() {
        defer wg.Done()
        transfer(backendConn, clientConn)
    }()

    wg.Wait()
}

// Stop 停止所有转发服务
func (fs *ForwardServer) Stop() {
    fs.mu.Lock()
    defer fs.mu.Unlock()

    for _, listener := range fs.listeners {
        listener.Close()
    }
}

// 辅助函数

func transfer(src, dst net.Conn) {
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

func getIPFromAddr(addr net.Addr) net.IP {
    tcpAddr, ok := addr.(*net.TCPAddr)
    if !ok {
        return nil
    }
    return tcpAddr.IP
}

func convertIPToBytes(ip net.IP) []byte {
    ip = ip.To4()
    if ip == nil {
        return make([]byte, 4)
    }
    
    bytes := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes, binary.BigEndian.Uint32(ip))
    return bytes
}

// 添加新的辅助函数
func formatAddress(host string, port int) string {
    // 检查是否是IPv6地址
    if strings.Contains(host, ":") {
        return fmt.Sprintf("[%s]:%d", host, port)
    }
    return fmt.Sprintf("%s:%d", host, port)
}
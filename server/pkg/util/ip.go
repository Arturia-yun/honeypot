package util

import (
    "net"
    "time"
    "honeypot/server/pkg/vars"
    "fmt"
    
)

type IPInfo struct {
    IP        net.IP
    Timestamp time.Time
}

// GetIPFromBytes 从字节数组中解析IP地址
func GetIPFromBytes(data []byte) net.IP {
    if len(data) < 4 {
        return nil
    }
    return net.IPv4(data[0], data[1], data[2], data[3])
}

// DelExpireIps 清理过期的IP地址
func DelExpireIps(expireSeconds int64) {
    now := time.Now()
    vars.RawIps.Range(func(key, value interface{}) bool {
        if ipInfo, ok := value.(IPInfo); ok {
            if now.Sub(ipInfo.Timestamp).Seconds() > float64(expireSeconds) {
                vars.RawIps.Delete(key)
            }
        }
        return true
    })
}

// GetRawIp 从RawIps中获取真实IP
func GetRawIp(remoteAddr, localAddr string) net.IP {
    key := fmt.Sprintf("%v_%v", remoteAddr, localAddr)
    if value, ok := vars.RawIps.Load(key); ok {
        if ipInfo, ok := value.(IPInfo); ok {
            return ipInfo.IP
        }
    }
    return nil
}

// GetRawIpByConn 通过连接获取真实IP
func GetRawIpByConn(conn net.Conn) net.IP {
    if conn == nil {
        return nil
    }
    return GetRawIp(conn.RemoteAddr().String(), conn.LocalAddr().String())
}
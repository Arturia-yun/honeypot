package models

import (
    "encoding/json"
    "time"
)

type PacketInfo struct {
    Protocol    string    `json:"protocol"`
    SrcIP       string    `json:"src_ip"`
    SrcPort     string    `json:"src_port"`
    DstIP       string    `json:"dst_ip"`
    DstPort     string    `json:"dst_port"`
    IsHTTP      bool      `json:"is_http"`
    CreateTime  time.Time `json:"create_time"`
}

func NewPacketInfo(connInfo *ConnectionInfo, createTime time.Time) *PacketInfo {
    return &PacketInfo{
        Protocol:   connInfo.Protocol,
        SrcIP:      connInfo.SrcIP,
        SrcPort:    connInfo.SrcPort,
        DstIP:      connInfo.DstIP,
        DstPort:    connInfo.DstPort,
        IsHTTP:     connInfo.IsHTTP,
        CreateTime: createTime,
    }
}

func (p *PacketInfo) String() (string, error) {
    bytes, err := json.Marshal(p)
    if err != nil {
        return "", err
    }
    return string(bytes), nil
}
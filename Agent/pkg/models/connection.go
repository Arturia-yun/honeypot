package models

type ConnectionInfo struct {
    Protocol    string `json:"protocol"`
    SrcIP       string `json:"src_ip"`
    SrcPort     string `json:"src_port"`
    DstIP       string `json:"dst_ip"`
    DstPort     string `json:"dst_port"`
    IsHTTP      bool   `json:"is_http"`
}

func NewConnectionInfo(protocol, srcIP, srcPort, dstIP, dstPort string, isHTTP bool) *ConnectionInfo {
    return &ConnectionInfo{
        Protocol: protocol,
        SrcIP:    srcIP,
        SrcPort:  srcPort,
        DstIP:    dstIP,
        DstPort:  dstPort,
        IsHTTP:   isHTTP,
    }
}
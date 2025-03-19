package capture

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"honeypot/Agent/pkg/logger"
	"honeypot/Agent/pkg/models"
	"honeypot/Agent/pkg/policy"
	"strings"
	"time"
)

type PacketCapture struct {
	handle *pcap.Handle
	iface  string
}

func NewPacketCapture(iface string) (*PacketCapture, error) {
	handle, err := pcap.OpenLive(iface, 65535, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}

	return &PacketCapture{
		handle: handle,
		iface:  iface,
	}, nil
}

func (pc *PacketCapture) Start() error {
	packetSource := gopacket.NewPacketSource(pc.handle, pc.handle.LinkType())
	for packet := range packetSource.Packets() {
		go pc.processPacket(packet)
	}
	return nil
}

func (pc *PacketCapture) Stop() {
	if pc.handle != nil {
		pc.handle.Close()
	}
}

// SplitPortService 从端口字符串中提取端口号
func SplitPortService(portStr string) string {
	parts := strings.Split(portStr, "(")
	if len(parts) > 0 {
		return parts[0]
	}
	return portStr
}

// IsInWhite 检查连接是否在白名单中
func IsInWhite(info *models.ConnectionInfo) bool {
	currentPolicy := policy.GetPolicy()
	if currentPolicy == nil || len(currentPolicy.Policy) == 0 {
		return false
	}

	// 检查IP白名单
	for _, p := range currentPolicy.Policy {
		for _, ip := range p.WhiteIps {
			if ip == info.SrcIP {
				return true
			}
		}
		// 检查端口白名单
		for _, port := range p.WhitePorts {
			if port == info.DstPort {
				return true
			}
		}
	}
	return false
}

// CheckSelfPacker 检查是否是自身产生的数据包
func CheckSelfPacker(info *models.ConnectionInfo) bool {
	// TODO: 实现自身数据包检查逻辑
	return false
}

func (pc *PacketCapture) processPacket(packet gopacket.Packet) {
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, ok := ipLayer.(*layers.IPv4)
		if ok {
			switch ip.Protocol {
			case layers.IPProtocolTCP:
				tcpLayer := packet.Layer(layers.LayerTypeTCP)
				if tcpLayer != nil {
					tcp, _ := tcpLayer.(*layers.TCP)

					srcPort := SplitPortService(tcp.SrcPort.String())
					dstPort := SplitPortService(tcp.DstPort.String())
					isHttp := false

					applicationLayer := packet.ApplicationLayer()
					if applicationLayer != nil {
						if strings.Contains(string(applicationLayer.Payload()), "HTTP") {
							isHttp = true
						}
					}

					connInfo := models.NewConnectionInfo(
						"tcp",
						ip.SrcIP.String(),
						srcPort,
						ip.DstIP.String(),
						dstPort,
						isHttp,
					)

					go func(info *models.ConnectionInfo) {
						if !IsInWhite(info) &&
							!CheckSelfPacker(info) &&
							(tcp.SYN && !tcp.ACK) {
							err := SendPacker(info)
							logger.Log.Debugf("[TCP] %v:%v -> %v:%v, err: %v",
								ip.SrcIP, tcp.SrcPort.String(),
								ip.DstIP, tcp.DstPort.String(), err)
						}
					}(connInfo)
				}
			}
		}
	}
}

// SendPacker 发送数据包信息到日志服务器
func SendPacker(info *models.ConnectionInfo) error {
    packetInfo := models.NewPacketInfo(info, time.Now())
    jsonPacket, err := packetInfo.String()
    if err != nil {
        return err
    }
    
    go logger.LogReport.WithField("api", "/api/packet/").Info(jsonPacket)
    return nil
}

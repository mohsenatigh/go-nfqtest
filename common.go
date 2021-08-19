package main

import (
	"net"
	"time"
)

//public utility functions

const DEFAULT_NET = "0.0.0.0/0"
const MAX_FILE_SIZE = 40960000
const (
	PROTOCOL_TCP = 6
	PROTOCOL_UDP = 17
	PROTOCOL_ANY = 256
)

func GetProtocolNumber(protocolName string) uint16 {
	if protocolName == "tcp" {
		return PROTOCOL_TCP
	} else if protocolName == "udp" {
		return PROTOCOL_UDP
	}
	return PROTOCOL_ANY
}

func GetProtocolName(protocolNum uint8) string {
	if protocolNum == PROTOCOL_TCP {
		return "tcp"
	} else if protocolNum == PROTOCOL_UDP {
		return "udp"
	}
	return "any"
}

// packet data structure and protocols
type SPacket struct {
	SIp       net.IP `json:"source"`
	DIp       net.IP `json:"destination"`
	Protocol  uint8  `json:"protocol"`
	IpVersion uint8  `json:"ip_version"`
	DataSize  uint16 `json:"data_size"`
}

const (
	PacketProcessResultOK   = 0
	PacketProcessResultDrop = 1
)

// packet providers common interface.
type IPacketProvider interface {
	Start() error
	Stop() error
	Dump() string
}

// conversations protocol info and utility functions
type SConversationProtocolStatus struct {
	Send      uint64 `json:"send"`
	Receive   uint64 `json:"receive"`
	StartTime int64  `json:"start_time"`
}

func (thisPt SConversationProtocolStatus) TotalData() uint64 {
	return thisPt.Send + thisPt.Receive
}

func (thisPt SConversationProtocolStatus) Duration() int64 {
	if thisPt.StartTime == 0 {
		return 0
	}
	return (time.Now().Unix() - thisPt.StartTime)
}

// conversations status tracker and utility functions
const (
	ConversationDirectionSend    = 0
	ConversationDirectionReceive = 1
)

type SConversationStatus struct {
	SrcIP       net.IP                      `json:"src_ip"`
	DstIP       net.IP                      `json:"dst_ip"`
	TCPStatus   SConversationProtocolStatus `json:"tcp"`
	UDPStatus   SConversationProtocolStatus `json:"udp"`
	OtherStatus SConversationProtocolStatus `json:"other"`
}

func (thisPt SConversationStatus) Duration() int64 {
	MAX := func(A int64, B int64) int64 {
		if A > B {
			return A
		}
		return B
	}
	duration := MAX(MAX(thisPt.TCPStatus.Duration(), thisPt.UDPStatus.Duration()), thisPt.OtherStatus.Duration())
	return duration
}

func (thisPt SConversationStatus) TotalData() uint64 {
	return thisPt.TCPStatus.TotalData() + thisPt.UDPStatus.TotalData() + thisPt.OtherStatus.TotalData()
}

func (thisPt SConversationStatus) Direction(packet *SPacket) int {
	if packet.DIp.Equal(thisPt.DstIP) {
		return ConversationDirectionSend
	}
	return ConversationDirectionReceive
}

//conversation tracker interface
type IConversationTracker interface {
	GetStatus(packet *SPacket, timeStamp int64) (bool, SConversationStatus)
	Dump() string
}

// common rules data structure
type SRule struct {
	Name        string `json:"name"`
	Destination string `json:"destination"`
	UsageTime   string `json:"usage_time"`
	UsageSize   string `json:"usage_size"`
	L4Protocol  string `json:"protocol"`
}

//rules repository
type IRuleRepository interface {
	GetRules() []SRule
}

// rule matchers common interface
type IRuleMatcher interface {
	Match(packet *SPacket, timeStamp int64) (int, string)
}

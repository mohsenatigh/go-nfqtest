package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Telefonica/nfqueue"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type SNFQStatus struct {
	Totalpackets uint64 `json:"total_packets"`
	Blocked      uint64 `json:"blocked"`
}

//---------------------------------------------------------------------------------------
//tun packet provider implement IPacketProvider
type CNFQPacketProvider struct {
	queue          *nfqueue.Queue
	matcher        IRuleMatcher
	queueNum       uint16
	gwMode         bool
	runIPTCommands bool
	stat           SNFQStatus
}

//---------------------------------------------------------------------------------------
func (thisPt *CNFQPacketProvider) execCommand(command string) error {
	//double space will cause  problem
	strList := strings.Split(strings.ReplaceAll(command, "  ", " "), " ")
	if _, err := exec.Command("iptables", strList...).Output(); err != nil {
		return err
	}
	return nil
}

//---------------------------------------------------------------------------------------
func (thisPt *CNFQPacketProvider) setIpTablesRule(add bool) error {

	if !thisPt.runIPTCommands {
		return nil
	}

	op := "-A"
	if !add {
		op = "-D"
	}

	if thisPt.gwMode {
		return thisPt.execCommand(fmt.Sprintf("%s FORWARD -j NFQUEUE --queue-num %d", op, thisPt.queueNum))
	}

	if err := thisPt.execCommand(fmt.Sprintf("%s INPUT -j NFQUEUE --queue-num %d", op, thisPt.queueNum)); err != nil {
		return err
	}

	return thisPt.execCommand(fmt.Sprintf("%s OUTPUT -j NFQUEUE --queue-num %d", op, thisPt.queueNum))
}

//---------------------------------------------------------------------------------------

func (thisPt *CNFQPacketProvider) processPacket(data []byte) (bool, SPacket) {

	out := SPacket{}

	layer := layers.LayerTypeIPv4
	if (data[0] & 0xf0) == 0x60 {
		layer = layers.LayerTypeIPv6
	}

	out.DataSize = uint16(len(data))

	lpacket := gopacket.NewPacket(data, layer, gopacket.NoCopy)
	network := lpacket.NetworkLayer()

	if network.LayerType() == layers.LayerTypeIPv6 {
		ipv6 := network.(*layers.IPv6)
		out.SIp = ipv6.SrcIP
		out.DIp = ipv6.DstIP
		out.IpVersion = 6
		out.Protocol = uint8(ipv6.NextHeader)
	} else if network.LayerType() == layers.LayerTypeIPv4 {
		ipv4 := network.(*layers.IPv4)
		out.SIp = ipv4.SrcIP.To4()
		out.DIp = ipv4.DstIP.To4()
		out.IpVersion = 4
		out.Protocol = uint8(ipv4.Protocol)
	} else {
		return false, out
	}
	return true, out
}

//---------------------------------------------------------------------------------------
// implement  IPacketProvider.Start
func (thisPt *CNFQPacketProvider) Start() error {
	//setup iptables, first remove existing rules
	thisPt.setIpTablesRule(false)
	if err := thisPt.setIpTablesRule(true); err != nil {
		return err
	}
	go thisPt.queue.Start()
	time.Sleep(100 * time.Millisecond)
	return nil
}

//---------------------------------------------------------------------------------------
// implement  IPacketProvider.Stop
func (thisPt *CNFQPacketProvider) Stop() error {
	if err := thisPt.setIpTablesRule(false); err != nil {
		return err
	}
	return thisPt.queue.Stop()
}

//---------------------------------------------------------------------------------------
// implement  nfqueue.PacketHandler
func (thisPt *CNFQPacketProvider) Handle(p *nfqueue.Packet) {

	if res, packet := thisPt.processPacket(p.Buffer); res {
		if thisPt.matcher != nil {
			thisPt.stat.Totalpackets++
			if res, _ := thisPt.matcher.Match(&packet, 0); res == PacketProcessResultDrop {
				thisPt.stat.Blocked++
				p.Drop()
			}
		}
	}
	p.Accept()
}

//---------------------------------------------------------------------------------------
func (thisPt *CNFQPacketProvider) Dump() string {
	out, _ := json.Marshal(thisPt.stat)
	return string(out)
}

//---------------------------------------------------------------------------------------
//NFQUEUE provider factory function

func CreateNFQProvider(queueNum uint16, gwMode bool, runIPTCommands bool, matcher IRuleMatcher) IPacketProvider {

	provider := new(CNFQPacketProvider)

	queueCfg := &nfqueue.QueueConfig{
		MaxPackets: 2048,
		BufferSize: 1600000,
		QueueFlags: []nfqueue.QueueFlag{nfqueue.FailOpen},
	}

	provider.queue = nfqueue.NewQueue(queueNum, provider, queueCfg)

	provider.queueNum = queueNum
	provider.gwMode = gwMode
	provider.runIPTCommands = runIPTCommands
	provider.matcher = matcher
	return provider
}

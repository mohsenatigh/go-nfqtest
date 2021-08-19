package main

import (
	"net"
	"testing"
	"time"
)

func TestMatcher(t *testing.T) {

	//create rule repos
	rules := `
	{
		"rules":[
			{
				"name":"test1",
				"destination":"192.168.1.0/24",
				"usage_time":"1h",
				"usage_size":"2kb",
				"protocol" : "any"
			},
			{
				"name":"test2",
				"destination":"192.168.2.0/24",
				"usage_time":"1h",
				"protocol" : "udp"
			},
			{
				"name":"default",
				"destination":"0.0.0.0/0",
				"usage_size":"256mb",
				"protocol" : "udp"
			}
		]
	}
	`
	//send packet
	spacket := SPacket{}
	spacket.SIp = net.ParseIP("192.168.0.1").To4()
	spacket.DIp = net.ParseIP("192.168.1.2").To4()
	spacket.Protocol = PROTOCOL_TCP
	spacket.IpVersion = 4
	spacket.DataSize = 1000

	//receive packet
	rpacket := spacket
	rpacket.SIp, rpacket.DIp = rpacket.DIp, rpacket.SIp

	repos := CreateJsonRuleRepositoryFromStr(rules)
	conv := CreateConversationTracker(3600, 2048)
	matcher := CreateMatcher(repos, conv)

	checkSenario := func(packet *SPacket, policyName string, result int, timeStamp int64) {
		res, name := matcher.Match(packet, timeStamp)
		if name != policyName || res != result {
			t.Fatal("match failed")
		}
	}

	//check send packet
	checkSenario(&spacket, "test1", PacketProcessResultOK, 0)

	//check receive packet
	checkSenario(&rpacket, "test1", PacketProcessResultOK, 0)

	//check data usage
	checkSenario(&rpacket, "test1", PacketProcessResultDrop, 0)

	//check no packet matching
	spacket.DIp = net.ParseIP("192.168.2.1")
	checkSenario(&spacket, "", PacketProcessResultOK, 0)

	//check time based policy
	spacket.Protocol = PROTOCOL_UDP
	checkSenario(&spacket, "test2", PacketProcessResultDrop, time.Now().Unix()-3700)

	//check default policy
	spacket.DIp = net.ParseIP("192.168.3.1")
	checkSenario(&spacket, "default", PacketProcessResultOK, 0)

}

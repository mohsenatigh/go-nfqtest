package main

import (
	"log"
	"net"
	"testing"
	"time"
)

func TestConversationTracker(t *testing.T) {

	conv := CreateConversationTracker(3600, 64)
	//create dummy packet
	packet := SPacket{}

	packet.SIp = net.ParseIP("192.168.1.1").To4()
	packet.DIp = net.ParseIP("192.168.1.2").To4()
	packet.DataSize = 50
	packet.IpVersion = 4
	packet.Protocol = PROTOCOL_TCP

	//get status

	res, stat := conv.GetStatus(&packet, 0)
	if !res {
		t.Fatal("test failed")
	}

	if !stat.SrcIP.Equal(packet.SIp) || !stat.DstIP.Equal(packet.DIp) {
		t.Fatal("invalid stat info")
	}

	if stat.TCPStatus.TotalData() != uint64(packet.DataSize) {
		t.Fatal("invalid stat info")
	}

	if stat.Direction(&packet) != ConversationDirectionSend {
		t.Fatal("invalid stat info")
	}

	//check receive
	packet.SIp, packet.DIp = packet.DIp, packet.SIp
	res, stat = conv.GetStatus(&packet, 0)
	if !res {
		t.Fatal("test failed")
	}

	if stat.Direction(&packet) != ConversationDirectionReceive {
		t.Fatal("invalid stat info")
	}

	if stat.TCPStatus.TotalData() != uint64(packet.DataSize*2) {
		t.Fatal("invalid stat info")
	}

	convInt := conv.(*CConversationTracker)

	if convInt.hashLinkList.GetItemsCount() != 1 {
		t.Fatal("invalid item count")
	}

	log.Printf("\n %s \n", conv.Dump())

	//check for auto remove
	for i := 0; i < HashBucketSize; i++ {
		convInt.checkForRemove(time.Now().Unix() + 40000)
	}

	if convInt.hashLinkList.GetItemsCount() != 0 {
		t.Fatal("invalid item count")
	}

}

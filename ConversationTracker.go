package main

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
	"time"
)

const HashBucketSize = 256000

//---------------------------------------------------------------------------------------
type CConversationTracker struct {
	hashLinkList cHashLinkList
	maxItems     uint32
}

//---------------------------------------------------------------------------------------
//remove inactive conversation
func (thisPt *CConversationTracker) checkForRemove(ctime int64) {
	thisPt.hashLinkList.CheckForTimeOut(nil, 0, ctime)
}

//---------------------------------------------------------------------------------------
func (thisPt *CConversationTracker) startTimeOutProcess() {
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			thisPt.checkForRemove(time.Now().Unix())
		}
	}()
}

//---------------------------------------------------------------------------------------
func (thisPt *CConversationTracker) getKey(packet *SPacket) uint64 {
	var srcIpIndex uint64
	var dstIpIndex uint64

	if packet.IpVersion == 4 {
		srcIpIndex = uint64(binary.LittleEndian.Uint32([]byte(packet.SIp)))
		dstIpIndex = uint64(binary.LittleEndian.Uint32([]byte(packet.DIp)))
	} else {
		calcV6Index := func(ip net.IP) uint64 {
			data := []byte(ip)
			p1 := binary.LittleEndian.Uint64(data[:8])
			p2 := binary.LittleEndian.Uint64(data[8:])
			return (p1 ^ p2)
		}
		srcIpIndex = calcV6Index(packet.SIp)
		dstIpIndex = calcV6Index(packet.DIp)
	}

	key := (srcIpIndex ^ dstIpIndex)
	return key
}

//---------------------------------------------------------------------------------------
func (thisPt *CConversationTracker) updateStat(info *SConversationStatus, packet *SPacket, timeStamp int64) {
	var stat *SConversationProtocolStatus
	if packet.Protocol == PROTOCOL_TCP {
		stat = &info.TCPStatus
	} else if packet.Protocol == PROTOCOL_UDP {
		stat = &info.UDPStatus
	} else {
		stat = &info.OtherStatus
	}

	//update time stamp
	if stat.StartTime == 0 {
		if timeStamp == 0 {
			timeStamp = time.Now().Unix()
		}
		stat.StartTime = timeStamp
	}

	if info.Direction(packet) == ConversationDirectionSend {
		stat.Send += uint64(packet.DataSize)
	} else {
		stat.Receive += uint64(packet.DataSize)
	}
}

//---------------------------------------------------------------------------------------
func (thisPt *CConversationTracker) createNew(packet *SPacket, timeStamp int64) (bool, SConversationStatus) {
	//get conversation key
	key := thisPt.getKey(packet)

	//check for max track table
	if thisPt.hashLinkList.GetItemsCount() > thisPt.maxItems {
		log.Printf("conversation table is full \n")
		return false, SConversationStatus{}
	}

	status := new(SConversationStatus)
	status.SrcIP = packet.SIp
	status.DstIP = packet.DIp
	thisPt.hashLinkList.Add(key, status)
	thisPt.updateStat(status, packet, timeStamp)
	return true, *status
}

//---------------------------------------------------------------------------------------
// implement  IConversationTracker.GetStatus
func (thisPt *CConversationTracker) GetStatus(packet *SPacket, timeStamp int64) (bool, SConversationStatus) {

	//get conversation key
	key := thisPt.getKey(packet)

	//add or update
	if data := thisPt.hashLinkList.Find(key, nil, nil); data != nil {
		status := data.(*SConversationStatus)
		thisPt.updateStat(status, packet, timeStamp)
		return true, *status
	}
	return thisPt.createNew(packet, timeStamp)
}

//---------------------------------------------------------------------------------------
// implement  IConversationTracker.Dump
func (thisPt *CConversationTracker) Dump() string {

	/*
		For simplicity, we ignore any search query and maximum row count
	*/

	out := []SConversationStatus{}

	callBack := func(inHashData interface{}) bool {
		status := inHashData.(*SConversationStatus)
		out = append(out, *status)
		return true
	}
	thisPt.hashLinkList.Iterate(callBack)

	//convert to json
	jsonRes, _ := json.Marshal(out)
	return string(jsonRes)
}

//---------------------------------------------------------------------------------------
//create tracker object
func CreateConversationTracker(inactivityTimeOut int64, maxItems uint32) IConversationTracker {

	tracker := new(CConversationTracker)
	if !tracker.hashLinkList.Init(HashBucketSize, inactivityTimeOut) {
		log.Fatalln("can not init hash linklist")
	}

	tracker.maxItems = maxItems
	tracker.hashLinkList.minInActiveTime = inactivityTimeOut

	//init inactive conversations remove goroutine
	tracker.startTimeOutProcess()

	return tracker
}

//---------------------------------------------------------------------------------------

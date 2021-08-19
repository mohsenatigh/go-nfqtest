package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

//---------------------------------------------------------------------------------------
type sCompiledRule struct {
	Name      string
	DataLimit int64
	TimeLimit int64
	Protocol  uint16
	Network   string
}

type sCompiledRulesList []sCompiledRule

//---------------------------------------------------------------------------------------
type CRuleMatcher struct {
	defaultRules        sCompiledRulesList
	accessLock          sync.RWMutex
	ruleParseRegx       *regexp.Regexp
	ipTri               cIPTrie
	ruleRepos           IRuleRepository
	conversationTracker IConversationTracker
}

//---------------------------------------------------------------------------------------
func (thisPt *CRuleMatcher) getValues(item string) (int64, string, error) {
	//var re =

	out := thisPt.ruleParseRegx.FindAllStringSubmatch(item, -1)
	if len(out[0]) != 3 {
		return 0, "", errors.New("invalid value")
	}

	nVal, _ := strconv.Atoi(out[0][1])

	return int64(nVal), strings.ToLower(out[0][2]), nil
}

//---------------------------------------------------------------------------------------
//convert SRule to sCompiledRules
func (thisPt *CRuleMatcher) compileRule(rule SRule) (sCompiledRule, error) {
	cmpRule := sCompiledRule{}

	cmpRule.Name = rule.Name
	cmpRule.Network = rule.Destination
	cmpRule.Protocol = GetProtocolNumber(rule.L4Protocol)

	//check destination
	if _, _, err := net.ParseCIDR(rule.Destination); err != nil {
		if ip, err := net.ResolveIPAddr("ip", rule.Destination); err != nil {
			return cmpRule, errors.New("invalid network")
		} else {
			cmpRule.Network = fmt.Sprintf("%s/32", ip.String())
		}
	}

	cmpRule.TimeLimit = -1
	cmpRule.DataLimit = -1

	//process data
	if len(rule.UsageSize) > 0 {
		if data, unit, err := thisPt.getValues(rule.UsageSize); err == nil {
			if unit == "kb" {
				cmpRule.DataLimit = int64(data * 1024)
			} else if unit == "mb" {
				cmpRule.DataLimit = int64(data * 1024 * 1024)
			} else if unit == "gb" {
				cmpRule.DataLimit = int64(data * 1024 * 1024 * 1024)
			} else {
				return cmpRule, errors.New("invalid usage unit")
			}
		} else {
			return cmpRule, errors.New("invalid usage value")
		}
	}

	//process time
	if len(rule.UsageTime) > 0 {
		if data, unit, err := thisPt.getValues(rule.UsageTime); err == nil {
			if unit == "s" {
				cmpRule.TimeLimit = data
			} else if unit == "m" {
				cmpRule.TimeLimit = int64(data * 60)
			} else if unit == "h" {
				cmpRule.TimeLimit = int64(data * 3600)
			} else {
				return cmpRule, errors.New("invalid time unit")
			}
		} else {
			return cmpRule, errors.New("invalid time value")
		}
	}

	return cmpRule, nil
}

//---------------------------------------------------------------------------------------
//return all the possible rules for a network
func (thisPt *CRuleMatcher) findRule(ip net.IP, protocol uint16) (bool, sCompiledRule) {

	//find best matched rule
	findBestRuleInRuleList := func(list *sCompiledRulesList, protocol uint16) (bool, sCompiledRule) {
		best := -1
		for i, rule := range *list {
			if rule.Protocol == protocol {
				best = i
				break
			} else if rule.Protocol == PROTOCOL_ANY {
				best = i
			}
		}
		if best != -1 {
			return true, (*list)[best]
		}
		return false, sCompiledRule{}
	}

	//check ip TRI
	if ruleListIn := thisPt.ipTri.Search(ip); ruleListIn != nil {
		return findBestRuleInRuleList(ruleListIn.(*sCompiledRulesList), protocol)
	}

	//check for default rules
	return findBestRuleInRuleList(&thisPt.defaultRules, protocol)
}

//---------------------------------------------------------------------------------------
func (thisPt *CRuleMatcher) loadRules() error {

	thisPt.accessLock.Lock()
	defer thisPt.accessLock.Unlock()

	//We should first make sure about the correctness of the rules. After that, we can free existing rules and add them to the Tri
	rules := thisPt.ruleRepos.GetRules()
	cmpRules := []sCompiledRule{}
	for _, r := range rules {
		if cmpRule, err := thisPt.compileRule(r); err != nil {
			return err
		} else {
			cmpRules = append(cmpRules, cmpRule)
		}
	}

	//check for duplicate rules for a subnet
	checkForDuplicate := func(ruleList *sCompiledRulesList, protocol uint16) bool {
		for _, r := range *ruleList {
			if r.Protocol == protocol {
				return true
			}
		}
		return false
	}

	//every thing seems good :)
	defaultRules := sCompiledRulesList{}
	thisPt.ipTri.Flush()
	for _, cmp := range cmpRules {

		//We may have different rules for each protocol in a subnet for example 192.168.1.0:udp and 192.168.1.0:tcp or 192.168.1.0:any
		var ruleList *sCompiledRulesList

		//default policy
		if cmp.Network == DEFAULT_NET {
			ruleList = &defaultRules
		} else if listInter := thisPt.ipTri.SearchExactString(cmp.Network); listInter != nil {
			ruleList = listInter.(*sCompiledRulesList)
		} else {
			ruleList = new(sCompiledRulesList)
			if err := thisPt.ipTri.AddString(cmp.Network, ruleList); err != nil {
				return err
			}
		}

		//check for duplicate rules
		if checkForDuplicate(ruleList, cmp.Protocol) {
			return errors.New("duplicate rules detected")
		}

		*ruleList = append(*ruleList, cmp)
	}

	//set default routes
	thisPt.defaultRules = defaultRules

	return nil
}

//---------------------------------------------------------------------------------------
func (thisPt *CRuleMatcher) checkRule(packet *SPacket, rule *sCompiledRule, conversation *SConversationStatus) int {
	usage := conversation.TotalData()
	duration := conversation.Duration()

	if rule.Protocol == PROTOCOL_TCP {
		usage = conversation.TCPStatus.TotalData()
		duration = conversation.TCPStatus.Duration()
	} else if rule.Protocol == PROTOCOL_UDP {
		usage = conversation.UDPStatus.TotalData()
		duration = conversation.UDPStatus.Duration()
	}

	if rule.TimeLimit != -1 && duration >= rule.TimeLimit {
		return PacketProcessResultDrop
	}

	if rule.DataLimit != -1 && usage >= uint64(rule.DataLimit) {
		return PacketProcessResultDrop
	}

	return PacketProcessResultOK
}

//---------------------------------------------------------------------------------------
func (thisPt *CRuleMatcher) Match(packet *SPacket, timeStamp int64) (int, string) {

	thisPt.accessLock.RLock()
	defer thisPt.accessLock.RUnlock()

	//get active conversation
	fnd, status := thisPt.conversationTracker.GetStatus(packet, timeStamp)
	if !fnd {
		//can not find any conversation, usually because the conversation table is full
		return PacketProcessResultOK, ""
	}

	/*
		To improve the performance of the system it is possible to save the rule info into the conversation.
		This will reduce the per-packet rule matching overhead.
		to keep the solution simple I ignored this mechanism
	*/

	//find active rule
	ip := packet.DIp
	if status.Direction(packet) == ConversationDirectionReceive {
		ip = packet.SIp
	}

	fnd, rule := thisPt.findRule(ip, uint16(packet.Protocol))
	if !fnd {
		return PacketProcessResultOK, ""
	}

	//check rule against the conversation info
	return thisPt.checkRule(packet, &rule, &status), rule.Name
}

//---------------------------------------------------------------------------------------

func CreateMatcher(ruleRepos IRuleRepository, conversation IConversationTracker) IRuleMatcher {
	matcher := new(CRuleMatcher)
	matcher.ruleParseRegx = regexp.MustCompile(`(?m)(\d+)(\w{1,2})`)
	matcher.conversationTracker = conversation
	matcher.ruleRepos = ruleRepos
	//just IPV4
	matcher.ipTri.Init(4)

	if err := matcher.loadRules(); err != nil {
		log.Fatalln(err)
	}
	return matcher
}

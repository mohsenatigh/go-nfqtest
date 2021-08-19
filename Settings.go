package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type SSettings struct {
	MaxConversations                uint32 `json:"max_conversation"`
	MaxInactiveConversationLifeTime uint32 `json:"max_inactive_conversation_life_time"`
	NFQueueNumber                   uint16 `json:"nfq_number"`
	GWMode                          bool   `json:"gw_mode"`
	RunIPCommands                   bool   `json:"run_iptables_command"`
}

func LoadSettings(fileName string) (SSettings, error) {
	set := SSettings{}

	//fill defaults
	set.MaxConversations = 64000
	set.MaxInactiveConversationLifeTime = 3600 //second

	set.GWMode = false
	set.NFQueueNumber = 64
	set.RunIPCommands = true

	if stat, err := os.Stat(fileName); err != nil || stat.Size() > MAX_FILE_SIZE {
		log.Fatalln(err)
	}

	//load
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalln(err)
	}

	if err := json.Unmarshal(data, &set); err != nil {
		log.Fatalln(err)
	}

	//for simplicity there is not any settings checking

	return set, nil
}

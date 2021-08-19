package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	settingFile := flag.String("f", "", "configuration file")
	flag.Parse()

	if len(*settingFile) < 1 {
		log.Fatalf("please define valid config file \n")
	}

	//load setting
	settings, err := LoadSettings(*settingFile)
	if err != nil {
		log.Fatalln(err)
	}

	//create rules repository
	ruleRespos := CreateJsonRuleRepository(*settingFile)

	//create conversation tracker
	conversation := CreateConversationTracker(int64(settings.MaxInactiveConversationLifeTime), settings.MaxConversations)

	//create rule matcher
	ruleMatcher := CreateMatcher(ruleRespos, conversation)

	//create packet provider
	packetProvider := CreateNFQProvider(settings.NFQueueNumber, settings.GWMode, settings.RunIPCommands, ruleMatcher)

	//start provider
	if err := packetProvider.Start(); err != nil {
		log.Fatalln(err)
	}

	//start API server
	CreateApiServer(conversation, packetProvider)

	log.Printf("simplefw started successfully \n")

	//wait for termination
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	packetProvider.Stop()
	time.Sleep(1 * time.Second)
	log.Printf("successfully terminated\n")

}

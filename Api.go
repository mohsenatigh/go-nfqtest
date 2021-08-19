package main

import (
	"net/http"
)

type CApi struct {
	conversation IConversationTracker
	provider     IPacketProvider
}

//---------------------------------------------------------------------------------------
func (thisPt *CApi) dumpProvider(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(thisPt.provider.Dump()))
}

//---------------------------------------------------------------------------------------
func (thisPt *CApi) dumpConversations(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(thisPt.conversation.Dump()))
}

//---------------------------------------------------------------------------------------
func (thisPt *CApi) serve() {
	http.HandleFunc("/conversations", thisPt.dumpConversations)
	http.HandleFunc("/provider", thisPt.dumpProvider)
	http.ListenAndServe("127.0.0.1:8080", nil)
}

//---------------------------------------------------------------------------------------
func CreateApiServer(conv IConversationTracker, provider IPacketProvider) {
	api := CApi{}
	api.conversation = conv
	api.provider = provider
	go api.serve()
}

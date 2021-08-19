package main

import (
	"testing"
)

func TestNFQ(t *testing.T) {

	nfq := CreateNFQProvider(64, false, true, nil)
	if err := nfq.Start(); err != nil {
		t.Fatal(err)
	}

	nfq.Stop()
}

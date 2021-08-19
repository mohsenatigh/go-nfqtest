package main

import (
	"testing"
)

func TestRuleRep(t *testing.T) {
	ruleRep := CreateJsonRuleRepository("data/setting.json")
	if ruleRep == nil {
		t.Fatalf("can not load rules")
	}

	if len(ruleRep.GetRules()) < 1 {
		t.Fatalf("invalid rules")
	}

}

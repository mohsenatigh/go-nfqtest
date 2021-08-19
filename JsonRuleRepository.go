package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
)

//---------------------------------------------------------------------------------------
type SRuleList []SRule

type CJsonRuleRepository struct {
	Rules SRuleList `json:"rules"`
}

func (thisPt *CJsonRuleRepository) loadRulesFromString(rules string) error {
	tempObj := CJsonRuleRepository{}
	if err := json.Unmarshal([]byte(rules), &tempObj); err != nil {
		return err
	}
	thisPt.Rules = tempObj.Rules
	return nil
}

//---------------------------------------------------------------------------------------
func (thisPt *CJsonRuleRepository) loadRules(fileName string) error {
	stat, err := os.Stat(fileName)
	if err != nil {
		return err
	}

	if stat.Size() > MAX_FILE_SIZE {
		return errors.New("invalid rule file size")
	}

	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	return thisPt.loadRulesFromString(string(buf))
}

//---------------------------------------------------------------------------------------
// implement  IRuleRepository.GetRules
func (thisPt *CJsonRuleRepository) GetRules() []SRule {
	return thisPt.Rules
}

//---------------------------------------------------------------------------------------

func CreateJsonRuleRepository(fileName string) IRuleRepository {
	ruleRep := new(CJsonRuleRepository)
	if err := ruleRep.loadRules(fileName); err != nil {
		log.Fatalln(err)
	}
	return ruleRep
}

//---------------------------------------------------------------------------------------
func CreateJsonRuleRepositoryFromStr(data string) IRuleRepository {
	ruleRep := new(CJsonRuleRepository)
	if err := ruleRep.loadRulesFromString(data); err != nil {
		log.Fatalln(err)
	}
	return ruleRep
}

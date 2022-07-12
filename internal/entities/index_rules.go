package entities

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

type IndexRules map[string]map[string]struct{}

func NewIndexRules(cfgKey string, cfg *viper.Viper) (res IndexRules, err error) {
	res = IndexRules{}
	for ruleType, params := range cfg.Get(cfgKey).(map[string]interface{}) {
		if ruleType != "only" && ruleType != "except" {
			continue
		}
		for key, items := range params.(map[string]interface{}) {
			if _, ok := res["only-"+key]; ruleType == "except" && ok {
				continue
			}
			index := make(map[string]struct{}, len(items.([]interface{})))
			for _, val := range items.([]interface{}) {
				if val := strings.TrimSpace(strings.ToLower(val.(string))); val != "" {
					index[val] = struct{}{}
				}
			}
			if len(index) > 0 {
				res[ruleType+"-"+key] = index
				if ruleType == "only" {
					delete(res, "except-"+key)
				}
			}
		}
	}

	return
}

func (r IndexRules) Check(book *Book) error {
	if !r.checkOnly(r["only-langs"], book.Info.Lang) || !r.checkExcept(r["except-langs"], book.Info.Lang) {
		return errors.New("rule langs")
	}

	items := book.Genres()
	if !r.checkOnly(r["only-genres"], items...) || !r.checkExcept(r["except-genres"], items...) {
		return errors.New("rule genres")
	}

	items = book.Authors()
	if !r.checkOnly(r["only-authors"], items...) || !r.checkExcept(r["except-authors"], items...) {
		return errors.New("rule authors")
	}

	items = book.Titles()
	if !r.checkOnly(r["only-titles"], items...) || !r.checkExcept(r["except-titles"], items...) {
		return errors.New("rule title")
	}

	return nil
}

func (r IndexRules) checkOnly(want map[string]struct{}, got ...string) bool {
	if len(want) == 0 {
		return true
	}

	gotStr := strings.ToLower(strings.Join(got, " "))
	if gotStr == "" {
		return true
	}

	for rule := range want {
		if strings.Contains(gotStr, rule) {
			return true
		}
	}

	return false
}

func (r IndexRules) checkExcept(want map[string]struct{}, got ...string) bool {
	if len(want) == 0 {
		return true
	}

	gotStr := strings.ToLower(strings.Join(got, " "))
	if gotStr == "" {
		return true
	}

	for rule := range want {
		if strings.Contains(gotStr, rule) {
			return false
		}
	}

	return true
}

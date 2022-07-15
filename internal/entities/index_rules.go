package entities

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type IndexRules map[string]map[string]struct{}

func NewIndexRules(cfgKey string, cfg *viper.Viper) (res IndexRules, err error) {
	var ruleType string
	res = IndexRules{}

	for field, values := range cfg.GetStringMapStringSlice(cfgKey) {
		for _, val := range values {
			if ruleType = "only"; val[0:1] == "-" {
				ruleType = "except"
			}

			if val = strings.TrimSpace(strings.ToLower(strings.TrimLeft(val, "+-"))); val == "" {
				continue
			}

			if _, ok := res[field+"-"+ruleType]; !ok {
				res[field+"-"+ruleType] = map[string]struct{}{}
			}

			res[field+"-"+ruleType][val] = struct{}{}
		}

		if len(res[field+"-only"]) > 0 && len(res[field+"-except"]) > 0 {
			for k := range res[field+"-except"] {
				delete(res[field+"-only"], k)
			}

			delete(res, field+"-except")
		}
	}

	return
}

func (r IndexRules) Check(book *Book) error {
	index := book.Index()
	for _, field := range []IndexField{IdxFLang, IdxFGenre, IdxFISBN, IdxFAuthor, IdxFTranslator, IdxFSerie, IdxFDate, IdxFPublisher, IdxFTitle} {
		if err := r.matchRules(field, &index); err != nil {
			return err
		}
	}

	return nil
}

func (r IndexRules) matchRules(field IndexField, index *BookIndex) (err error) {
	var val string

	switch field {
	case IdxFLang:
		if val = index.Lang; index.Lang == "" {
			val = "ru"
		}
	case IdxFGenre:
		val = index.Genre
	case IdxFISBN:
		val = index.ISBN
	case IdxFAuthor:
		val = index.Author
	case IdxFTranslator:
		val = index.Translator
	case IdxFSerie:
		val = index.Serie
	case IdxFDate:
		val = index.Date
	case IdxFPublisher:
		val = index.Publisher
	case IdxFTitle:
		val = index.Title
	default:
		return fmt.Errorf("invalid rule field type - %v", field)
	}

	val = strings.ToLower(strings.TrimSpace(val))

	if !r.checkOnly(field, val) {
		return fmt.Errorf("rule %s: %s", field, val)
	}

	if !r.checkExcept(field, val) {
		err = fmt.Errorf("rule %s: %s", field, val)
	}

	return
}

func (r IndexRules) checkOnly(rule IndexField, got string) bool {
	if len(r[string(rule)+"-only"]) == 0 {
		return true
	}

	if got == "" {
		return false
	}

	for rule := range r[string(rule)+"-only"] {
		if strings.Contains(got, rule) {
			return true
		}
	}

	return false
}

func (r IndexRules) checkExcept(rule IndexField, got string) bool {
	if len(r[string(rule)+"-except"]) == 0 {
		return true
	}

	if got == "" {
		return true
	}

	for rule := range r[string(rule)+"-except"] {
		if strings.Contains(got, rule) {
			return false
		}
	}

	return true
}

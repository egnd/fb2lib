package xmlparse

// Rule middleware for TokenHandler.
type Rule func(next TokenHandler) TokenHandler

// WrapRules builds chain of rules.
func WrapRules(rules []Rule, handler TokenHandler) TokenHandler {
	if len(rules) == 0 {
		return handler
	}

	for i := len(rules) - 1; i >= 0; i-- {
		handler = rules[i](handler)
	}

	return handler
}

package fb2parse

import (
	"encoding/xml"
)

// TokenHandler handler for fb2 tokens.
type TokenHandler func(section interface{}, node xml.StartElement, r xml.TokenReader) error

// HandlingRule middleware for TokenHandler.
type HandlingRule func(next TokenHandler) TokenHandler

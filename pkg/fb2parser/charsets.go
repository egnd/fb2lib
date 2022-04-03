package fb2parser

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// The actual FB2 XML encodings in the wild.
var charsetsMap = map[string]encoding.Encoding{
	"utf-8":        unicode.UTF8,
	"utf8":         unicode.UTF8,
	"windows-1251": charmap.Windows1251,
	"windows-1252": charmap.Windows1252,
	"koi8-r":       charmap.KOI8R,
	"iso-8859-1":   charmap.ISO8859_1,
	"iso-8859-5":   charmap.ISO8859_5,
	"utf-16":       unicode.UTF16(unicode.LittleEndian, unicode.UseBOM),
	"utf-16le":     unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
	"utf-16be":     unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM),
}

// CharsetReader returns a new charset-conversion reader, converting from the provided charset into UTF-8.
func CharsetReader(label string, input io.Reader) (io.Reader, error) {
	encoder := charsetsMap[strings.ToLower(strings.TrimSpace(label))]
	if encoder == nil {
		return nil, fmt.Errorf("fb2 charset error: unsupported charset %s", label)
	}

	return transform.NewReader(input, encoder.NewDecoder()), nil
}

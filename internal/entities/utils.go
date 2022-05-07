package entities

import (
	"crypto/md5"
	"encoding/hex"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	regexpYearPattern = regexp.MustCompile("[0-9]{3,}")
	currentYear       = time.Now().Year()
)

func ParseYear(date string) (res uint16) {
	if date == "" {
		return
	}

	for _, year := range regexpYearPattern.FindAllString(date, -1) {
		if len(year) <= 4 && !strings.HasPrefix(year, "0") {
			val, _ := strconv.ParseUint(year, 10, 16)
			res = uint16(val)

			if res > uint16(currentYear) {
				res = 0
			}

			break
		}
	}

	return
}

func GenerateID(args ...[]string) string {
	hasher := md5.New()

	for _, vals := range args {
		sort.Strings(vals)

		for _, str := range vals {
			str = strings.TrimSpace(str)

			if str != "" {
				hasher.Write([]byte(str))
			}
		}
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

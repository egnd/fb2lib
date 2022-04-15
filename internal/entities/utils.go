package entities

import (
	"regexp"
	"strings"
)

var (
	regexpYearPattern = regexp.MustCompile("[12][0-9]{3}")
)

func parseYear(date string) string {
	if len(date) < 4 {
		return ""
	}

	yearsIdx := map[string]struct{}{}
	years := make([]string, 0, 2)

	for _, year := range regexpYearPattern.FindAllString(date, -1) {
		if _, ok := yearsIdx[year]; ok || len(years) > 1 {
			continue
		}

		yearsIdx[year] = struct{}{}
		years = append(years, year)
	}

	return strings.Join(years, "-")
}

package response

import (
	"fmt"
	"strings"
)

func BuildBookURL(path, urlPrefix, pathPrefix string) string {
	return fmt.Sprintf("%s/%s",
		strings.Trim(urlPrefix, "/"),
		strings.Trim(strings.TrimPrefix(path, pathPrefix), "/"),
	)
}

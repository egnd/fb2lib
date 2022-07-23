package xmlparse

// GetStrFrom returns first not empty string from slice.
func GetStrFrom(items []string) string {
	for _, item := range items {
		if item != "" {
			return item
		}
	}

	return ""
}

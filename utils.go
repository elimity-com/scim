package scim

func contains(arr []string, el string) bool {
	for _, item := range arr {
		if item == el {
			return true
		}
	}

	return false
}

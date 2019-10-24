package scim

func contains(arr []string, el string) bool {
	for _, item := range arr {
		if item == el {
			return true
		}
	}

	return false
}

func clamp(offset, limit, length int) (int, int) {
	start := length
	if offset < length {
		start = offset
	}
	end := length
	if limit < length-start {
		end = start + limit
	}
	return start, end
}

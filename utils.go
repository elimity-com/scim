package scim

import (
	"bytes"
	"io"
	"net/http"
)

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

func contains(arr []string, el string) bool {
	for _, item := range arr {
		if item == el {
			return true
		}
	}

	return false
}

func readBody(r *http.Request) ([]byte, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(data))
	return data, nil
}

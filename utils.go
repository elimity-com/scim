package scim

import (
	"bytes"
	"io"
	"net/http"
	"slices"
)

func clamp(offset, limit, length int) (int, int) {
	start := min(offset, length)
	end := length
	if limit < length-start {
		end = start + limit
	}
	return start, end
}

func contains(arr []string, el string) bool {
	return slices.Contains(arr, el)
}

func readBody(r *http.Request) ([]byte, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(data))
	return data, nil
}

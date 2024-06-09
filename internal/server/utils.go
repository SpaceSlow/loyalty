package server

import (
	"io"
	"strconv"
)

func getOrderNumber(body io.ReadCloser) (int, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return -1, err
	}
	defer body.Close()
	return strconv.Atoi(string(data))
}

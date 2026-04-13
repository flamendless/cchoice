package services

import "bytes"

type closerBuffer struct {
	*bytes.Buffer
}

func (closerBuffer) Close() error { return nil }

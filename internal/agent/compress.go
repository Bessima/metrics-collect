package agent

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

func Compress(data []byte) (bytes.Buffer, error) {
	var b bytes.Buffer

	writer, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return b, fmt.Errorf("failed init compress writer: %v", err)
	}

	_, err = writer.Write(data)
	if err != nil {
		return b, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return b, fmt.Errorf("failed compress data: %v", err)
	}

	return b, nil
}

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func unmarshalData(r io.Reader, result any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	if len(data) == 0 {
		return errors.New("data is empty")
	}

	return json.Unmarshal(data, result)
}

func marshalData(data any) (io.Reader, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(jsonData), nil
}

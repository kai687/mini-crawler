package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// JSONLWriter writes records as newline-delimited JSON objects.
type JSONLWriter struct {
	writer *bufio.Writer
}

// NewJSONLWriter wraps an io.Writer with buffered JSONL output.
func NewJSONLWriter(w io.Writer) *JSONLWriter {
	return &JSONLWriter{writer: bufio.NewWriter(w)}
}

// Write serializes one record as a single JSONL line.
func (w *JSONLWriter) Write(record any) error {
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}

	if _, err := w.writer.Write(data); err != nil {
		return fmt.Errorf("write record: %w", err)
	}

	if err := w.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	return nil
}

// Close flushes buffered output.
func (w *JSONLWriter) Close() error {
	return w.writer.Flush()
}

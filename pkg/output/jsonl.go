package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// JSONLWriter writes records as newline-delimited JSON objects.
type JSONLWriter struct {
	writer  *bufio.Writer
	encoder *json.Encoder
}

const defaultBufferSize = 1 << 20

// NewJSONLWriter wraps an io.Writer with buffered JSONL output.
func NewJSONLWriter(w io.Writer) *JSONLWriter {
	writer := bufio.NewWriterSize(w, defaultBufferSize)

	return &JSONLWriter{
		writer:  writer,
		encoder: json.NewEncoder(writer),
	}
}

// Write serializes one record as a single JSONL line.
func (w *JSONLWriter) Write(record any) error {
	if err := w.encoder.Encode(record); err != nil {
		return fmt.Errorf("write record: %w", err)
	}

	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("flush record: %w", err)
	}

	return nil
}

// Close flushes buffered output.
func (w *JSONLWriter) Close() error {
	return w.writer.Flush()
}

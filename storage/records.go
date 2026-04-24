package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"mmth-etl/types"
	"os"
	"path/filepath"
	"sync"
)

// RecordsWriter handles appending records to JSONL files
type RecordsWriter struct {
	mu          sync.Mutex
	baseDir     string
	files       map[string]*bufio.Writer // key: item type (diamond, rune_ticket, etc.)
	fileHandles map[string]*os.File
}

// NewRecordsWriter creates a new records writer
func NewRecordsWriter(baseDir string) *RecordsWriter {
	return &RecordsWriter{
		baseDir:     baseDir,
		files:       make(map[string]*bufio.Writer),
		fileHandles: make(map[string]*os.File),
	}
}

// AppendRecord appends a record to the appropriate JSONL file
func (w *RecordsWriter) AppendRecord(itemType string, record types.ChangeRecord) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	writer, err := w.getWriter(itemType)
	if err != nil {
		return err
	}

	// Write record as JSON line
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("write record: %w", err)
	}
	if err := writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	return nil
}

// Flush flushes all buffered writers
func (w *RecordsWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for itemType, writer := range w.files {
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("flush %s: %w", itemType, err)
		}
	}
	return nil
}

// Close closes all file handles
func (w *RecordsWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for itemType, writer := range w.files {
		if err := writer.Flush(); err != nil {
			fmt.Printf("flush %s: %v\n", itemType, err)
		}
	}
	w.files = make(map[string]*bufio.Writer)

	for itemType, f := range w.fileHandles {
		if err := f.Close(); err != nil {
			fmt.Printf("close %s: %v\n", itemType, err)
		}
	}
	w.fileHandles = make(map[string]*os.File)

	return nil
}

func (w *RecordsWriter) getWriter(itemType string) (*bufio.Writer, error) {
	if writer, ok := w.files[itemType]; ok {
		return writer, nil
	}

	// Open file for appending
	filename := fmt.Sprintf("%s_records.jsonl", itemType)
	path := filepath.Join(w.baseDir, filename)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	writer := bufio.NewWriter(f)
	w.files[itemType] = writer
	w.fileHandles[itemType] = f

	return writer, nil
}

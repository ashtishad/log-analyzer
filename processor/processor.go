package processor

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"
)

const NumOfChunks = 4 // NumOfChunks = runtime.NumCPU()

// LogInfo represents a single log entry with parsed user activity data.
type LogInfo struct {
	UserID    int64
	PageName  string
	Timestamp time.Time
}

// filePart represents a section of the file to be processed independently.
type filePart struct {
	offset int64
	size   int64
}

// splitFile divides the file for concurrent processing while preserving log integrity.
// Uses direct file I/O operations for performance, avoiding full file read into memory.
func splitFile(filePath string, numParts int) ([]filePart, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := fileInfo.Size()
	partSize := fileSize / int64(numParts)

	var parts []filePart

	for i := 0; i < numParts; i++ {
		offset := int64(i) * partSize
		size := partSize

		if i == numParts-1 {
			size = fileSize - offset
		} else {
			_, err := file.Seek(offset+size, io.SeekStart)
			if err != nil {
				return nil, err
			}

			buf := make([]byte, 100)
			n, err := file.Read(buf)

			if err != nil && err != io.EOF {
				return nil, err
			}

			for j := 0; j < n; j++ {
				if buf[j] == '\n' {
					size += int64(j) + 1
					break
				}
			}
		}

		parts = append(parts, filePart{offset: offset, size: size})
	}

	return parts, nil
}

// ReadLogFile implements concurrent log processing for improved performance on multi-core systems.
// Uses goroutines and channels for parallel execution and result aggregation.
func ReadLogFile(ctx context.Context, filename string) ([]LogInfo, error) {
	parts, err := splitFile(filename, NumOfChunks)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	resultsCh := make(chan []LogInfo, NumOfChunks)
	errCh := make(chan error, 1)

	for _, part := range parts {
		wg.Add(1)

		go func(p filePart) {
			defer wg.Done()

			logs, err := processFilePart(filename, p.offset, p.size)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}

			resultsCh <- logs
		}(part)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
		close(errCh)
	}()

	var allLogs []LogInfo
	for logs := range resultsCh {
		allLogs = append(allLogs, logs...)
	}

	if err := <-errCh; err != nil {
		return nil, err
	}

	return allLogs, nil
}

// processFilePart optimizes file reading with buffered I/O and preallocated slices.
// Uses a scanner for efficient line-by-line processing, suitable for large files.
func processFilePart(filename string, offset, size int64) ([]LogInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err = file.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	reader := io.LimitReader(file, size)
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	logs := make([]LogInfo, 0, size/100)

	for scanner.Scan() {
		log, err := parseLogLine(scanner.Bytes())
		if err != nil {
			continue // Skip invalid lines for robustness
		}

		logs = append(logs, log)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// parseLogLine uses byte-level operations for parsing, avoiding the overhead of encoding/json.
// Tailored for the specific JSON structure: `{"userId": int, "pageName": string, "timestamp": ISO8601}`.
// Finds field positions using byte operations, more efficient than string manipulations
// Uses strconv for faster integer parsing and time.Parse for reliable timestamp parsing
// This approach is significantly faster than generic JSON parsing for this known structure.
func parseLogLine(line []byte) (LogInfo, error) {
	var log LogInfo

	if len(line) < 20 {
		return log, fmt.Errorf("line too short: %s", line)
	}

	userIDStart := bytes.IndexByte(line, ':') + 1
	if userIDStart <= 0 {
		return log, fmt.Errorf("invalid userID format: %s", line)
	}

	userIDEnd := bytes.IndexByte(line[userIDStart:], ',')
	if userIDEnd == -1 {
		return log, fmt.Errorf("invalid userID end: %s", line)
	}

	userIDEnd += userIDStart

	pageNameStart := bytes.Index(line[userIDEnd:], []byte(`:"`)) + userIDEnd + 2
	if pageNameStart <= userIDEnd+2 {
		return log, fmt.Errorf("invalid pageName start: %s", line)
	}

	pageNameEnd := bytes.IndexByte(line[pageNameStart:], '"')
	if pageNameEnd == -1 {
		return log, fmt.Errorf("invalid pageName end: %s", line)
	}

	pageNameEnd += pageNameStart

	timestampStart := bytes.LastIndex(line, []byte(`:"`)) + 2
	if timestampStart < pageNameEnd {
		return log, fmt.Errorf("invalid timestamp start: %s", line)
	}

	timestampEnd := len(line) - 2
	if timestampEnd <= timestampStart {
		return log, fmt.Errorf("invalid timestamp end: %s", line)
	}

	var err error
	log.UserID, err = strconv.ParseInt(string(line[userIDStart:userIDEnd]), 10, 64)

	if err != nil {
		return log, fmt.Errorf("failed to parse userID: %w", err)
	}

	log.PageName = string(line[pageNameStart:pageNameEnd])

	log.Timestamp, err = time.Parse(time.RFC3339, string(line[timestampStart:timestampEnd]))
	if err != nil {
		return log, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return log, nil
}

package proccessor

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"time"
)

// NumOfChunks defines the number of parts to split the file into for concurrent processing
const NumOfChunks = 4

// LogInfo represents a single log entry with user activity data.
// It's structured to match the JSON format of the log/slog entries.
type LogInfo struct {
	UserID    int64     `json:"userId"`
	PageName  string    `json:"pageName"`
	Timestamp time.Time `json:"timestamp"`
}

// filePart represents a section of the file to be processed independently.
type filePart struct {
	offset int64
	size   int64
}

// splitFile divides the file into parts for concurrent processing.
// It ensures each part ends on a complete line to maintain log integrity.
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

// ReadLogFile reads and parses a log file in parallel.
// It utilizes goroutines to process file parts concurrently, improving performance for large files.
func ReadLogFile(ctx context.Context, filename string) ([]LogInfo, error) {
	parts, err := splitFile(filename, NumOfChunks)
	if err != nil {
		return nil, err
	}

	resultsCh := make(chan []LogInfo, NumOfChunks)
	errCh := make(chan error, 1)

	for _, part := range parts {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			go func(p filePart) {
				logs, err := processFilePart(filename, p.offset, p.size)
				if err != nil {
					errCh <- err
					return
				}

				resultsCh <- logs
			}(part)
		}
	}

	var allLogs []LogInfo

	for i := 0; i < NumOfChunks; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errCh:
			return nil, err
		case logs := <-resultsCh:
			allLogs = append(allLogs, logs...)
		}
	}

	return allLogs, nil
}

// processFilePart reads and parses a portion of the log file.
// It uses a scanner for efficient line-by-line reading and json.Unmarshal for parsing.
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

	var logs []LogInfo

	for scanner.Scan() {
		var log LogInfo

		if err := json.Unmarshal(scanner.Bytes(), &log); err != nil {
			// skip invalid JSON lines
			continue
		}

		logs = append(logs, log)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

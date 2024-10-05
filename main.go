package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ashtishad/log-analyzer/seed"
)

// LogInfo represents a single log entry with user activity data.
type LogInfo struct {
	UserID    int64     `json:"userId"`
	PageName  string    `json:"pageName"`
	Timestamp time.Time `json:"timestamp"`
}

const (
	minLoyalPages = 4
	timeout       = 1000 * time.Millisecond
)

// File format: Timestamp, PageId, CustomerId
// every day new file(each has 10000 log entries)
// find "loyal customers" from 10000 users.
// (a) they came on both days.
// (b) they visited at least 4 unique pages.
func main() {
	if err := seed.GenerateLogFiles(); err != nil {
		slog.Error("failed to generate log files", "err", err)
		return
	}

	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	file1, file2, err := processLogs(ctx)
	if err != nil {
		fmt.Println("failed to process logs:", err)
		return
	}

	loyalCustomers, err := getLoyalCustomers(ctx, file1, file2)
	if err != nil {
		fmt.Println("failed to get loyal customers:", err)
		return
	}

	elapsed := time.Since(start)

	fmt.Printf("Loyal Customer Count: %d\n", len(loyalCustomers))
	fmt.Printf("Time elapsed: %v\n", elapsed)
	fmt.Printf("Loyal Customer IDs: %v\n", loyalCustomers)
}

// processLogs reads and parses log files for two consecutive days.
func processLogs(ctx context.Context) ([]LogInfo, []LogInfo, error) {
	file1, err := readLogFile(ctx, filepath.Join("seed", "files", "logs_2024-10-01.log"))
	if err != nil {
		return nil, nil, fmt.Errorf("error reading file 1: %w", err)
	}

	file2, err := readLogFile(ctx, filepath.Join("seed", "files", "logs_2024-10-02.log"))
	if err != nil {
		return nil, nil, fmt.Errorf("error reading file 2: %w", err)
	}

	return file1, file2, nil
}

// readLogFile reads a log file and returns a slice of LogInfo.
func readLogFile(ctx context.Context, filename string) ([]LogInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var logs []LogInfo

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			var log LogInfo

			if err := json.Unmarshal(scanner.Bytes(), &log); err != nil {
				return nil, err
			}

			logs = append(logs, log)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// getLoyalCustomers identifies loyal customers based on visit frequency and unique page views.
// Time Complexity: O(n log n), where n is the total number of log entries
// Space Complexity: O(m), where m is the number of unique users
func getLoyalCustomers(ctx context.Context, page1, page2 []LogInfo) ([]int64, error) {
	userPages := make(map[int64]struct {
		pages     map[string]bool
		visitDays int
	})

	processDayLogs := func(logs []LogInfo, day int) error {
		for _, log := range logs {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if _, ok := userPages[log.UserID]; !ok {
					userPages[log.UserID] = struct {
						pages     map[string]bool
						visitDays int
					}{pages: make(map[string]bool), visitDays: 0}
				}

				info := userPages[log.UserID]
				info.pages[log.PageName] = true

				if info.visitDays < day {
					info.visitDays = day
				}

				userPages[log.UserID] = info
			}
		}

		return nil
	}

	if err := processDayLogs(page1, 1); err != nil {
		return nil, err
	}

	if err := processDayLogs(page2, 2); err != nil {
		return nil, err
	}

	loyalCustomersMap := make(map[int64]bool)

	for userID, info := range userPages {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if info.visitDays == 2 && len(info.pages) >= minLoyalPages {
				loyalCustomersMap[userID] = true
			}
		}
	}

	loyalCustomers := make([]int64, 0, len(loyalCustomersMap))
	for userID := range loyalCustomersMap {
		loyalCustomers = append(loyalCustomers, userID)
	}

	sort.Slice(loyalCustomers, func(i, j int) bool {
		return loyalCustomers[i] < loyalCustomers[j]
	})

	return loyalCustomers, nil
}

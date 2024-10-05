package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/ashtishad/log-analyzer/proccessor"
)

const (
	minLoyalPages = 4
	timeout       = 200 * time.Millisecond
)

// UserPageInfo holds information about a user's page visits
type UserPageInfo struct {
	pages     map[string]bool
	visitDays int
}

// File format: Timestamp, PageId, CustomerId
// every day new file(each has 10000 log entries)
// find "loyal customers" from 10000 users.
// (a) they came on both days.
// (b) they visited at least 4 unique pages.
func main() {
	// if err := seed.GenerateLogFiles(); err != nil {
	// 	slog.Error("failed to generate log files", "err", err)
	// 	return
	// }

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
	fmt.Printf("Loyal Customers: %v\n", loyalCustomers)
}

// processLogs reads log files for two consecutive days.
// It utilizes concurrent processing for efficient handling of large log files.
func processLogs(ctx context.Context) ([]proccessor.LogInfo, []proccessor.LogInfo, error) {
	file1, err := proccessor.ReadLogFile(ctx, filepath.Join("seed", "files", "logs_2024-10-01.log"))
	if err != nil {
		return nil, nil, fmt.Errorf("error reading file 1: %w", err)
	}

	file2, err := proccessor.ReadLogFile(ctx, filepath.Join("seed", "files", "logs_2024-10-02.log"))
	if err != nil {
		return nil, nil, fmt.Errorf("error reading file 2: %w", err)
	}

	return file1, file2, nil
}

// getLoyalCustomers identifies loyal customers based on visit frequency and unique page views.
// It processes logs from two days and applies the loyalty criteria to generate a sorted list of loyal customer IDs.
func getLoyalCustomers(ctx context.Context, page1, page2 []proccessor.LogInfo) ([]int64, error) {
	userPages := make(map[int64]UserPageInfo)

	if err := processDayLogs(ctx, userPages, page1, 1); err != nil {
		return nil, err
	}

	if err := processDayLogs(ctx, userPages, page2, 2); err != nil {
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

// processDayLogs processes logs for a single day and updates the userPages map.
// It's designed to be called for each day's logs, updating visit counts and page views efficiently.
func processDayLogs(ctx context.Context, userPages map[int64]UserPageInfo, logs []proccessor.LogInfo, day int) error {
	for _, log := range logs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			info, ok := userPages[log.UserID]
			if !ok {
				info = UserPageInfo{pages: make(map[string]bool)}
			}

			info.pages[log.PageName] = true

			if info.visitDays < day {
				info.visitDays = day
			}

			userPages[log.UserID] = info
		}
	}

	return nil
}

package main

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/ashtishad/log-analyzer/seed"
)

func TestGetLoyalCustomers(t *testing.T) {
	day1 := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
	day2 := day1.AddDate(0, 0, 1)

	file1 := []proccessor.LogInfo{
		{UserID: 1, PageName: "blog", Timestamp: day1},
		{UserID: 1, PageName: "dashboard", Timestamp: day1},
		{UserID: 1, PageName: "shop", Timestamp: day1},
		{UserID: 2, PageName: "blog", Timestamp: day1},
		{UserID: 3, PageName: "blog", Timestamp: day1},
		{UserID: 3, PageName: "profile", Timestamp: day1},
		{UserID: 3, PageName: "shop", Timestamp: day1},
		{UserID: 4, PageName: "contact", Timestamp: day1},
		{UserID: 5, PageName: "blog", Timestamp: day1},
		{UserID: 6, PageName: "blog", Timestamp: day1},
		{UserID: 6, PageName: "profile", Timestamp: day1},
		{UserID: 7, PageName: "blog", Timestamp: day1},
	}

	file2 := []proccessor.LogInfo{
		{UserID: 1, PageName: "blog", Timestamp: day2},
		{UserID: 1, PageName: "about", Timestamp: day2},
		{UserID: 2, PageName: "blog", Timestamp: day2},
		{UserID: 3, PageName: "shop", Timestamp: day2},
		{UserID: 3, PageName: "contact", Timestamp: day2},
		{UserID: 4, PageName: "blog", Timestamp: day2},
		{UserID: 5, PageName: "blog", Timestamp: day2},
		{UserID: 6, PageName: "contact", Timestamp: day2},
		{UserID: 6, PageName: "shop", Timestamp: day1},
		{UserID: 8, PageName: "blog", Timestamp: day2},
	}

	expected := []int64{1, 3, 6}
	result, err := getLoyalCustomers(context.Background(), file1, file2)

	if err != nil {
		t.Errorf("getLoyalCustomers() returned an error: %v", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("getLoyalCustomers() = %v, want %v", result, expected)
	}
}

func FuzzGetLoyalCustomers(f *testing.F) {
	f.Add([]byte(`[{"userId":1,"pageName":"home","timestamp":"2024-10-01T00:00:00Z"}]`),
		[]byte(`[{"userId":1,"pageName":"shop","timestamp":"2024-10-02T00:00:00Z"}]`))

	f.Fuzz(func(t *testing.T, data1, data2 []byte) {
		var file1, file2 []proccessor.LogInfo

		if err := json.Unmarshal(data1, &file1); err != nil {
			return // invalid input, skip this iteration
		}
		if err := json.Unmarshal(data2, &file2); err != nil {
			return // invalid input, skip this iteration
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		result, err := getLoyalCustomers(ctx, file1, file2)

		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("getLoyalCustomers() returned an unexpected error: %v", err)
		}

		for _, userID := range result {
			found := false
			for _, log := range file1 {
				if log.UserID == userID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("User %d in result not found in file1", userID)
			}

			found = false
			for _, log := range file2 {
				if log.UserID == userID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("User %d in result not found in file2", userID)
			}

			uniquePages := make(map[string]bool)
			for _, log := range append(file1, file2...) {
				if log.UserID == userID {
					uniquePages[log.PageName] = true
				}
			}
			if len(uniquePages) < minLoyalPages {
				t.Errorf("User %d has fewer than %d unique pages", userID, minLoyalPages)
			}
		}
	})
}

func BenchmarkGetLoyalCustomers(b *testing.B) {
	if err := seed.GenerateLogFiles(); err != nil {
		b.Fatalf("Failed to generate log files: %v", err)
	}

	ctx := context.Background()
	file1, file2, err := processLogs(ctx)
	if err != nil {
		b.Fatalf("Failed to process logs: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := getLoyalCustomers(ctx, file1, file2)
		if err != nil {
			b.Fatalf("getLoyalCustomers() returned an error: %v", err)
		}
	}
}

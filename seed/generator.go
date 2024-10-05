package seed

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// LogEntry represents a single log entry with user activity data.
type LogEntry struct {
	UserID    int64     `json:"userId"`
	PageName  string    `json:"pageName"`
	Timestamp time.Time `json:"timestamp"`
}

const (
	totalUsers        = 10000
	totalEntries      = 10000
	loyalCustomerRate = 0.18 // 18% loyal customers
	minLoyalPages     = 4    // Minimum number of unique pages for loyal customers
)

var pages = []string{
	"home", "blog", "shop", "about", "contact", "profile", "dashboard",
	"products", "services", "faq", "support", "news", "events", "gallery",
	"forum", "reviews", "careers", "partners", "pricing", "testimonials",
}

// GenerateLogFiles creates two log files with simulated user activity data in log/slog format
func GenerateLogFiles() error {
	loyalCustomers := generateLoyalCustomers()

	if err := generateFile("logs_2024-10-01.log", time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC), loyalCustomers); err != nil {
		return err
	}

	if err := generateFile("logs_2024-10-02.log", time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC), loyalCustomers); err != nil {
		return err
	}

	return nil
}

// generateLoyalCustomers creates a map of loyal customer IDs,
// purpose is to ensure distinct loyal customer id's from non-loyal one's while log entry generation.
func generateLoyalCustomers() map[int64]bool {
	loyalCount := int(float64(totalUsers) * loyalCustomerRate)
	loyalCustomers := make(map[int64]bool, loyalCount)

	for len(loyalCustomers) < loyalCount {
		userID := rand.Int63n(totalUsers) + 1
		loyalCustomers[userID] = true
	}

	return loyalCustomers
}

// generateFile creates a log file with simulated user activity for a specific date.
// Generates entries for loyal customers, loyal customer has 6-6 pages visits
// Then generate entries for non-loyal customers separately.
func generateFile(filename string, date time.Time, loyalCustomers map[int64]bool) error {
	dir := filepath.Join("seed", "files")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	file, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	logger := log.New(file, "", 0)

	entriesGenerated := 0

	for userID := range loyalCustomers {
		pageCount := rand.Intn(3) + minLoyalPages
		userPages := getUniquePages(pages, pageCount)

		for _, page := range userPages {
			entry := LogEntry{
				UserID:    userID,
				PageName:  page,
				Timestamp: date.Add(time.Duration(rand.Intn(86400)) * time.Second),
			}

			jsonEntry, err := json.Marshal(entry)
			if err != nil {
				return fmt.Errorf("error marshaling log entry: %w", err)
			}

			logger.Println(string(jsonEntry))

			entriesGenerated++
		}
	}

	nonLoyalUsers := make([]int64, 0, totalUsers-len(loyalCustomers))

	for i := int64(1); i <= totalUsers; i++ {
		if !loyalCustomers[i] {
			nonLoyalUsers = append(nonLoyalUsers, i)
		}
	}

	for entriesGenerated < totalEntries {
		userID := nonLoyalUsers[rand.Intn(len(nonLoyalUsers))]
		page := pages[rand.Intn(len(pages))]

		entry := LogEntry{
			UserID:    userID,
			PageName:  page,
			Timestamp: date.Add(time.Duration(rand.Intn(86400)) * time.Second),
		}

		jsonEntry, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("error marshaling log entry: %w", err)
		}

		logger.Println(string(jsonEntry))

		entriesGenerated++
	}

	return nil
}

// getUniquePages returns a slice of unique page names, 20 pages
func getUniquePages(pages []string, count int) []string {
	if count > len(pages) {
		count = len(pages)
	}

	shuffled := make([]string, len(pages))
	copy(shuffled, pages)

	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}

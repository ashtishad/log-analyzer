# Log Analyzer for Loyal Customer Identification

A tool to identify the most loyal customers by analyzing their activity logs, utilizing concurrent processing using file chunks for improved performance.

## Input Data

- Daily logs of customer activity in a structured JSON format with log/slog.
- Log entry format: `{"userId": int, "pageName": string, "timestamp": ISO8601}`
- **1 million** log entries per day, **1 million** unique customers, 20 website pages/router group

## Loyal Customer Criteria

1. Visited on both days under analysis
2. Accessed at least 4 unique pages

# Expected Output
```
Loyal Customer Count: 180066
Time elapsed: 418.303584ms

// commented in main.go to avoid large output slice
Loyal Customers: [2 8 13 15 19 22 25 35 40 45 ....................... ]

```

## Implementation

### 1. Log Generation (seed/generator.go)

- Generate structured slog log files for consecutive days
- Provide a controlled dataset with mixed loyal and non-loyal customer activities

### 2. Concurrent Log Processing (processor/processor.go)

- Split files into chunks for parallel processing while preserving log integrity
- NumOfChunks can be set to runtime.NumCPU() to utilize all cores.
- Implement concurrent processing using goroutines and channels
- Optimize file reading with buffered I/O and preallocated slices
- Parse logs using fast byte-level operations, tailored for our JSON structure

### 3. Loyal Customer Identification Algorithm (main.go)

- Process log data from both days concurrently
- Track user visits and unique page views using a map for efficient lookups
- Apply loyal customer criteria: visits on both days and minimum unique pages
- Generate a sorted list of loyal customer IDs

**Further Reading:** [The One Billion Row Challenge in Go](https://benhoyt.com/writings/go-1brc/)

# Directory structure
```
log-analyzer/
│
├── .github/
│   └── workflows/
│       └── test.yaml          # CI workflow for automated testing
│
├── seed/
│   └── generator.go           # Generates sample log data for testing
│
├── files/
│   ├── logs_2024-10-01.log    # generated 1 million log entries from day 1
│   └── logs_2024-10-02.log    # generated 1 million log entries from day 2
│
├── proccessor/
│   └── processor.go           # Core log processing logic
│
├── main.go                    # Main application entry point
├── main_test.go               # Unit, Fuzz and Benchmark tests.
├── README.md                  # Project documentation
└── go.mod                     # Go module definition
```

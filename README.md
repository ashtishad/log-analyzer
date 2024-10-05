# Log Analyzer for Loyal Customer Identification

A tool to identify the most loyal customers by analyzing their activity logs, utilizing concurrent processing using file chunks for improved performance.

## Expected Output
```
Loyal Customer Count: 1800
Time elapsed: 21.90775ms
Loyal Customers: [2 8 13 15 19 22 25 35 40 47 ....................... ]
```

## Directory structure
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
│   ├── logs_2024-10-01.log    # Sample log file for day 1
│   └── logs_2024-10-02.log    # Sample log file for day 2
│
├── proccessor/
│   └── processor.go           # Core log processing logic
│
├── main.go                    # Main application entry point
├── main_test.go               # Unit, Fuzz and Benchmark tests.
├── README.md                  # Project documentation
└── go.mod                     # Go module definition
```

## Input Data

- Daily logs of customer activity in a structured JSON format
- Log entry format: `{"userId": int, "pageName": string, "timestamp": ISO8601}`
- 10,000 log entries per day, 10,000 unique customers, 20 pages

## Loyal Customer Criteria

1. Visited on both days under analysis
2. Accessed at least 4 unique pages

## Implementation

### 1. Log Generation (Seed Data)

**Purpose**: Create test data for development and testing.

- Generate two structured slog log files for consecutive days
- Include a mix of loyal and non-loyal customer activities
- Provide a controlled dataset for testing

### 2. Concurrent Log Processing

**Purpose**: Efficiently read and parse large log files.

- Split each log file into multiple chunks (default: 4)
- Process chunks concurrently using goroutines
- Merge results from all chunks
- Implement context-aware processing for timeout and cancellation handling

### 3. Log Reading and Parsing

**Purpose**: Extract structured data from log files.

- Open and read log file chunks
- Parse JSON-formatted log entries using `encoding/json`
- Convert log data into `LogInfo` structs
- Skip invalid JSON lines for robustness

### 4. Loyal Customer Identification Algorithm

**Purpose**: Apply loyalty criteria to identify loyal customers.

- Process log data from both days
- Track user visits and unique page views using a map
- Apply loyal customer criteria (visits on both days and minimum unique pages)
- Generate a sorted list of loyal customer IDs

**Complexity**:
- Time: O(n logn), where n is the total number of log entries
- Space: O(m), where m is the number of unique users

### 5. Performance Optimization

- Concurrent processing of log files for improved speed
- Use of efficient data structures (maps) for tracking user activity
- Separate processing logic for each day's logs

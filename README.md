# Log Analyzer for Loyal Customer Identification

A tool to identify the most loyal customers by analyzing their activity logs.

## Input Data

- Daily logs of customer activity in a structured format (slog/log)
- Log entry format: `Timestamp, PageId, CustomerId`
- 10,000 log entries per day
- 10,000 unique customers

## Loyal Customer Criteria

1. Visited on both days under analysis
2. Accessed at least 4 unique pages

## Implementation

### 1. Log Generation (Seed Data)

**Purpose**: Create test data for development and testing.

- Generate two log files for consecutive days
- Include a mix of loyal and non-loyal customer activities
- Provide a controlled dataset for testing

### 2. Log Processing

**Purpose**: Orchestrate the log analysis workflow.

- Read log files for two consecutive days
- Invoke log reading and analysis functions
- Handle errors and timeout scenarios
- Measure and report processing time

### 3. Log Reading

**Purpose**: Extract structured data from log files.

- Open and read log files
- Parse JSON-formatted log entries
- Convert log data into `LogInfo` structs
- Implement context-aware processing for timeout handling

### 4. Loyal Customer Identification Algorithm

**Purpose**: Apply loyalty criteria to identify loyal customers.

- Process log data from both days
- Track user visits and unique page views
- Apply loyal customer criteria
- Generate a sorted list of loyal customer IDs, their total count and time elapsed

**Complexity**:
- Time: O(n log n), where n is the total number of log entries
- Space: O(m), where m is the number of unique users

### 5. Testing Suite

- **Unit Tests**: `TestGetLoyalCustomers` for specific scenarios
- **Fuzz Tests**: `FuzzGetLoyalCustomers` for robust edge case testing
- **Benchmark Tests**: `BenchmarkGetLoyalCustomers` using seed data
  - Utilizes `GenerateLogFiles()` and `processLogs()` from seed package for realistic data processing

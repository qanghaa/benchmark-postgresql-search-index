# Benchmark Summary

This document provides a high-level summary of the performance benchmarks for the PostgreSQL search index project.

## Pre-setup

To run the benchmarks yourself, follow these steps:

1.  **Prerequisites**: Ensure you have Docker, Go (1.21+), and Make installed.
2.  **Start Database**:
    ```bash
    make docker-up
    ```
3.  **Run Migrations**:
    ```bash
    make migrate-up
    ```
4.  **Run Benchmark**:
    ```bash
    go run cmd/benchmark/main.go
    ```

## Terminology

We test with three different record sizes to simulate various real-world scenarios:

### Small Records
*   **Description**: Basic event logs with minimal fields.
*   **Size**: ~10 fields.
*   **Example**:
    ```json
    {
      "action_type": "view",
      "status": "success",
      "timestamp": 1700000000,
      "user_agent": "Mozilla/5.0..."
    }
    ```

### Medium Records
*   **Description**: Standard application logs with metadata, location, and device info.
*   **Size**: 50-100 fields.
*   **Example**:
    ```json
    {
      "action_type": "login",
      "city": "New York",
      "country": "US",
      "device_id": "device_123",
      "latency": 45,
      "plan": "premium",
      "region": "us-east-1",
      "status": "success"
    }
    ```

### Large Records
*   **Description**: Complex documents with extensive metadata, nested objects, arrays, and multi-byte characters (e.g., Japanese).
*   **Size**: 300-500 fields.
*   **Example**:
    ```json
    {
      "action_type": "purchase",
      "array_field_0": ["item1", "item2"],
      "description": "Long description with multi-byte characters...",
      "nested_obj_0": { "nested_field_0": "value" },
      "japanese_field_0": "こんにちは"
    }
    ```

## Benchmark Results

The following table summarizes the query performance for different dataset sizes and record types. All times are in milliseconds (ms).

| Dataset Size | Record Size | Test Case | Duration (ms) |
| :--- | :--- | :--- | :--- |
| **1,000** | Small | FTS Found | 18.12 |
| | | FTS Not Found | 1.46 |
| | | Partial Found | 7.46 |
| | Medium | FTS Found | 2.13 |
| | | FTS Not Found | 1.19 |
| | | Partial Found | 22.71 |
| | Large | FTS Found | *Running...* |
| **10,000** | Small | FTS Found | *Pending...* |

*Note: "FTS" stands for Full-Text Search (using GIN index). "Partial" refers to partial match searches (e.g., `LIKE '%term%'`).*

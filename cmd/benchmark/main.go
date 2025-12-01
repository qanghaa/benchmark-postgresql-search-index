package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"log-project/internal/db"
	"log-project/models"
	"log-project/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Config holds the benchmark configuration
type Config struct {
	DatasetSize int
	RecordSize  string // "Small", "Medium", "Large"
}

// Result holds the benchmark result for a single test case
type Result struct {
	TestCase string
	Duration time.Duration
}

func main() {
	ctx := context.Background()
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://loguser:logpassword@localhost:5435/logdb?sslmode=disable"
	}

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	queries := db.New(conn)

	datasetSizes := []int{1000, 10000}
	recordSizes := []string{
		"small",
		"medium",
		"large",
	}

	// Initialize tabwriter for output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Dataset\tRecordSize\tTestCase\tDuration")

	for _, size := range datasetSizes {
		for _, recordSize := range recordSizes {
			cfg := Config{
				DatasetSize: size,
				RecordSize:  recordSize,
			}
			log.Printf("Running benchmark for Dataset: %d, RecordSize: %s", size, recordSize)

			foundTerm, err := seedData(ctx, queries, cfg)
			if err != nil {
				log.Fatalf("Failed to seed data: %v", err)
			}

			results := runQueries(ctx, queries, cfg, foundTerm)
			for _, res := range results {
				fmt.Fprintf(w, "%d\t%s\t%s\t%v\n", size, recordSize, res.TestCase, res.Duration)
			}
			w.Flush()
			time.Sleep(2 * time.Second)
		}
	}
}

func seedData(ctx context.Context, q *db.Queries, cfg Config) (string, error) {
	// Truncate table first
	if err := q.TruncateLogs(ctx); err != nil {
		return "", fmt.Errorf("failed to truncate logs: %w", err)
	}

	batchSize := 1000
	var batch []db.BulkInsertLogsParams
	var lastContent models.Content

	for i := 0; i < cfg.DatasetSize; i++ {
		content := utils.GenerateSampleContent(cfg.RecordSize)
		lastContent = content
		contentBytes, _ := json.Marshal(content)

		batch = append(batch, db.BulkInsertLogsParams{
			UserID:    pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Domain:    "example.com",
			Action:    "login",
			Content:   contentBytes,
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		})

		if len(batch) >= batchSize {
			if _, err := q.BulkInsertLogs(ctx, batch); err != nil {
				return "", fmt.Errorf("failed to bulk insert: %w", err)
			}
			batch = nil
		}
	}

	if len(batch) > 0 {
		if _, err := q.BulkInsertLogs(ctx, batch); err != nil {
			return "", fmt.Errorf("failed to bulk insert: %w", err)
		}
	}

	// Return a value from the last content to search for
	// Prefer description or notes if available, otherwise fallback
	if val, ok := lastContent["description"].(string); ok && val != "" {
		return val, nil
	}
	if val, ok := lastContent["user_agent"].(string); ok && val != "" {
		return val, nil
	}

	return "login", nil
}

func runQueries(ctx context.Context, q *db.Queries, cfg Config, foundTerm string) []Result {
	var results []Result

	// Define search terms
	notFoundTerm := uuid.New().String()
	shortTerm := "lo"
	runes := []rune(foundTerm)
	if len(runes) >= 2 {
		shortTerm = string(runes[:2])
	}

	// Helper to measure execution
	measure := func(name string, fn func() error) {
		start := time.Now()
		err := fn()
		duration := time.Since(start)
		if err != nil {
			log.Printf("Error in %s: %v", name, err)
		}
		results = append(results, Result{TestCase: name, Duration: duration})
	}

	// 1. FTS - Found
	measure("FTS Found", func() error {
		_, err := q.ListLogsWithFilters(ctx, db.ListLogsWithFiltersParams{
			Limit:         100,
			Offset:        0,
			ContentSearch: pgtype.Text{String: foundTerm, Valid: true},
		})
		return err
	})

	// 2. FTS - Not Found
	measure("FTS Not Found", func() error {
		_, err := q.ListLogsWithFilters(ctx, db.ListLogsWithFiltersParams{
			Limit:         100,
			Offset:        0,
			ContentSearch: pgtype.Text{String: notFoundTerm, Valid: true},
		})
		return err
	})

	// 3. FTS - Short Input
	measure("FTS Short Input", func() error {
		_, err := q.ListLogsWithFilters(ctx, db.ListLogsWithFiltersParams{
			Limit:         100,
			Offset:        0,
			ContentSearch: pgtype.Text{String: shortTerm, Valid: true},
		})
		return err
	})

	// 4. FTS - No Limit (Large Limit)
	measure("FTS No Limit", func() error {
		_, err := q.ListLogsWithFilters(ctx, db.ListLogsWithFiltersParams{
			Limit:         int32(cfg.DatasetSize),
			Offset:        0,
			ContentSearch: pgtype.Text{String: foundTerm, Valid: true},
		})
		return err
	})

	// 5. Partial - Found
	measure("Partial Found", func() error {
		_, err := q.SearchLogsPartial(ctx, db.SearchLogsPartialParams{
			Limit:      pgtype.Int4{Int32: 100, Valid: true},
			Offset:     pgtype.Int4{Int32: 0, Valid: true},
			SearchTerm: pgtype.Text{String: foundTerm, Valid: true},
		})
		return err
	})

	// 6. Partial - Not Found
	measure("Partial Not Found", func() error {
		_, err := q.SearchLogsPartial(ctx, db.SearchLogsPartialParams{
			Limit:      pgtype.Int4{Int32: 100, Valid: true},
			Offset:     pgtype.Int4{Int32: 0, Valid: true},
			SearchTerm: pgtype.Text{String: notFoundTerm, Valid: true},
		})
		return err
	})

	// 7. Partial - Short Input
	measure("Partial Short Input", func() error {
		_, err := q.SearchLogsPartial(ctx, db.SearchLogsPartialParams{
			Limit:      pgtype.Int4{Int32: 100, Valid: true},
			Offset:     pgtype.Int4{Int32: 0, Valid: true},
			SearchTerm: pgtype.Text{String: shortTerm, Valid: true},
		})
		return err
	})

	// 8. Partial - No Limit
	measure("Partial No Limit", func() error {
		_, err := q.SearchLogsPartial(ctx, db.SearchLogsPartialParams{
			Limit:      pgtype.Int4{Int32: int32(cfg.DatasetSize), Valid: true},
			Offset:     pgtype.Int4{Int32: 0, Valid: true},
			SearchTerm: pgtype.Text{String: foundTerm, Valid: true},
		})
		return err
	})

	return results
}

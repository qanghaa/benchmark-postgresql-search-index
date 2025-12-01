package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"log-project/internal/db"
	"log-project/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type BenchmarkCase struct {
	Name       string
	SearchType string // "FTS" or "Partial"
	Term       string
	Limit      int32 // 0 means "No Limit" (effectively dataset size)
	Desc       string
}

type Result struct {
	Case      BenchmarkCase
	Duration  time.Duration
	RowsFound int
	Error     error
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

	// 1. Get Dataset Stats
	count, err := queries.CountLogs(ctx)
	if err != nil {
		log.Fatalf("Failed to count logs: %v", err)
	}
	log.Printf("Dataset Size: %d", count)

	// 2. Discover Terms (Common vs Rare)
	log.Println("Analyzing data to find Common and Rare terms...")
	commonTerm, rareTerm, err := discoverTerms(ctx, queries)
	if err != nil {
		log.Printf("Warning: Failed to discover terms, using defaults: %v", err)
		commonTerm = "login"
		rareTerm = "error"
	}
	log.Printf("Terms Discovered:\n - Common (Many matches): '%s'\n - Rare (Few matches): '%s'", commonTerm, rareTerm)

	// 3. Define Test Cases
	notFoundTerm := uuid.New().String()
	shortTerm := "lo"
	if len(commonTerm) >= 2 {
		shortTerm = commonTerm[:2]
	}

	cases := []BenchmarkCase{
		// --- FTS Cases ---
		{Name: "FTS Not Found", SearchType: "FTS", Term: notFoundTerm, Limit: 100, Desc: "Random UUID"},
		{Name: "FTS Rare (Few)", SearchType: "FTS", Term: rareTerm, Limit: 100, Desc: "Rare term"},
		{Name: "FTS Common (Many) Limit", SearchType: "FTS", Term: commonTerm, Limit: 100, Desc: "Common term, Limit 100"},
		{Name: "FTS Common (Many) NoLimit", SearchType: "FTS", Term: commonTerm, Limit: int32(count), Desc: "Common term, Full Scan"},
		{Name: "FTS Short Input", SearchType: "FTS", Term: shortTerm, Limit: 100, Desc: "1-2 chars"},

		// --- Partial Cases ---
		{Name: "Partial Not Found", SearchType: "Partial", Term: notFoundTerm, Limit: 100, Desc: "Random UUID"},
		{Name: "Partial Rare (Few)", SearchType: "Partial", Term: rareTerm, Limit: 100, Desc: "Rare term"},
		{Name: "Partial Common (Many) Limit", SearchType: "Partial", Term: commonTerm, Limit: 100, Desc: "Common term, Limit 100"},
		{Name: "Partial Common (Many) NoLimit", SearchType: "Partial", Term: commonTerm, Limit: int32(count), Desc: "Common term, Full Scan"},
		{Name: "Partial Short Input", SearchType: "Partial", Term: shortTerm, Limit: 100, Desc: "1-2 chars"},
	}

	// 4. Warm Up
	log.Println("Warming up...")
	warmUp(ctx, queries, commonTerm)

	// 5. Run Benchmark
	log.Println("Running benchmark...")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Type\tCase\tLimit\tDuration\tRows\tDescription")

	for _, c := range cases {
		res := runCase(ctx, queries, c)
		if res.Error != nil {
			log.Printf("Error in %s: %v", c.Name, res.Error)
			continue
		}
		limitStr := fmt.Sprintf("%d", c.Limit)
		if c.Limit == int32(count) {
			limitStr = "ALL"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%d\t%s\n", c.SearchType, c.Name, limitStr, res.Duration, res.RowsFound, c.Desc)
	}
	w.Flush()
}

func runCase(ctx context.Context, q *db.Queries, c BenchmarkCase) Result {
	start := time.Now()
	var count int
	var err error

	if c.SearchType == "FTS" {
		var logs []db.Log
		logs, err = q.ListLogsWithFilters(ctx, db.ListLogsWithFiltersParams{
			Limit:         c.Limit,
			Offset:        0,
			ContentSearch: pgtype.Text{String: c.Term, Valid: true},
		})
		count = len(logs)
	} else {
		var logs []db.Log
		logs, err = q.SearchLogsPartial(ctx, db.SearchLogsPartialParams{
			Limit:      pgtype.Int4{Int32: c.Limit, Valid: true},
			Offset:     pgtype.Int4{Int32: 0, Valid: true},
			SearchTerm: pgtype.Text{String: c.Term, Valid: true},
		})
		count = len(logs)
	}

	return Result{
		Case:      c,
		Duration:  time.Since(start),
		RowsFound: count,
		Error:     err,
	}
}

func discoverTerms(ctx context.Context, q *db.Queries) (string, string, error) {
	// Fetch sample logs
	logs, err := q.ListLogs(ctx, db.ListLogsParams{Limit: 1000, Offset: 0})
	if err != nil {
		return "", "", err
	}

	wordCounts := make(map[string]int)
	for _, l := range logs {
		var content models.Content
		if err := json.Unmarshal(l.Content, &content); err != nil {
			continue
		}

		// Extract text from values
		for _, v := range content {
			if str, ok := v.(string); ok {
				// Simple tokenization: split by space, lowercase, trim
				words := strings.Fields(str)
				for _, w := range words {
					w = strings.ToLower(strings.Trim(w, ".,!?-()[]{}\""))
					if len(w) > 3 { // Ignore short words
						wordCounts[w]++
					}
				}
			}
		}
	}

	if len(wordCounts) == 0 {
		return "login", "error", nil
	}

	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range wordCounts {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	common := sorted[0].Key
	rare := sorted[len(sorted)-1].Key

	// Try to find a rare term that appears at least once but not too many times
	for i := len(sorted) - 1; i >= 0; i-- {
		if sorted[i].Value >= 1 && sorted[i].Value < 5 {
			rare = sorted[i].Key
			break
		}
	}

	return common, rare, nil
}

func warmUp(ctx context.Context, q *db.Queries, term string) {
	q.ListLogsWithFilters(ctx, db.ListLogsWithFiltersParams{
		Limit:         10,
		Offset:        0,
		ContentSearch: pgtype.Text{String: term, Valid: true},
	})
}

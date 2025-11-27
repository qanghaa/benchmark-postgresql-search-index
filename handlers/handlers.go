package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"log-project/internal/db"
	"log-project/models"
	"log-project/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func New(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:    pool,
		queries: db.New(pool),
	}
}

// InitializeData godoc
// @Summary Initialize database with sample data
// @Description Generate and insert sample logs into the database using COPY FROM for optimal performance
// @Tags initialization
// @Accept json
// @Produce json
// @Param request body models.InitializeRequest true "Initialization parameters"
// @Success 200 {object} map[string]interface{}
// @Router /initialize [post]
func (h *Handler) InitializeData(c *gin.Context) {
	var req models.InitializeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start timing
	start := time.Now()

	// Generate all data first
	log.Printf("Generating %d records with %s content size...\n", req.RecordCount, req.ContentSize)

	batchSize := 1000
	totalBatches := (req.RecordCount + batchSize - 1) / batchSize

	ctx := context.Background()
	totalInserted := int64(0)

	for batch := 0; batch < totalBatches; batch++ {
		currentBatchSize := batchSize
		if batch == totalBatches-1 {
			currentBatchSize = req.RecordCount - (batch * batchSize)
		}
		userID := uuid.New()
		domain := getRandomDomain()
		params := make([]db.BulkInsertLogsParams, currentBatchSize)

		for i := 0; i < currentBatchSize; i++ {
			action := getRandomAction()
			content := utils.GenerateSampleContent(req.ContentSize)
			createdAt := time.Now().Add(-time.Duration(rand.Intn(86400*30)) * time.Second)

			// Convert content to JSON bytes
			contentBytes, err := json.Marshal(content)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal content"})
				return
			}

			params[i] = db.BulkInsertLogsParams{
				UserID:  pgtype.UUID{Bytes: userID, Valid: true},
				Domain:  domain,
				Action:  action,
				Content: contentBytes,
				CreatedAt: pgtype.Timestamptz{
					Time:  createdAt,
					Valid: true,
				},
			}
		}

		// Use CopyFrom for bulk insert
		rowsInserted, err := h.queries.BulkInsertLogs(ctx, params)
		if err != nil {
			log.Printf("Failed to insert batch %d: %v\n", batch, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to insert batch: %v", err)})
			return
		}

		totalInserted += rowsInserted

		// Progress feedback
		progress := float64(batch+1) / float64(totalBatches) * 100
		log.Printf("Progress: %.2f%% (Inserted %d rows in batch %d)\n", progress, rowsInserted, batch+1)
	}

	duration := time.Since(start)
	recordsPerSecond := float64(totalInserted) / duration.Seconds()

	log.Printf("Completed! Inserted %d records in %s (%.2f records/sec)\n", totalInserted, duration, recordsPerSecond)

	c.JSON(http.StatusOK, gin.H{
		"message":            "Data initialized successfully",
		"record_count":       totalInserted,
		"content_size":       req.ContentSize,
		"duration":           duration.String(),
		"records_per_second": fmt.Sprintf("%.2f", recordsPerSecond),
	})
}

// GetLogs godoc
// @Summary Get logs with filtering
// @Description Retrieve logs with optional filtering parameters
// @Tags logs
// @Accept json
// @Produce json
// @Param user_id query string false "User ID filter"
// @Param domain query string false "Domain filter"
// @Param created_at query string false "Created date from filter (YYYY-MM-DD)"
// @Param created_at_to query string false "Created date to filter (YYYY-MM-DD)"
// @Param content_like query string false "Content search filter"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} map[string]interface{}
// @Router /logs [get]
func (h *Handler) GetLogs(c *gin.Context) {
	var filter models.LogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	// Start timing for query performance
	queryStart := time.Now()

	// Build filter parameters
	var userID pgtype.UUID
	var domain pgtype.Text
	var createdAtFrom pgtype.Timestamptz
	var createdAtTo pgtype.Timestamptz
	var contentSearch pgtype.Text

	if filter.UserID != nil && *filter.UserID != "" {
		parsedUUID, err := uuid.Parse(*filter.UserID)
		if err == nil {
			userID = pgtype.UUID{Bytes: parsedUUID, Valid: true}
		}
	}

	if filter.Domain != nil && *filter.Domain != "" {
		domain = pgtype.Text{String: *filter.Domain, Valid: true}
	}

	if filter.CreatedAt != nil && *filter.CreatedAt != "" {
		t, err := time.Parse("2006-01-02", *filter.CreatedAt)
		if err == nil {
			createdAtFrom = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	if filter.CreatedAtTo != nil && *filter.CreatedAtTo != "" {
		t, err := time.Parse("2006-01-02", *filter.CreatedAtTo)
		if err == nil {
			// Set to end of day
			t = t.Add(24*time.Hour - time.Second)
			createdAtTo = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	if filter.ContentLike != nil && *filter.ContentLike != "" {
		contentSearch = pgtype.Text{String: *filter.ContentLike, Valid: true}
	}

	// Count total records
	total, err := h.queries.CountLogsWithFilters(ctx, db.CountLogsWithFiltersParams{
		UserID:        userID,
		Domain:        domain,
		CreatedAtFrom: createdAtFrom,
		CreatedAtTo:   createdAtTo,
		ContentSearch: contentSearch,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count records"})
		return
	}

	// Calculate pagination
	offset := int32((filter.Page - 1) * filter.Limit)
	limit := int32(filter.Limit)

	// Fetch logs
	logs, err := h.queries.ListLogsWithFilters(ctx, db.ListLogsWithFiltersParams{
		UserID:        userID,
		Domain:        domain,
		CreatedAtFrom: createdAtFrom,
		CreatedAtTo:   createdAtTo,
		ContentSearch: contentSearch,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query logs"})
		return
	}

	queryDuration := time.Since(queryStart)

	// Convert to response format
	response := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		var content map[string]interface{}
		if err := json.Unmarshal(log.Content, &content); err != nil {
			content = map[string]interface{}{"raw": string(log.Content)}
		}

		response[i] = map[string]interface{}{
			"id":         uuidToString(log.ID),
			"user_id":    uuidToString(log.UserID),
			"domain":     log.Domain,
			"action":     log.Action,
			"content":    content,
			"created_at": log.CreatedAt.Time,
		}
	}

	totalPages := int((total + int64(filter.Limit) - 1) / int64(filter.Limit))

	c.JSON(http.StatusOK, gin.H{
		"data":           response,
		"total":          total,
		"page":           filter.Page,
		"limit":          filter.Limit,
		"total_pages":    totalPages,
		"query_duration": queryDuration.String(),
	})
}

// TruncateDatabase godoc
// @Summary Truncate all logs
// @Description Remove all log records from the database
// @Tags database
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /truncate [delete]
func (h *Handler) TruncateDatabase(c *gin.Context) {
	ctx := context.Background()

	err := h.queries.TruncateLogs(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to truncate database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Database truncated successfully",
	})
}

// SearchLogsPartial godoc
// @Summary Search logs using partial match (ILIKE)
// @Description Search logs using efficient partial matching with pg_trgm index
// @Tags logs
// @Accept json
// @Produce json
// @Param user_id query string false "User ID filter"
// @Param domain query string false "Domain filter"
// @Param created_at query string false "Created date from filter (YYYY-MM-DD)"
// @Param created_at_to query string false "Created date to filter (YYYY-MM-DD)"
// @Param search_term query string true "Partial search term"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} map[string]interface{}
// @Router /search/partial [get]
func (h *Handler) SearchLogsPartial(c *gin.Context) {
	var filter models.LogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if filter.SearchTerm == nil || *filter.SearchTerm == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search_term is required"})
		return
	}

	ctx := context.Background()

	// Start timing for query performance
	queryStart := time.Now()

	// Build filter parameters
	var userID pgtype.UUID
	var domain pgtype.Text
	var createdAtFrom pgtype.Timestamptz
	var createdAtTo pgtype.Timestamptz
	var searchTerm pgtype.Text

	if filter.UserID != nil && *filter.UserID != "" {
		parsedUUID, err := uuid.Parse(*filter.UserID)
		if err == nil {
			userID = pgtype.UUID{Bytes: parsedUUID, Valid: true}
		}
	}

	if filter.Domain != nil && *filter.Domain != "" {
		domain = pgtype.Text{String: *filter.Domain, Valid: true}
	}

	if filter.CreatedAt != nil && *filter.CreatedAt != "" {
		t, err := time.Parse("2006-01-02", *filter.CreatedAt)
		if err == nil {
			createdAtFrom = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	if filter.CreatedAtTo != nil && *filter.CreatedAtTo != "" {
		t, err := time.Parse("2006-01-02", *filter.CreatedAtTo)
		if err == nil {
			// Set to end of day
			t = t.Add(24*time.Hour - time.Second)
			createdAtTo = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	searchTerm = pgtype.Text{String: *filter.SearchTerm, Valid: true}

	// Count total records
	total, err := h.queries.CountLogsPartial(ctx, db.CountLogsPartialParams{
		UserID:        userID,
		Domain:        domain,
		CreatedAtFrom: createdAtFrom,
		CreatedAtTo:   createdAtTo,
		SearchTerm:    searchTerm,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count records"})
		return
	}

	// Calculate pagination
	offset := int32((filter.Page - 1) * filter.Limit)
	limit := int32(filter.Limit)

	// Fetch logs
	logs, err := h.queries.SearchLogsPartial(ctx, db.SearchLogsPartialParams{
		UserID:        userID,
		Domain:        domain,
		CreatedAtFrom: createdAtFrom,
		CreatedAtTo:   createdAtTo,
		SearchTerm:    searchTerm,
		Limit:         pgtype.Int4{Int32: limit, Valid: true},
		Offset:        pgtype.Int4{Int32: offset, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query logs"})
		return
	}

	queryDuration := time.Since(queryStart)

	// Convert to response format
	response := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		var content map[string]interface{}
		if err := json.Unmarshal(log.Content, &content); err != nil {
			content = map[string]interface{}{"raw": string(log.Content)}
		}

		response[i] = map[string]interface{}{
			"id":         uuidToString(log.ID),
			"user_id":    uuidToString(log.UserID),
			"domain":     log.Domain,
			"action":     log.Action,
			"content":    content,
			"created_at": log.CreatedAt.Time,
		}
	}

	totalPages := int((total + int64(filter.Limit) - 1) / int64(filter.Limit))

	c.JSON(http.StatusOK, gin.H{
		"data":           response,
		"total":          total,
		"page":           filter.Page,
		"limit":          filter.Limit,
		"total_pages":    totalPages,
		"query_duration": queryDuration.String(),
	})
}

// Helper functions
func getRandomDomain() string {
	domains := []string{
		"example.com", "test.org", "demo.net", "app.io", "api.service.com",
		"web.portal.com", "mobile.app.net", "admin.system.org", "user.platform.io",
		"data.analytics.com", "payments.service.net", "content.media.org", "social.platform.io",
	}
	return domains[rand.Intn(len(domains))]
}

func getRandomAction() string {
	actions := []string{
		"user_login", "user_logout", "page_view", "button_click", "form_submit",
		"file_upload", "file_download", "search_query", "filter_apply", "sort_change",
		"create_record", "update_record", "delete_record", "export_data", "import_data",
		"send_message", "receive_message", "share_content", "like_post", "comment_post",
		"subscribe", "unsubscribe", "follow_user", "unfollow_user", "report_issue",
		"request_feature", "update_settings", "change_password", "reset_password",
	}
	return actions[rand.Intn(len(actions))]
}

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}

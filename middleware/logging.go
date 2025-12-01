package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"time"

	"log-project/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLogger(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}
		// Restore the io.ReadCloser to its original state
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// Collect params
			uriParams := make(map[string]string)
			for _, p := range c.Params {
				uriParams[p.Key] = p.Value
			}

			queryParams := c.Request.URL.Query()

			var bodyJSON interface{}
			if len(bodyBytes) > 0 {
				_ = json.Unmarshal(bodyBytes, &bodyJSON)
			}

			logData := map[string]interface{}{
				"uri_params":   uriParams,
				"query_params": queryParams,
				"body":         bodyJSON,
				"method":       c.Request.Method,
				"path":         c.Request.URL.Path,
				"status":       c.Writer.Status(),
			}

			contentBytes, err := json.Marshal(logData)
			if err != nil {
				log.Printf("Failed to marshal log data: %v", err)
				return
			}

			queries := db.New(pool)

			// Generate a random UUID for user_id since we don't have auth yet
			// In a real app, this would come from the context
			userID := uuid.New()
			var pgUserID pgtype.UUID
			pgUserID.Bytes = userID
			pgUserID.Valid = true

			// Use current time
			var pgCreatedAt pgtype.Timestamptz
			pgCreatedAt.Time = time.Now()
			pgCreatedAt.Valid = true

			_, err = queries.CreateLog(c.Request.Context(), db.CreateLogParams{
				UserID:    pgUserID,
				Domain:    "example-api",
				Action:    c.Request.Method + " " + c.Request.URL.Path,
				Content:   contentBytes,
				CreatedAt: pgCreatedAt,
			})
			if err != nil {
				log.Printf("Failed to create log: %v", err)
			}
		}
	}
}

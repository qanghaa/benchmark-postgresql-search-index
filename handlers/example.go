package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ExampleRequest struct {
	ID   string `uri:"id" binding:"required"`
	Type string `form:"type"`
	Name string `json:"name"`
}

func (h *Handler) ExampleAPI(c *gin.Context) {
	var req ExampleRequest

	// Bind URI
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Bind Query
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    req,
	})
}

package helper

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationParams holds pagination request parameters.
type PaginationParams struct {
	Page  int
	Limit int
}

// PageResponse wraps paginated data with metadata.
type PageResponse struct {
	Data      interface{} `json:"data"`
	Page      int         `json:"page"`
	Limit     int         `json:"limit"`
	Total     int         `json:"total"`
	TotalPage int         `json:"total_page"`
	HasMore   bool        `json:"has_more"`
}

// ParsePaginationParams extracts pagination params from query string.
// Default: page=1, limit=20. Max limit=100.
func ParsePaginationParams(c *gin.Context) PaginationParams {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	// Validation
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Max limit to prevent DOS
	}

	return PaginationParams{
		Page:  page,
		Limit: limit,
	}
}

// CalculateOffset converts page number to SQL OFFSET.
// page=1, limit=20 → offset=0
// page=2, limit=20 → offset=20
func (p PaginationParams) CalculateOffset() int {
	return (p.Page - 1) * p.Limit
}

// NewPageResponse creates a paginated response wrapper.
// totalCount: total records in database matching query
func (p PaginationParams) NewPageResponse(data interface{}, totalCount int) PageResponse {
	totalPage := (totalCount + p.Limit - 1) / p.Limit // Ceiling division
	if totalPage <= 0 {
		totalPage = 1
	}

	return PageResponse{
		Data:      data,
		Page:      p.Page,
		Limit:     p.Limit,
		Total:     totalCount,
		TotalPage: totalPage,
		HasMore:   p.Page < totalPage,
	}
}

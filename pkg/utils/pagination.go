package utils

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type PaginationParams struct {
	Page   int
	Limit  int
	Search string
	Sort   string
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

func PaginationFromContext(c *fiber.Ctx) PaginationParams {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: c.Query("search", ""),
		Sort:   c.Query("sort", ""),
	}
}

func (p PaginationParams) Apply(db *gorm.DB) *gorm.DB {
	offset := (p.Page - 1) * p.Limit
	return db.Offset(offset).Limit(p.Limit)
}

func BuildPaginationMeta(p PaginationParams, total int64) PaginationMeta {
	totalPages := int((total + int64(p.Limit) - 1) / int64(p.Limit))
	if totalPages < 1 {
		totalPages = 1
	}
	return PaginationMeta{Page: p.Page, Limit: p.Limit, TotalItems: total, TotalPages: totalPages}
}

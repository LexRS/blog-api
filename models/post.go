package models

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PaginatedPosts struct {
	Posts []Post `json:"posts"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore bool `json:"has_more"`
}

type Cursor struct {
	ID int `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (c *Cursor) Encode() string {
	data := []byte(fmt.Sprintf("%d|%d", c.ID, c.CreatedAt.UnixNano()))
	return base64.RawURLEncoding.EncodeToString(data)
}

func DecodeCursor(cursorStr string) (*Cursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(cursorStr)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(string(data), "|")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid cursor format")
	}

	id, err := strconv.Atoi(parts[0])
    if err != nil {
        return nil, err
    }
    
    nano, err := strconv.ParseInt(parts[1], 10, 64)
    if err != nil {
        return nil, err
    }
    
    return &Cursor{
        ID:        id,
        CreatedAt: time.Unix(0, nano),
    }, nil
}

// Query parameters
type PostQuery struct {
    Cursor   string `json:"cursor" query:"cursor"`     // Encoded cursor
    Limit    int    `json:"limit" query:"limit"`       // Page size (max 100)
    SortBy   string `json:"sort_by" query:"sort_by"`   // created_at, updated_at, title
    SortDir  string `json:"sort_dir" query:"sort_dir"` // asc, desc
    Author   string `json:"author" query:"author"`     // Filter by author
    Search   string `json:"search" query:"search"`     // Search term
}

// Default query values
func DefaultPostQuery() PostQuery {
    return PostQuery{
        Limit:    20,
        SortBy:   "created_at",
        SortDir:  "desc",
    }
}

// Validation
func (q *PostQuery) Validate() error {
    if q.Limit < 1 || q.Limit > 100 {
        return fmt.Errorf("limit must be between 1 and 100")
    }
    
    allowedSortBy := map[string]bool{"created_at": true, "updated_at": true, "title": true}
    if !allowedSortBy[q.SortBy] {
        return fmt.Errorf("sort_by must be one of: created_at, updated_at, title")
    }
    
    if q.SortDir != "asc" && q.SortDir != "desc" {
        return fmt.Errorf("sort_dir must be 'asc' or 'desc'")
    }
    
    return nil
}

type CreatePostRequest struct {
	Title   string `json:"title" validate:"required"`
	Content string `json:"content" validate:"required"`
	Author  string `json:"author" validate:"required"`
}

type UpdatePostRequest struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Author  string `json:"author,omitempty"`
}

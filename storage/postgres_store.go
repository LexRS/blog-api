package storage

import (
	"blog-api/models"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func (s *PostgresStore) getPaginatedPosts(query models.PostQuery) (*models.PaginatedPosts, error) {
	// Validate query
	if err := query.Validate(); err != nil {
		return nil, err
	}

	whereClause, args := s.buildWhereClause(query)

	orderClause := s.buildOrderClause(query)

	// Get total count (optional, can be expensive)
	// total, err := s.getTotalPosts(whereClause, args)
	// if err != nil {
	//     log.Printf("Warning: Could not get total count: %v", err)
	//     total = -1 // Indicate unknown
	// }

	// Build main query with cursor
	mainQuery, args := s.buildPaginatedQuery(query, whereClause, orderClause, args)

	// Execute query
	rows, err := s.db.Query(mainQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	// Parse results
	posts, nextCursor, err := s.scanPaginatedPosts(rows, query)
	if err != nil {
		return nil, err
	}

	// Determine if there are more results
	hasMore := false
	if len(posts) > 0 && len(posts) == query.Limit {
		hasMore = true
	}

	// Generate previous cursor (if we have a cursor)
	prevCursor := ""
	if query.Cursor != "" {
		// For simplicity, we'll use the first post's cursor as prev
		// In production, you might want to implement proper backward pagination
		if len(posts) > 0 {
			firstCursor := &models.Cursor{
				ID:        posts[0].ID,
				CreatedAt: posts[0].CreatedAt,
			}
			prevCursor = firstCursor.Encode()
		}
	}

	return &models.PaginatedPosts{
		Posts:      posts,
		PrevCursor: prevCursor,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		//Total: total
	}, nil
}

func (s *PostgresStore) buildWhereClause(query models.PostQuery) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Handle cursor
	if query.Cursor != "" {
		cursor, err := models.DecodeCursor(query.Cursor)
		if err == nil {
			operator := "<"
			if query.SortDir == "asc" {
				operator = ">"
			}

			if query.SortBy == "created_at" {
				conditions = append(conditions,
					fmt.Sprintf("(created_at, id) %s ($%d, $%d)",
						operator, argIndex, argIndex+1))
				args = append(args, cursor.CreatedAt, cursor.ID)
				argIndex += 2
			} else if query.SortBy == "id" {
				conditions = append(conditions,
					fmt.Sprintf("id %s $%d", operator, argIndex))
				args = append(args, cursor.ID)
				argIndex++
			}
		}
	}

	// Filter by author
	if query.Author != "" {
		conditions = append(conditions, fmt.Sprintf("author = $%d", argIndex))
		args = append(args, query.Author)
		argIndex++
	}
	// Search
	if query.Search != "" {
		conditions = append(conditions,
			fmt.Sprintf("(title ILIKE $%d OR content ILIKE $%d)",
				argIndex, argIndex+1))
		args = append(args, "%"+query.Search+"%", "%"+query.Search+"%")
		argIndex += 2
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (s *PostgresStore) buildOrderClause(query models.PostQuery) string {
	if query.SortBy == "created_at" {
		return fmt.Sprintf("ORDER BY created_at %s, id %s", strings.ToUpper(query.SortDir), strings.ToUpper(query.SortDir))
	}

	return fmt.Sprintf("ORDER BY %s %s", query.SortBy, strings.ToUpper(query.SortDir))
}

// func (s *PostgresStore) getTotalPosts(whereClause string, args []interface{}) (int, error) {
//     query := fmt.Sprintf("SELECT COUNT(*) FROM posts %s", whereClause)

//     var total int
//     err := s.db.QueryRow(query, args...).Scan(&total)
//     if err != nil {
//         return 0, err
//     }
//     return total, nil
// }

func (s *PostgresStore) buildPaginatedQuery(
	query models.PostQuery,
	whereClause, orderClause string,
	args []interface{},
) (string, []interface{}) {
	// Add limit
	limitClause := fmt.Sprintf("LIMIT $%d", len(args)+1)
	args = append(args, query.Limit+1) // Fetch one extra to check if there's more

	// Build final query
	sql := fmt.Sprintf(`
        SELECT id, title, content, author, created_at, updated_at 
        FROM posts 
        %s 
        %s 
        %s
    `, whereClause, orderClause, limitClause)

	return sql, args
}

func (s *PostgresStore) scanPaginatedPosts(
	rows *sql.Rows,
	query models.PostQuery,
) ([]models.Post, string, error) {
	var posts []models.Post
	var nextCursor string

	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.Author,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, "", err
		}
		posts = append(posts, post)
	}

	// Check if we have extra row for "has more"
	if len(posts) > query.Limit {
		posts = posts[:query.Limit] // Remove the extra
		lastPost := posts[len(posts)-1]

		// Create cursor for next page
		cursor := &models.Cursor{
			ID:        lastPost.ID,
			CreatedAt: lastPost.CreatedAt,
		}
		nextCursor = cursor.Encode()
	}

	return posts, nextCursor, nil
}

func NewPostgresStore(connectionString string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to PostgreSQL database")
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Init() error {
	return s.createPostsTable()
}

func (s *PostgresStore) createPostsTable() error {
	query := `
    CREATE TABLE IF NOT EXISTS posts (
        id SERIAL PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        content TEXT NOT NULL,
        author VARCHAR(100) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);
    CREATE INDEX IF NOT EXISTS idx_posts_author ON posts(author);
    `

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) GetAll() ([]models.Post, error) {
	query := `
    SELECT id, title, content, author, created_at, updated_at 
    FROM posts 
    ORDER BY created_at DESC
    `

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.Author,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (s *PostgresStore) GetByID(id int) (*models.Post, error) {
	query := `
    SELECT id, title, content, author, created_at, updated_at 
    FROM posts 
    WHERE id = $1
    `

	row := s.db.QueryRow(query, id)

	var post models.Post
	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Author,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *PostgresStore) Create(post models.Post) (*models.Post, error) {
	query := `
    INSERT INTO posts (title, content, author) 
    VALUES ($1, $2, $3) 
    RETURNING id, created_at, updated_at
    `

	var id int
	var createdAt, updatedAt time.Time

	err := s.db.QueryRow(
		query,
		post.Title,
		post.Content,
		post.Author,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	post.ID = id
	post.CreatedAt = createdAt
	post.UpdatedAt = updatedAt

	return &post, nil
}

func (s *PostgresStore) Update(id int, updated models.Post) (*models.Post, error) {
	query := `
    UPDATE posts 
    SET 
        title = COALESCE(NULLIF($1, ''), title),
        content = COALESCE(NULLIF($2, ''), content),
        author = COALESCE(NULLIF($3, ''), author),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = $4
    RETURNING title, content, author, created_at, updated_at
    `

	row := s.db.QueryRow(
		query,
		updated.Title,
		updated.Content,
		updated.Author,
		id,
	)

	var post models.Post
	post.ID = id

	err := row.Scan(
		&post.Title,
		&post.Content,
		&post.Author,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *PostgresStore) Delete(id int) error {
	query := `DELETE FROM posts WHERE id = $1`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return nil // Or return an error if you want to indicate not found
	}

	return nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

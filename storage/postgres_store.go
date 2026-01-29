package storage

import (
	"blog-api/models"
	"database/sql"
	"log"
	"time"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
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

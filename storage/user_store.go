package storage

import (
    "database/sql"
    "time"
    "blog-api/models"
    "golang.org/x/crypto/bcrypt"
)

type UserStore interface {
    Create(user models.User) (*models.User, error)
    GetByEmail(email string) (*models.User, error)
    GetByID(id int) (*models.User, error)
}

type PostgresUserStore struct {
    db *sql.DB
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
    return &PostgresUserStore{db: db}
}

func (s *PostgresUserStore) Create(user models.User) (*models.User, error) {
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    
    query := `
        INSERT INTO users (username, email, password, role, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, updated_at
    `
    
    now := time.Now()
    var id int
    var createdAt, updatedAt time.Time
    
    err = s.db.QueryRow(query, user.Username, user.Email, string(hashedPassword), 
        user.Role, now, now).Scan(&id, &createdAt, &updatedAt)
    
    if err != nil {
        return nil, err
    }
    
    user.ID = id
    user.CreatedAt = createdAt
    user.UpdatedAt = updatedAt
    user.Password = "" // Clear password
    
    return &user, nil
}

func (s *PostgresUserStore) GetByEmail(email string) (*models.User, error) {
    query := `SELECT id, username, email, password, role, created_at, updated_at 
              FROM users WHERE email = $1`
    
    var user models.User
    err := s.db.QueryRow(query, email).Scan(
        &user.ID, &user.Username, &user.Email, &user.Password,
        &user.Role, &user.CreatedAt, &user.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

func (s *PostgresUserStore) GetByID(id int) (*models.User, error) {
    query := `SELECT id, username, email, role, created_at, updated_at 
              FROM users WHERE id = $1`
    
    var user models.User
    err := s.db.QueryRow(query, id).Scan(
        &user.ID, &user.Username, &user.Email,
        &user.Role, &user.CreatedAt, &user.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

func (s *PostgresUserStore) VerifyPassword(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}
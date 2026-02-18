package storage

import "blog-api/models"

type PostStore interface {
	GetPostsPaginated(query models.PostQuery) (*models.PaginatedPosts, error)
	GetAll() ([]models.Post, error)
	GetByID(id int) (*models.Post, error)
	Create(post models.Post) (*models.Post, error)
	Update(id int, post models.Post) (*models.Post, error)
	Delete(id int) error
	Close() error
}
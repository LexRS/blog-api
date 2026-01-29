package storage

// import (
// 	"blog-api/models"
// 	"sync"
// 	"time"
// )

// type PostStore interface {
// 	GetAll() ([]models.Post, error)
// 	GetByID(id int) (*models.Post, error)
// 	Create(post models.Post) (*models.Post, error)
// 	Update(id int, post models.Post) (*models.Post, error)
// 	Delete(id int) error
// }

// type InMemoryPostStore struct {
// 	mu    sync.RWMutex
// 	posts map[int]models.Post
// 	id    int
// }

// func NewInMemoryPostStore() *InMemoryPostStore {
// 	return &InMemoryPostStore{
// 		posts: make(map[int]models.Post),
// 		id:    1,
// 	}
// }

// func (ps *InMemoryPostStore) GetAll() ([]models.Post, error) {
// 	ps.mu.RLock()
// 	defer ps.mu.RUnlock()

// 	posts := make([]models.Post, 0, len(ps.posts))

// 	for _, post := range ps.posts {
// 		posts = append(posts, post)
// 	}
// 	return posts, nil
// }

// func (ps *InMemoryPostStore) GetByID(id int) (*models.Post, error) {
// 	ps.mu.RLock()
// 	defer ps.mu.RUnlock()

// 	post, exists := ps.posts[id]
// 	if !exists {
// 		return nil, nil
// 	}
// 	return &post, nil
// }

// func (ps *InMemoryPostStore) Create(post models.Post) (*models.Post, error) {
// 	ps.mu.Lock()
// 	defer ps.mu.Unlock()

// 	post.ID = ps.id
// 	post.CreatedAt = time.Now()
// 	post.UpdatedAt = time.Now()

// 	ps.posts[ps.id] = post
// 	ps.id++

// 	return &post, nil
// }

// func (s *InMemoryPostStore) Update(id int, updated models.Post) (*models.Post, error) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	existing, exists := s.posts[id]
// 	if !exists {
// 		return nil, nil
// 	}

// 	if updated.Title != "" {
// 		existing.Title = updated.Title
// 	}
// 	if updated.Content != "" {
// 		existing.Content = updated.Content
// 	}
// 	if updated.Author != "" {
// 		existing.Author = updated.Author
// 	}

// 	existing.UpdatedAt = time.Now()
// 	s.posts[id] = existing

// 	return &existing, nil
// }

// func (s *InMemoryPostStore) Delete(id int) error {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	if _, exists := s.posts[id]; !exists {
// 		return nil
// 	}

// 	delete(s.posts, id)
// 	return nil
// }

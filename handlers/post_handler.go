package handlers

import (
	"blog-api/models"
	"blog-api/storage"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type PostHandler struct {
	store storage.PostStore
}

func NewPostStoreHandler(store storage.PostStore) *PostHandler {
	return &PostHandler{store: store}
}

func (h *PostHandler) GetPostsPaginated(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := models.DefaultPostQuery()

	// Get cursor from query parameters
	if cursor := r.URL.Query().Get("cursor"); cursor != "" {
		query.Cursor = cursor
	}

	// Get limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			query.Limit = limit
		}
	}

	// Get sort parameters
    if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
        query.SortBy = sortBy
    }
    
    if sortDir := r.URL.Query().Get("sort_dir"); sortDir != "" {
        query.SortDir = sortDir
    }
    
    // Get filters
    if author := r.URL.Query().Get("author"); author != "" {
        query.Author = author
    }
    
    if search := r.URL.Query().Get("search"); search != "" {
        query.Search = search
    }

	// Validate query
    if err := query.Validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	// Get paginated posts
	paginatedPosts, err := h.store.(interface {
		GetPostsPaginated(query models.PostQuery) (*models.PaginatedPosts, error)
	}).GetPostsPaginated(query)

	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(paginatedPosts)
}

// Update GetAllPosts to use pagination (backward compatibility)
func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
    // Check if using new pagination query params
    if r.URL.Query().Get("cursor") != "" || 
       r.URL.Query().Get("limit") != "" ||
       r.URL.Query().Get("page") != "" {
        
        // Redirect to paginated endpoint
        h.GetPostsPaginated(w, r)
        return
    }
    
    // Old behavior: get all posts (limit to 100 for safety)
    posts, err := h.store.GetAll()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	// Limit to 100 posts for backward compatibility
    if len(posts) > 100 {
        posts = posts[:100]
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := h.store.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if post == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// ===================== Create Post ===============================================================
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePostRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Simple validation
	if req.Title == "" || req.Content == "" || req.Author == "" {
		http.Error(w, "Title, content, and author are required", http.StatusBadRequest)
		return
	}

	post := models.Post{
		Title:   req.Title,
		Content: req.Content,
		Author:  req.Author,
	}

	createdPost, err := h.store.Create(post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPost)
}

//===================== Update Post ===============================================================

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	var req models.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post := models.Post{
		Title:   req.Title,
		Content: req.Content,
		Author:  req.Author,
	}

	updatedPost, err := h.store.Update(id, post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if updatedPost == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedPost)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	err = h.store.Delete(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

package handlers

import (
    "encoding/json"
    "net/http"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "blog-api/models"
    "blog-api/storage"
)

type AuthHandler struct {
    userStore storage.UserStore
    jwtSecret []byte
}

func NewAuthHandler(userStore storage.UserStore, jwtSecret string) *AuthHandler {
    return &AuthHandler{
        userStore: userStore,
        jwtSecret: []byte(jwtSecret),
    }
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req models.RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Check if user exists
    existing, _ := h.userStore.GetByEmail(req.Email)
    if existing != nil {
        http.Error(w, "User already exists", http.StatusConflict)
        return
    }
    
    // Create user
    user := models.User{
        Username: req.Username,
        Email:    req.Email,
        Password: req.Password,
        Role:     "reader", // Default role
    }
    
    created, err := h.userStore.Create(user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Generate token
    token, err := h.generateToken(created)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(models.AuthResponse{
        Token: token,
        User:  *created,
    })
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req models.LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Get user
    user, err := h.userStore.GetByEmail(req.Email)
    if err != nil || user == nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    
    // Verify password
    pgStore, ok := h.userStore.(*storage.PostgresUserStore)
    if !ok || !pgStore.VerifyPassword(user.Password, req.Password) {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    
    // Clear password from user object
    user.Password = ""
    
    // Generate token
    token, err := h.generateToken(user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(models.AuthResponse{
        Token: token,
        User:  *user,
    })
}

func (h *AuthHandler) generateToken(user *models.User) (string, error) {
    claims := &models.Claims{
        UserID:   user.ID,
        Username: user.Username,
        Email:    user.Email,
        Role:     user.Role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(h.jwtSecret)
}
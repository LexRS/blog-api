# Blog API - Go Backend Service

A production-ready RESTful blog API built with Go, featuring JWT authentication, PostgreSQL database, and cursor-based pagination.

## 🚀 Features

- ✅ **CRUD Operations** - Create, read, update, delete blog posts
- ✅ **JWT Authentication** - Secure user registration and login
- ✅ **Role-Based Access Control** - Admin, author, and reader roles
- ✅ **Cursor-Based Pagination** - Efficient navigation through large datasets
- ✅ **Advanced Filtering** - Search by title/content, filter by author
- ✅ **PostgreSQL Integration** - Reliable data persistence
- ✅ **Graceful Shutdown** - Proper cleanup on server termination
- ✅ **Request Logging** - Comprehensive request tracking
- ✅ **Connection Pooling** - Optimized database connections

## 📦 Tech Stack

| Technology | Purpose |
|------------|---------|
| **Go 1.21+** | Core language |
| **Gorilla Mux** | HTTP routing |
| **lib/pq** | PostgreSQL driver |
| **golang-jwt/jwt** | JWT authentication |
| **bcrypt** | Password hashing |
| **PostgreSQL 15+** | Database |


## 🚦 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Docker (optional)

### Installation

```bash
# Clone the repository
git clone https://github.com/LexRS/blog-api.git
cd blog-api

# Install dependencies
go mod download

# Create .env file
cp .env.example .env

# Run database migrations
psql -U postgres -d blogdb -f database/migrations/001_create_posts_table.sql
psql -U postgres -d blogdb -f database/migrations/002_create_users_table.sql

# Start the server
go run main.go
```

## 📝 API Examples

### Create Post

```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "My First Post",
    "content": "This is the content of my first blog post.",
    "author": "john"
  }'
```

### Get Paginated Posts

```bash
# First page
curl "http://localhost:8080/api/v1/posts/paginated?limit=10"

# Next page (using cursor)
curl "http://localhost:8080/api/v1/posts/paginated?cursor=MTIzfDE2NDA5OTUyMDAw..."

# Search posts
curl "http://localhost:8080/api/v1/posts/paginated?search=golang"

# Filter by author
curl "http://localhost:8080/api/v1/posts/paginated?author=john"
```

## 📈 Performance Optimizations

Database Indexes - Optimized queries with proper indexing  
Connection Pooling - Efficient database connection management  
Cursor Pagination - Constant time pagination on large datasets  
Request Timeouts - Prevents hanging requests  
Graceful Shutdown - Clean resource cleanup  

## 🤝 Contributing

Fork the repository  
Create your feature branch (git checkout -b feature/amazing-feature)  
Commit your changes (git commit -m 'Add amazing feature')  
Push to branch (git push origin feature/amazing-feature)  
Open a Pull Request  

## 📄 License

MIT License

## 👨‍💻 Author

Your Name - [LexRS](https://github.com/LexRS)

## 🙏 Acknowledgments

Gorilla Mux for excellent routing  
PostgreSQL community  
All contributors and users  


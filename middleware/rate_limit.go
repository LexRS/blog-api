package middleware

import (
    "net"
    "net/http"
    "sync"
    
    "golang.org/x/time/rate"
)

type IPRateLimiter struct {
    ips map[string]*rate.Limiter
    mu  sync.RWMutex
    r   rate.Limit
    b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
    return &IPRateLimiter{
        ips: make(map[string]*rate.Limiter),
        r:   r,
        b:   b,
    }
}

func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()
    
    limiter := rate.NewLimiter(i.r, i.b)
    i.ips[ip] = limiter
    
    return limiter
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()
    
    limiter, exists := i.ips[ip]
    if !exists {
        return i.AddIP(ip)
    }
    
    return limiter
}

func RateLimit(next http.Handler) http.Handler {
    // 100 requests per minute per IP
    limiter := NewIPRateLimiter(rate.Limit(100/60.0), 100)
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip, _, err := net.SplitHostPort(r.RemoteAddr)
        if err != nil {
            ip = r.RemoteAddr
        }
        
        limiter := limiter.GetLimiter(ip)
        if !limiter.Allow() {
            http.Error(w, http.StatusText(http.StatusTooManyRequests), 
                http.StatusTooManyRequests)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
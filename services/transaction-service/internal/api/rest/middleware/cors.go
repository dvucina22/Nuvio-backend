package middleware

import (
	"log"
	"net/http"
)

type CORSMiddleware struct {
	allowedOrigins []string
}

func NewCORSMiddleware(allowedOrigins []string) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: allowedOrigins,
	}
}

func (m *CORSMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		log.Printf("CORS: Request from origin: %s, Method: %s, Path: %s", origin, r.Method, r.URL.Path)
		log.Printf("CORS: Allowed origins: %v", m.allowedOrigins)

		allowed := false
		for _, allowedOrigin := range m.allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			log.Printf("CORS: Headers set for origin: %s", origin)
		} else {
			log.Printf("CORS: Origin not allowed: %s", origin)
		}

		if r.Method == "OPTIONS" {
			log.Printf("CORS: Handling OPTIONS preflight request")
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

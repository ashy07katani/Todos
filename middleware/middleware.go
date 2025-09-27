package middleware

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"todos/repository"
	"todos/utilities"
)

type Bucket struct {
	MaxCapacity   float64
	CurrentTokens float64
	RateRefill    float64
	LastRefill    time.Time
	Mu            sync.Mutex
}

func NewBucket() *Bucket {
	b := new(Bucket)
	b.MaxCapacity = 10.0
	b.CurrentTokens = b.MaxCapacity
	b.RateRefill = 2 / 5.0
	b.LastRefill = time.Now()
	b.Mu = sync.Mutex{}
	return b
}

type RateLimiter struct {
	MuRate  sync.RWMutex
	Buckets map[string]*Bucket
}

func (rl *RateLimiter) RateLimiterMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		IP := extractIP(r)
		rl.MuRate.RLock()
		val, ok := rl.Buckets[IP]
		rl.MuRate.RUnlock()

		if !ok {
			rl.MuRate.Lock()
			rl.Buckets = make(map[string]*Bucket)
			rl.Buckets[IP] = NewBucket()
			val = rl.Buckets[IP]
			rl.MuRate.Unlock()
		}
		val.Mu.Lock()
		elapsed := time.Since(val.LastRefill)
		if elapsed > 0 {
			tokenRefill := elapsed.Seconds() * val.RateRefill
			val.CurrentTokens = min(val.CurrentTokens+tokenRefill, val.MaxCapacity)
			val.LastRefill = time.Now()
		}
		val.Mu.Unlock()
		if val.CurrentTokens < 1.0 {
			utilities.WriteError("request denied, too many requests", w, http.StatusTooManyRequests)
			w.Header().Set("Content-Type", "application/json")
			return
		}

		val.CurrentTokens -= 1.0

		rl.MuRate.Lock()
		rl.Buckets[IP] = val
		rl.MuRate.Unlock()

		next.ServeHTTP(w, r)
	})
}

func AuthMiddleWare(secret string, db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			header := r.Header.Get("Authorization")
			arr := strings.Split(header, " ")
			if len(arr) != 2 || arr[0] != "Bearer" {
				utilities.WriteError("invalid Authorization header in the request", w, http.StatusUnauthorized)
				return
			}
			claim, err := utilities.GetClaimFromJWT(arr[1], secret)
			if err != nil {
				utilities.WriteError("error fetching claim from token", w, http.StatusUnauthorized)
				return
			}
			if time.Now().After(claim.Expires_At) {
				utilities.WriteError("token is already expired", w, http.StatusUnauthorized)
				return
			}
			userName := claim.UserName
			user, err := repository.FetchUserWithUserID(r.Context(), db, userName)
			if err != nil {
				utilities.WriteError("token is not linked to any real user", w, http.StatusUnauthorized)
				w.Header().Set("Content-Type", "application/json")
				return
			}
			ctx := context.WithValue(r.Context(), "userId", user.Id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CorsMiddleWare(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)

		})

	}
}

func extractIP(r *http.Request) string {
	// prefer X-Forwarded-For if present (trusting proxy is another concern)
	if f := r.Header.Get("X-Forwarded-For"); f != "" {
		parts := strings.Split(f, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

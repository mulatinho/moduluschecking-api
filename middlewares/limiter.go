package middlewares

import (
	"net/http"
	"strconv"

	"github.com/etcinit/speedbump"
	"github.com/mulatinho/moduluschecking-api/responses"
	"gopkg.in/redis.v5"
)

type Limiter struct {
	redisConnexion *redis.Client
}

// This middleware check a client is not over the rate limit
func NewLimiter(redisConnexion *redis.Client) *Limiter {
	return &Limiter{
		redisConnexion: redisConnexion,
	}
}

// The middleware handler
func (l *Limiter) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	apiKey := req.Header.Get("Api-Key")

	// Here, in real life, we would typically create a limiter
	// based on information from the given API key
	hasher := speedbump.PerMinuteHasher{}
	limiter := speedbump.NewLimiter(l.redisConnexion, hasher, 5)

	canAccess, err := limiter.Attempt(apiKey)
	// Trouble with Redis?
	if err != nil {
		respondRedisError(w)
		return
	}

	// Over the rate limit?
	if !canAccess {
		responses.WriteMessage(429, "rate_exceeded", "API rate exceeded. Too many requests.", w)
		return
	}

	requestsLeft, err := limiter.Left(apiKey)
	// Trouble with Redis?
	if err != nil {
		respondRedisError(w)
		return
	}
	// Add the number of remaining request as a header
	w.Header().Set("Api-Remaining", strconv.Itoa(int(requestsLeft)))

	// Call the next middleware handler
	next(w, req)
}

// Say that we got an error contacting Redis
func respondRedisError(w http.ResponseWriter) {
	responses.WriteMessage(http.StatusInternalServerError, "server_error", "Trouble contacting Redis. Aborting.", w)
	return
}

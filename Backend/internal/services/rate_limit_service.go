package services

import (
	"sync"
	"time"

	"filevault-backend/internal/config"

	"golang.org/x/time/rate"
)

type RateLimitService struct {
	cfg      *config.Config
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

type RateLimitResult struct {
	Allowed   bool
	Remaining int
	ResetTime time.Time
}

func NewRateLimitService(cfg *config.Config) *RateLimitService {
	return &RateLimitService{
		cfg:      cfg,
		limiters: make(map[string]*rate.Limiter),
	}
}

func (s *RateLimitService) Close() {
}

func (s *RateLimitService) CheckRateLimit(identifier string) *RateLimitResult {
	if !s.cfg.RateLimitEnabled {
		return &RateLimitResult{Allowed: true, Remaining: 999, ResetTime: time.Now().Add(time.Second)}
	}

	limiter := s.getLimiter(identifier)
	allowed := limiter.Allow()
	remaining := int(limiter.TokensAt(time.Now()))
	if remaining < 0 {
		remaining = 0
	}

	resetTime := time.Now().Add(time.Duration(float64(time.Second) / s.cfg.RateLimitPerSecond))

	return &RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: resetTime,
	}
}

func (s *RateLimitService) getLimiter(identifier string) *rate.Limiter {
	s.mu.RLock()
	limiter, exists := s.limiters[identifier]
	s.mu.RUnlock()

	if exists {
		return limiter
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := s.limiters[identifier]; exists {
		return limiter
	}

	// Create new limiter with the configured rate
	limiter = rate.NewLimiter(rate.Limit(s.cfg.RateLimitPerSecond), s.cfg.RateLimitBurstSize)
	s.limiters[identifier] = limiter
	return limiter
}

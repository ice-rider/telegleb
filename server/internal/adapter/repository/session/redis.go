package session

import (
	"encoding/json"
	"context"
	"errors"
	"time"
	"telegleb/internal/core/domain"
	"telegleb/internal/core/usecase/session"

	"github.com/redis/go-redis/v9"
)

var ErrSessionNotFound = errors.New("session not found")

type RedisSessionRepository struct {
	client *redis.Client
}

func NewRedisSessionRepository(client *redis.Client) session.SessionRepository {
	return &RedisSessionRepository{
		client: client,
	}
}

func (r *RedisSessionRepository) CreateSession(ctx context.Context, s *domain.AuthSession) error {
	return r.UpdateSession(ctx, s)
}

func (r *RedisSessionRepository) GetSessionByToken(ctx context.Context, token string) (*domain.AuthSession, error) {
	data, err := r.client.Get(ctx, token).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	var s domain.AuthSession
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	if time.Now().After(s.ExpiresAt) {
		_ = r.DeleteSession(ctx, token)
		return nil, ErrSessionNotFound
	}

	return &s, nil
}

func (r *RedisSessionRepository) UpdateSession(ctx context.Context, s *domain.AuthSession) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	ttl := time.Until(s.ExpiresAt)
	if ttl <= 0 {
		return ErrSessionNotFound
	}

	return r.client.Set(ctx, s.SessionToken, data, ttl).Err()
}

func (r *RedisSessionRepository) DeleteSession(ctx context.Context, token string) error {
	return r.client.Del(ctx, token).Err()
}

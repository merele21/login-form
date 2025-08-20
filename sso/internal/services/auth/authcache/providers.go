package authcache

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"loginform/sso/internal/domain/models"
	"loginform/sso/internal/lib/cache"
	"loginform/sso/internal/services/auth"
)

type CachedUserProvider struct {
	inner auth.UserProvider
	cache cache.Cache
	ttl   time.Duration
}

type CachedAppProvider struct {
	inner auth.AppProvider
	cache cache.Cache
	ttl   time.Duration
}

func NewCachedUserProvider(inner auth.UserProvider, c cache.Cache, ttl time.Duration) *CachedUserProvider {
	return &CachedUserProvider{
		inner: inner,
		cache: c,
		ttl:   ttl,
	}
}

func NewCachedAppProvider(inner auth.AppProvider, c cache.Cache, ttl time.Duration) *CachedAppProvider {
	return &CachedAppProvider{
		inner: inner,
		cache: c,
		ttl:   ttl,
	}
}

func (p *CachedUserProvider) User(ctx context.Context, email string) (models.User, error) {
	key := "user:by-email:" + strings.ToLower(strings.TrimSpace(email))

	// берем из кэша (redis)
	if s, err := p.cache.Get(ctx, key); err == nil && s != "" {
		var user models.User

		if json.Unmarshal([]byte(s), &user) == nil {
			return user, nil
		} // если раскодировать не получилось (вдруг данных нет или еще что-то, то просто идем в БД)
	}

	// идем в БД
	u, err := p.inner.User(ctx, email)
	if err != nil {
		return models.User{}, err
	}

	// кэшируем данные для следующего раза
	if b, err := json.Marshal(u); err == nil {
		_ = p.cache.SetEX(ctx, key, string(b), p.ttl)
	}

	return u, nil
}

func (p *CachedAppProvider) App(ctx context.Context, appID int) (models.App, error) {
	key := "app:" + strconv.Itoa(appID)

	if s, err := p.cache.Get(ctx, key); err == nil && s != "" {
		var app models.App

		if json.Unmarshal([]byte(s), &app) == nil {
			return app, nil
		}
	}

	a, err := p.inner.App(ctx, appID)
	if err != nil {
		return models.App{}, err
	}

	if b, err := json.Marshal(a); err == nil {
		_ = p.cache.SetEX(ctx, key, string(b), p.ttl)
	}

	return a, nil
}

package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	grpcapp "loginform/sso/internal/app/grpc"
	"loginform/sso/internal/config"
	"loginform/sso/internal/lib/cache/redis"
	"loginform/sso/internal/services/auth"
	"loginform/sso/internal/services/auth/authcache"
	"loginform/sso/internal/storage/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

type App struct {
	GRPCSrv *grpcapp.App
	pool    *pgxpool.Pool
	cache   *redis.Client
}

func New(log *slog.Logger, cfg *config.Config) *App {
	// init storage
	poolCtx, poolCancel := context.WithTimeout(context.Background(), 5*time.Second)
	pool, err := pgxpool.New(poolCtx, cfg.Postgres.DSN)
	poolCancel()
	if err != nil {
		panic(err)
	}

	rc := redis.NewClient(&goredis.Options{
		Addr:     cfg.Redis.Addr,
		DB:       cfg.Redis.DB,
		Password: cfg.Redis.Password,
	})

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	if err = waitForPostgres(pingCtx, pool, 5, 500*time.Millisecond, log); err != nil {
		panic(err)
	}
	if err = waitForRedis(pingCtx, rc, 5, 500*time.Millisecond, log); err != nil {
		panic(err)
	}

	pgStore := postgres.New(pool)
	userSaver, userProvider, appProvider := pgStore, pgStore, pgStore

	cachedUserProv := authcache.NewCachedUserProvider(userProvider, rc, 2*time.Minute)
	cachedAppProv := authcache.NewCachedAppProvider(appProvider, rc, 2*time.Minute)

	authService := auth.New(log, userSaver, cachedUserProv, cachedAppProv, cfg.TokenTTL)
	grpcApp := grpcapp.New(log, authService, cfg.GRPC.Port)

	return &App{
		GRPCSrv: grpcApp,
		pool:    pool,
		cache:   rc,
	}
}

func (a *App) Stop() {
	// останавливаем gRPC (graceful)
	a.GRPCSrv.Stop() // у grpcapp (пакета /app/grpc/app.go) уже есть GracefulStop внутри Stop()
	// закрываем внешние ресурсы
	if a.pool != nil {
		a.pool.Close()
	}

	if a.cache != nil {
		_ = a.cache.Close() // также закрываем redis
	}
}

func waitForPostgres(ctx context.Context, pool *pgxpool.Pool, attempts int, baseDelay time.Duration, log *slog.Logger) error {
	for i := 1; i <= attempts; i++ {
		// небольшие таймауты на попытки чтобы не зависать внутри общего ctx
		attemptCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		err := pool.Ping(attemptCtx)
		cancel()

		if err == nil {
			log.Info("postgres connection successful")
			return nil
		}

		if i == attempts || errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("postgres ping failed after %d attempts: %w", attempts, err)
		}

		d := time.Duration(i) * baseDelay // линейная/экспоненциальная задержка
		log.Warn("postgres not ready, retrying ...",
			slog.Int("attempts", i),
			slog.Duration("sleep", d),
			slog.String("err", err.Error()),
		)

		time.Sleep(d)
	}

	return fmt.Errorf("unreachable")
}

func waitForRedis(ctx context.Context, rc interface{ Ping(context.Context) error }, attempts int, baseDelay time.Duration, log *slog.Logger) error {
	for i := 1; i <= attempts; i++ {
		attemptCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
		err := rc.Ping(attemptCtx)
		cancel()

		if err == nil {
			log.Info("redis connection successful")
			return nil
		}

		if i == attempts || errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("redis ping failed after %d attempts: %w", attempts, err)
		}

		d := time.Duration(i) * baseDelay
		log.Warn("redis not ready, retrying ...",
			slog.Int("attempts", i),
			slog.Duration("sleep", d),
			slog.String("err", err.Error()),
		)

		time.Sleep(d)
	}

	return fmt.Errorf("unreachable")
}

package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"loginform/sso/internal/app"
	"loginform/sso/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// init cfg object
	cfg := config.MustLoad()

	fmt.Println(cfg)

	// init logger (slog)
	log := setupLogger(cfg.Env)

	attrs := []slog.Attr{
		slog.String("env", cfg.Env),
		// slog.Any("cfg", cfg), // могут храниться секретные данные
		slog.Int("port", cfg.GRPC.Port),
	}

	log.Info("starting application", slog.GroupAttrs("config", attrs...))

	// init app
	application := app.New(log, cfg)

	// run gRPC-server
	go application.GRPCSrv.MustRun()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	note := <-stop
	log.Info("stopping application", slog.String("signal", note.String()))
	application.Stop()
	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

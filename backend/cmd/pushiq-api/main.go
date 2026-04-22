package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/pushiq/pushiq-backend/internal/api"
    "github.com/pushiq/pushiq-backend/internal/config"
    "github.com/pushiq/pushiq-backend/internal/delivery"
    "github.com/pushiq/pushiq-backend/internal/device"
    "github.com/pushiq/pushiq-backend/internal/repository"
    "github.com/pushiq/pushiq-backend/internal/util"
    "github.com/sirupsen/logrus"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        logrus.Fatalf("failed to load config: %v", err)
    }

    logger := util.NewLogger(cfg.Environment)

    repo, err := repository.NewPostgresRepository(cfg.DatabaseURL)
    if err != nil {
        logger.Fatalf("failed to initialize repository: %v", err)
    }
    defer repo.Close()

    deliveryEngine, err := delivery.NewEngine(cfg, logger)
    if err != nil {
        logger.Fatalf("failed to initialize delivery engine: %v", err)
    }

    // Initialize device lifecycle service
    lifecycleService := device.NewLifecycleService(repo, logger)

    server := api.NewServer(cfg, repo, deliveryEngine, lifecycleService, logger)

    // Initialize and start retry engine
    retryEngine := delivery.NewRetryEngine(repo, deliveryEngine, logger)
    retryCtx, retryCancel := context.WithCancel(context.Background())
    retryEngine.Start(retryCtx, 30*time.Second)

    // Start device lifecycle service
    lifecycleService.Start()

    httpServer := &http.Server{
        Addr:         ":8080",
        Handler:      server.Router(),
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    go func() {
        logger.Infof("starting PushIQ API server on %s", httpServer.Addr)
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatalf("server failed: %v", err)
        }
    }()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
    <-stop

    logger.Info("shutting down server")
    retryCancel()
    retryEngine.Stop()
    lifecycleService.Stop()
    
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer shutdownCancel()

    if err := httpServer.Shutdown(shutdownCtx); err != nil {
        logger.Fatalf("shutdown failure: %v", err)
    }
}

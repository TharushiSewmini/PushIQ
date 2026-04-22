package api

import (
    "github.com/gorilla/mux"
    "github.com/pushiq/pushiq-backend/internal/config"
    "github.com/pushiq/pushiq-backend/internal/delivery"
    "github.com/pushiq/pushiq-backend/internal/repository"
    "github.com/pushiq/pushiq-backend/internal/device"
    "github.com/sirupsen/logrus"
)

type Server struct {
    cfg                *config.Config
    router             *mux.Router
    repo               *repository.Repository
    deliveryEngine     *delivery.Engine
    lifecycleService   *device.LifecycleService
    logger             *logrus.Logger
}

func NewServer(cfg *config.Config, repo *repository.Repository, deliveryEngine *delivery.Engine, lifecycleService *device.LifecycleService, logger *logrus.Logger) *Server {
    server := &Server{
        cfg:                cfg,
        router:             mux.NewRouter(),
        repo:               repo,
        deliveryEngine:     deliveryEngine,
        lifecycleService:   lifecycleService,
        logger:             logger,
    }
    server.routes()
    return server
}

func (s *Server) Router() *mux.Router {
    return s.router
}

func (s *Server) routes() {
    apiRouter := s.router.PathPrefix("/api/v1").Subrouter()
    apiRouter.Use(apiKeyMiddleware(s.cfg.APIKey, s.logger))

    // Device registration
    apiRouter.HandleFunc("/devices/register", s.registerDevice).Methods("POST")

    // Notification endpoints
    apiRouter.HandleFunc("/notifications/send", s.sendNotification).Methods("POST")
    apiRouter.HandleFunc("/notifications/batch-send", s.batchSend).Methods("POST")
    apiRouter.HandleFunc("/notifications/status", s.getNotificationStatus).Methods("POST")

    // M3: Device Management endpoints
    apiRouter.HandleFunc("/devices", s.ListDevices).Methods("GET")
    apiRouter.HandleFunc("/devices/{deviceID}/presence", s.UpdateDevicePresence).Methods("PUT")
    apiRouter.HandleFunc("/devices/{deviceID}/presence", s.GetDevicePresence).Methods("GET")
    apiRouter.HandleFunc("/devices/{deviceID}/history", s.GetDeviceHistory).Methods("GET")
    apiRouter.HandleFunc("/tokens/{tokenID}/expiration", s.SetTokenExpiration).Methods("PUT")
    apiRouter.HandleFunc("/tokens/cleanup", s.CleanupStaleTokens).Methods("POST")

    // M4: Analytics endpoints
    apiRouter.HandleFunc("/analytics/dashboard", s.GetAnalyticsDashboard).Methods("GET")
    apiRouter.HandleFunc("/analytics/metrics", s.GetDeliveryMetrics).Methods("GET")
    apiRouter.HandleFunc("/analytics/funnel", s.GetDeliveryFunnel).Methods("GET")
    apiRouter.HandleFunc("/analytics/platform", s.GetPlatformAnalytics).Methods("GET")
    apiRouter.HandleFunc("/analytics/retry", s.GetRetryAnalytics).Methods("GET")
    apiRouter.HandleFunc("/analytics/top-notifications", s.GetTopNotifications).Methods("GET")
    apiRouter.HandleFunc("/analytics/trends", s.GetHourlyTrends).Methods("GET")

    // Health check (no auth required)
    s.router.HandleFunc("/health", s.HealthCheck).Methods("GET")

    // Webhook endpoints (no auth required)
    s.router.HandleFunc("/webhooks/fcm", s.fcmWebhook).Methods("POST")
    s.router.HandleFunc("/webhooks/apns", s.apnsWebhook).Methods("POST")
}

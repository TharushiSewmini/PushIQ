package delivery

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pushiq/pushiq-backend/internal/config"
	"github.com/pushiq/pushiq-backend/internal/model"
	"github.com/sirupsen/logrus"
)

type DeliveryRequest struct {
	DeviceID uuid.UUID
	Token    string
	Platform model.Platform
	Title    string
	Body     string
	Data     map[string]any
	Priority string
}

type ProviderResponse struct {
	Provider          model.DeliveryProvider
	ProviderMessageID string
	RawResponse       map[string]any
}

type Provider interface {
	Send(ctx context.Context, request DeliveryRequest) (*ProviderResponse, error)
}

type Engine struct {
	providers map[model.Platform]Provider
	logger    *logrus.Logger
}

type MockProvider struct {
	provider model.DeliveryProvider
	logger   *logrus.Logger
}

func (p *MockProvider) Send(ctx context.Context, request DeliveryRequest) (*ProviderResponse, error) {
	p.logger.WithFields(logrus.Fields{
		"platform": request.Platform,
		"device":   request.DeviceID,
		"token":    request.Token,
	}).Info("mock delivery provider accepted notification")

	return &ProviderResponse{
		Provider:          p.provider,
		ProviderMessageID: "mock-message-id",
		RawResponse: map[string]any{
			"mock":   true,
			"status": "accepted",
		},
	}, nil
}

func NewEngine(cfg *config.Config, logger *logrus.Logger) (*Engine, error) {
	providers := map[model.Platform]Provider{}

	if shouldUseMockFCM(cfg) {
		logger.Warn("using mock FCM provider in development; real Android notifications are disabled")
		providers[model.PlatformAndroid] = &MockProvider{provider: model.ProviderFCM, logger: logger}
	} else {
		providers[model.PlatformAndroid] = NewFCMProvider(cfg.FCMServerKey, logger)
	}

	if shouldUseMockAPNS(cfg) {
		logger.Warn("using mock APNS provider in development; real iOS notifications are disabled")
		providers[model.PlatformiOS] = &MockProvider{provider: model.ProviderAPNS, logger: logger}
	} else {
		apnsProvider, err := NewAPNSProvider(cfg, logger)
		if err != nil {
			return nil, err
		}
		providers[model.PlatformiOS] = apnsProvider
	}

	return &Engine{providers: providers, logger: logger}, nil
}

func shouldUseMockFCM(cfg *config.Config) bool {
	if cfg.Environment != "development" {
		return false
	}
	key := strings.TrimSpace(cfg.FCMServerKey)
	return key == "" || strings.Contains(strings.ToLower(key), "dummy")
}

func shouldUseMockAPNS(cfg *config.Config) bool {
	if cfg.Environment != "development" {
		return false
	}

	path := strings.TrimSpace(cfg.APNSKeyPath)
	keyID := strings.TrimSpace(cfg.APNSKeyID)
	teamID := strings.TrimSpace(cfg.APNSTeamID)
	topic := strings.TrimSpace(cfg.APNSTopic)

	if path == "" || keyID == "" || teamID == "" || topic == "" {
		return true
	}

	placeholderPath := strings.Contains(path, "/path/to/")
	placeholderTopic := strings.Contains(strings.ToLower(topic), "example")
	return placeholderPath || placeholderTopic
}

func (e *Engine) Send(ctx context.Context, request DeliveryRequest) (*ProviderResponse, error) {
	provider, ok := e.providers[request.Platform]
	if !ok {
		return nil, fmt.Errorf("unsupported platform: %s", request.Platform)
	}
	return provider.Send(ctx, request)
}

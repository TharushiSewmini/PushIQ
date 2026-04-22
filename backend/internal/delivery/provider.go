package delivery

import (
	"context"
	"fmt"

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

func NewEngine(cfg *config.Config, logger *logrus.Logger) (*Engine, error) {
	providers := map[model.Platform]Provider{}

	fcmProvider := NewFCMProvider(cfg.FCMServerKey, logger)
	providers[model.PlatformAndroid] = fcmProvider

	apnsProvider, err := NewAPNSProvider(cfg, logger)
	if err != nil {
		return nil, err
	}
	providers[model.PlatformiOS] = apnsProvider

	return &Engine{providers: providers, logger: logger}, nil
}

func (e *Engine) Send(ctx context.Context, request DeliveryRequest) (*ProviderResponse, error) {
	provider, ok := e.providers[request.Platform]
	if !ok {
		return nil, fmt.Errorf("unsupported platform: %s", request.Platform)
	}
	return provider.Send(ctx, request)
}

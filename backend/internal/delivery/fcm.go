package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pushiq/pushiq-backend/internal/model"
	"github.com/sirupsen/logrus"
)

type FCMProvider struct {
	serverKey string
	logger    *logrus.Logger
	client    *http.Client
}

type fcmRequest struct {
	To           string            `json:"to"`
	Notification map[string]string `json:"notification"`
	Data         map[string]any    `json:"data,omitempty"`
	Priority     string            `json:"priority,omitempty"`
}

type fcmResponse struct {
	Success int   `json:"success"`
	Failure int   `json:"failure"`
	Results []any `json:"results"`
}

func NewFCMProvider(serverKey string, logger *logrus.Logger) *FCMProvider {
	return &FCMProvider{
		serverKey: serverKey,
		logger:    logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *FCMProvider) Send(ctx context.Context, request DeliveryRequest) (*ProviderResponse, error) {
	payload := fcmRequest{
		To: request.Token,
		Notification: map[string]string{
			"title": request.Title,
			"body":  request.Body,
		},
		Data:     request.Data,
		Priority: request.Priority,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://fcm.googleapis.com/fcm/send", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("key=%s", p.serverKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fcm server returned %d", resp.StatusCode)
	}

	var result fcmResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Failure > 0 {
		p.logger.WithFields(logrus.Fields{
			"token":    request.Token,
			"response": result,
		}).Warn("FCM returned failures")
	}

	raw := map[string]any{
		"success": result.Success,
		"failure": result.Failure,
		"results": result.Results,
	}

	return &ProviderResponse{
		Provider:          model.ProviderFCM,
		ProviderMessageID: "",
		RawResponse:       raw,
	}, nil
}

package delivery

import (
	"context"
	"math"
	"time"

	"github.com/pushiq/pushiq-backend/internal/model"
	"github.com/pushiq/pushiq-backend/internal/repository"
	"github.com/sirupsen/logrus"
)

type RetryEngine struct {
	repo   *repository.Repository
	engine *Engine
	logger *logrus.Logger
	stopCh chan struct{}
}

func NewRetryEngine(repo *repository.Repository, engine *Engine, logger *logrus.Logger) *RetryEngine {
	return &RetryEngine{
		repo:   repo,
		engine: engine,
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

func (re *RetryEngine) Start(ctx context.Context, tickInterval time.Duration) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-re.stopCh:
				return
			case <-ticker.C:
				re.processRetries(ctx)
			}
		}
	}()
}

func (re *RetryEngine) Stop() {
	close(re.stopCh)
}

func (re *RetryEngine) processRetries(ctx context.Context) {
	notifications, err := re.repo.GetPendingRetries(100)
	if err != nil {
		re.logger.WithError(err).Error("failed to fetch pending retries")
		return
	}

	for _, notification := range notifications {
		re.retryNotification(ctx, notification)
	}
}

func (re *RetryEngine) retryNotification(ctx context.Context, notification model.Notification) {
	attemptNumber := notification.AttemptCount + 1

	re.logger.WithFields(logrus.Fields{
		"notification_id": notification.ID,
		"attempt":         attemptNumber,
		"max_retries":     notification.MaxRetries,
	}).Info("retrying notification")

	// Get current token for device
	deviceToken, err := re.repo.GetActiveDeviceToken(notification.DeviceID, string(notification.Platform))
	if err != nil {
		re.logger.WithError(err).Error("failed to get device token for retry")
		_ = re.repo.RecordDeliveryAttempt(notification.ID, attemptNumber, "failed", ptrStr(err.Error()), nil)
		re.scheduleNextRetry(&notification, attemptNumber)
		return
	}

	if deviceToken == nil {
		re.logger.WithFields(logrus.Fields{
			"notification_id": notification.ID,
		}).Warn("no active token found for device")
		_ = re.repo.RecordDeliveryAttempt(notification.ID, attemptNumber, "failed", ptrStr("no active token"), nil)
		re.scheduleNextRetry(&notification, attemptNumber)
		return
	}

	// Retry delivery
	response, err := re.engine.Send(ctx, DeliveryRequest{
		DeviceID: notification.DeviceID,
		Token:    deviceToken.Token,
		Platform: notification.Platform,
		Title:    notification.Title,
		Body:     notification.Body,
		Data:     notification.Data,
		Priority: "",
	})

	if err != nil {
		re.logger.WithError(err).Error("delivery engine error on retry")
		_ = re.repo.RecordDeliveryAttempt(notification.ID, attemptNumber, "failed", ptrStr(err.Error()), nil)

		if attemptNumber < notification.MaxRetries {
			re.scheduleNextRetry(&notification, attemptNumber)
		} else {
			_ = re.repo.UpdateNotificationStatus(notification.ID, model.NotificationStatusFailed, map[string]any{"error": err.Error()})
		}
		return
	}

	// Success
	_ = re.repo.RecordDeliveryAttempt(notification.ID, attemptNumber, "sent", nil, response.RawResponse)
	_ = re.repo.UpdateNotificationStatus(notification.ID, model.NotificationStatusSent, response.RawResponse)

	re.logger.WithFields(logrus.Fields{
		"notification_id": notification.ID,
		"attempt":         attemptNumber,
	}).Info("notification sent on retry")
}

func (re *RetryEngine) scheduleNextRetry(notification *model.Notification, currentAttempt int) {
	if currentAttempt >= notification.MaxRetries {
		return
	}

	// Exponential backoff: 60s, 300s, 900s for attempts 1, 2, 3
	backoffSeconds := int(math.Pow(5, float64(currentAttempt))) * 10
	nextRetryTime := time.Now().UTC().Add(time.Duration(backoffSeconds) * time.Second)

	if err := re.repo.UpdateNotificationRetry(notification.ID, currentAttempt, nextRetryTime); err != nil {
		re.logger.WithError(err).Error("failed to schedule next retry")
	}
}

func ptrStr(s string) *string {
	return &s
}

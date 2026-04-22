package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pushiq/pushiq-backend/internal/delivery"
	"github.com/pushiq/pushiq-backend/internal/model"
)

type batchSendRequest struct {
	TenantID      string `json:"tenant_id,omitempty"`
	Notifications []struct {
		DeviceID string         `json:"device_id,omitempty"`
		Token    string         `json:"token,omitempty"`
		Platform string         `json:"platform,omitempty"`
		Title    string         `json:"title"`
		Body     string         `json:"body"`
		Data     map[string]any `json:"data,omitempty"`
		Priority string         `json:"priority,omitempty"`
	} `json:"notifications"`
}

type batchSendResponse struct {
	NotificationCount int      `json:"notification_count"`
	SuccessCount      int      `json:"success_count"`
	FailedCount       int      `json:"failed_count"`
	Message           string   `json:"message"`
	NotificationIDs   []string `json:"notification_ids,omitempty"`
}

type notificationStatusRequest struct {
	NotificationID string `json:"notification_id"`
}

type notificationStatusResponse struct {
	NotificationID string     `json:"notification_id"`
	Status         string     `json:"status"`
	Platform       string     `json:"platform"`
	AttemptCount   int        `json:"attempt_count"`
	MaxRetries     int        `json:"max_retries"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type webhookRequest struct {
	Message map[string]any `json:"message,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

func (s *Server) batchSend(w http.ResponseWriter, r *http.Request) {
	var payload batchSendRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(payload.Notifications) == 0 {
		http.Error(w, "notifications array is empty", http.StatusBadRequest)
		return
	}

	successCount := 0
	failedCount := 0
	notificationIDs := []string{}

	for _, notif := range payload.Notifications {
		notification := &model.Notification{
			TenantID:   sql.NullString{Valid: payload.TenantID != "", String: payload.TenantID},
			Platform:   model.Platform(strings.ToLower(notif.Platform)),
			Title:      notif.Title,
			Body:       notif.Body,
			Data:       notif.Data,
			Status:     model.NotificationStatusPending,
			MaxRetries: 3,
		}

		if notif.DeviceID != "" {
			if deviceID, err := uuid.Parse(notif.DeviceID); err == nil {
				notification.DeviceID = deviceID
			}
		}

		if err := s.repo.CreateNotification(notification); err != nil {
			s.logger.WithError(err).Error("failed to persist notification")
			failedCount++
			continue
		}

		token := notif.Token
		if token == "" && notif.DeviceID != "" {
			deviceToken, err := s.repo.GetActiveDeviceToken(notification.DeviceID, string(notification.Platform))
			if err != nil || deviceToken == nil {
				failedCount++
				continue
			}
			token = deviceToken.Token
		}

		providerResponse, err := s.deliveryEngine.Send(context.Background(), delivery.DeliveryRequest{
			DeviceID: notification.DeviceID,
			Token:    token,
			Platform: notification.Platform,
			Title:    notif.Title,
			Body:     notif.Body,
			Data:     notif.Data,
			Priority: notif.Priority,
		})

		if err != nil {
			s.logger.WithError(err).Error("failed to deliver notification")
			_ = s.repo.RecordDeliveryAttempt(notification.ID, 1, "failed", ptrStr(err.Error()), nil)
			_ = s.repo.UpdateNotificationRetry(notification.ID, 1, time.Now().UTC().Add(60*time.Second))
			failedCount++
		} else {
			notification.Provider = providerResponse.Provider
			notification.Status = model.NotificationStatusSent
			notification.ProviderResponse = providerResponse.RawResponse
			notification.SentAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
			_ = s.repo.UpdateNotificationStatus(notification.ID, model.NotificationStatusSent, providerResponse.RawResponse)
			successCount++
		}

		notificationIDs = append(notificationIDs, notification.ID.String())
	}

	respondJSON(w, http.StatusOK, batchSendResponse{
		NotificationCount: len(payload.Notifications),
		SuccessCount:      successCount,
		FailedCount:       failedCount,
		Message:           "batch send completed",
		NotificationIDs:   notificationIDs,
	})
}

func (s *Server) getNotificationStatus(w http.ResponseWriter, r *http.Request) {
	var payload notificationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if payload.NotificationID == "" {
		http.Error(w, "notification_id is required", http.StatusBadRequest)
		return
	}

	notificationID, err := uuid.Parse(payload.NotificationID)
	if err != nil {
		http.Error(w, "invalid notification_id", http.StatusBadRequest)
		return
	}

	notification, err := s.repo.GetNotificationByID(notificationID)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch notification")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if notification == nil {
		http.Error(w, "notification not found", http.StatusNotFound)
		return
	}

	response := notificationStatusResponse{
		NotificationID: notification.ID.String(),
		Status:         string(notification.Status),
		Platform:       string(notification.Platform),
		AttemptCount:   notification.AttemptCount,
		MaxRetries:     notification.MaxRetries,
		CreatedAt:      notification.CreatedAt,
	}

	if notification.SentAt.Valid {
		response.SentAt = &notification.SentAt.Time
	}
	if notification.DeliveredAt.Valid {
		response.DeliveredAt = &notification.DeliveredAt.Time
	}

	respondJSON(w, http.StatusOK, response)
}

func (s *Server) fcmWebhook(w http.ResponseWriter, r *http.Request) {
	var payload webhookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) apnsWebhook(w http.ResponseWriter, r *http.Request) {
	var payload webhookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func ptrStr(s string) *string {
	return &s
}

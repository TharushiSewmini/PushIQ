package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pushiq/pushiq-backend/internal/delivery"
	"github.com/pushiq/pushiq-backend/internal/model"
)

type registerDeviceRequest struct {
	TenantID   string `json:"tenant_id,omitempty"`
	UserID     string `json:"user_id"`
	Platform   string `json:"platform"`
	Token      string `json:"token"`
	AppVersion string `json:"app_version,omitempty"`
	Locale     string `json:"locale,omitempty"`
}

type registerDeviceResponse struct {
	DeviceID string `json:"device_id"`
}

type sendNotificationRequest struct {
	TenantID string         `json:"tenant_id,omitempty"`
	DeviceID string         `json:"device_id,omitempty"`
	Platform string         `json:"platform,omitempty"`
	Token    string         `json:"token,omitempty"`
	Title    string         `json:"title"`
	Body     string         `json:"body"`
	Data     map[string]any `json:"data,omitempty"`
	Priority string         `json:"priority,omitempty"`
}

type sendNotificationResponse struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
}

func (s *Server) registerDevice(w http.ResponseWriter, r *http.Request) {
	var payload registerDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if payload.UserID == "" || payload.Platform == "" || payload.Token == "" {
		http.Error(w, "user_id, platform, and token are required", http.StatusBadRequest)
		return
	}

	platform := model.Platform(payload.Platform)
	if platform != model.PlatformAndroid && platform != model.PlatformiOS {
		http.Error(w, "platform must be 'android' or 'ios'", http.StatusBadRequest)
		return
	}

	tenantID := sql.NullString{Valid: payload.TenantID != "", String: payload.TenantID}
	device, err := s.repo.GetDeviceByUserAndPlatform(payload.UserID, string(platform))
	if err != nil {
		s.logger.WithError(err).Error("failed to query device")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if device == nil {
		device = &model.Device{
			TenantID:   tenantID,
			UserID:     payload.UserID,
			Platform:   platform,
			AppVersion: payload.AppVersion,
			Locale:     payload.Locale,
		}
	} else {
		device.AppVersion = payload.AppVersion
		device.Locale = payload.Locale
	}

	if err := s.repo.UpsertDevice(device); err != nil {
		s.logger.WithError(err).Error("failed to upsert device")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	token := &model.DeviceToken{
		DeviceID: device.ID,
		Token:    payload.Token,
		Provider: string(platform),
		Status:   model.TokenStatusActive,
	}
	if err := s.repo.UpsertDeviceToken(token); err != nil {
		s.logger.WithError(err).Error("failed to upsert device token")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, registerDeviceResponse{DeviceID: device.ID.String()})
}

func (s *Server) sendNotification(w http.ResponseWriter, r *http.Request) {
	var payload sendNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if payload.Title == "" || payload.Body == "" {
		http.Error(w, "title and body are required", http.StatusBadRequest)
		return
	}
	if payload.Token == "" && payload.DeviceID == "" {
		http.Error(w, "device_id or token is required", http.StatusBadRequest)
		return
	}

	token := payload.Token
	var platform model.Platform
	if token == "" {
		deviceID, err := uuid.Parse(payload.DeviceID)
		if err != nil {
			http.Error(w, "invalid device_id", http.StatusBadRequest)
			return
		}

		device, err := s.repo.GetDeviceByID(deviceID)
		if err != nil {
			s.logger.WithError(err).Error("failed to fetch device")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if device == nil {
			http.Error(w, "device not found", http.StatusNotFound)
			return
		}

		platform = device.Platform
		deviceToken, err := s.repo.GetActiveDeviceToken(deviceID, string(platform))
		if err != nil {
			s.logger.WithError(err).Error("failed to fetch device token")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if deviceToken == nil {
			http.Error(w, "active token not found for device", http.StatusNotFound)
			return
		}

		token = deviceToken.Token
	} else {
		platform = model.Platform(payload.Platform)
		if platform != model.PlatformAndroid && platform != model.PlatformiOS {
			http.Error(w, "platform must be 'android' or 'ios'", http.StatusBadRequest)
			return
		}
	}

	notification := &model.Notification{
		TenantID: sql.NullString{Valid: payload.TenantID != "", String: payload.TenantID},
		Platform: platform,
		Title:    payload.Title,
		Body:     payload.Body,
		Data:     payload.Data,
		Status:   model.NotificationStatusPending,
	}

	if payload.DeviceID != "" {
		if deviceID, err := uuid.Parse(payload.DeviceID); err == nil {
			notification.DeviceID = deviceID
		}
	}

	if err := s.repo.CreateNotification(notification); err != nil {
		s.logger.WithError(err).Error("failed to persist notification")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	providerResponse, err := s.deliveryEngine.Send(context.Background(), delivery.DeliveryRequest{
		DeviceID: notification.DeviceID,
		Token:    token,
		Platform: platform,
		Title:    payload.Title,
		Body:     payload.Body,
		Data:     payload.Data,
		Priority: payload.Priority,
	})
	if err != nil {
		s.logger.WithError(err).Error("failed to deliver notification")
		_ = s.repo.UpdateNotificationStatus(notification.ID, model.NotificationStatusFailed, map[string]any{"error": err.Error()})
		http.Error(w, "delivery failed", http.StatusBadGateway)
		return
	}

	notification.Provider = providerResponse.Provider
	notification.Status = model.NotificationStatusSent
	notification.ProviderResponse = providerResponse.RawResponse
	notification.SentAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	if err := s.repo.UpdateNotificationStatus(notification.ID, notification.Status, notification.ProviderResponse); err != nil {
		s.logger.WithError(err).Warn("failed to update notification status")
	}

	respondJSON(w, http.StatusOK, sendNotificationResponse{NotificationID: notification.ID.String(), Status: string(notification.Status)})
}

func respondJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

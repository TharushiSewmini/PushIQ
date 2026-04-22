package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pushiq/pushiq-backend/internal/model"
)

// M3 API Handlers - Device Management & Token Refresh

type DeviceListResponse struct {
	Devices []model.DeviceWithPresence `json:"devices"`
	Total   int                        `json:"total"`
}

type DevicePresenceRequest struct {
	IsOnline bool `json:"is_online"`
}

type TokenExpirationRequest struct {
	ExpiresInDays int `json:"expires_in_days"`
}

type DeviceHistoryResponse struct {
	DeviceID string                 `json:"device_id"`
	Activity []model.DeviceActivity `json:"activity"`
	Total    int                    `json:"total"`
}

// ListDevices returns all devices for an application with optional online filter
func (s *Server) ListDevices(w http.ResponseWriter, r *http.Request) {
	appID := r.Context().Value("appID")
	if appID == nil {
		http.Error(w, "Missing app ID", http.StatusUnauthorized)
		return
	}

	parsedAppID, err := uuid.Parse(appID.(string))
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}

	onlineOnly := r.URL.Query().Get("online_only") == "true"

	devices, err := s.lifecycleService.ListDevices(parsedAppID, onlineOnly)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list devices")
		http.Error(w, "Failed to retrieve devices", http.StatusInternalServerError)
		return
	}

	response := DeviceListResponse{
		Devices: devices,
		Total:   len(devices),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateDevicePresence updates the presence status of a device
func (s *Server) UpdateDevicePresence(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceIDStr := vars["deviceID"]
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	var req DevicePresenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.lifecycleService.UpdateDevicePresence(deviceID, req.IsOnline); err != nil {
		s.logger.WithError(err).Error("Failed to update device presence")
		http.Error(w, "Failed to update presence", http.StatusInternalServerError)
		return
	}

	presence, err := s.lifecycleService.GetDevicePresence(deviceID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch updated presence")
		http.Error(w, "Failed to fetch presence", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(presence)
}

// GetDevicePresence retrieves the current presence status of a device
func (s *Server) GetDevicePresence(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceIDStr := vars["deviceID"]
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	presence, err := s.lifecycleService.GetDevicePresence(deviceID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get device presence")
		http.Error(w, "Failed to retrieve presence", http.StatusInternalServerError)
		return
	}

	if presence == nil {
		http.Error(w, "Device presence not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(presence)
}

// SetTokenExpiration sets the expiration time for a device token
func (s *Server) SetTokenExpiration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenIDStr := vars["tokenID"]
	tokenID, err := uuid.Parse(tokenIDStr)
	if err != nil {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}

	var req TokenExpirationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ExpiresInDays <= 0 {
		http.Error(w, "expires_in_days must be positive", http.StatusBadRequest)
		return
	}

	// Calculate expiration time
	expirationTime := time.Now().AddDate(0, 0, req.ExpiresInDays)
	if err := s.lifecycleService.SetTokenExpiration(tokenID, expirationTime); err != nil {
		s.logger.WithError(err).Error("Failed to set token expiration")
		http.Error(w, "Failed to set token expiration", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"token_id":   tokenID,
		"expires_at": expirationTime,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetDeviceHistory retrieves the activity history for a device
func (s *Server) GetDeviceHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceIDStr := vars["deviceID"]
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	activities, err := s.lifecycleService.GetDeviceHistory(deviceID, limit)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get device history")
		http.Error(w, "Failed to retrieve device history", http.StatusInternalServerError)
		return
	}

	response := DeviceHistoryResponse{
		DeviceID: deviceID.String(),
		Activity: activities,
		Total:    len(activities),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CleanupStaleTokens triggers manual cleanup of stale device tokens
func (s *Server) CleanupStaleTokens(w http.ResponseWriter, r *http.Request) {
	// Mark expired tokens as invalid
	_, err := s.repo.InvalidateExpiredTokens(0)
	if err != nil {
		s.logger.WithError(err).Error("Failed to invalidate expired tokens")
		http.Error(w, "Failed to cleanup tokens", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status": "cleanup_triggered",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

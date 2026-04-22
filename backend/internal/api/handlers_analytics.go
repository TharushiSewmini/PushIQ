package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// M4 Analytics API Handlers

// GetAnalyticsDashboard returns comprehensive analytics data
func (s *Server) GetAnalyticsDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		http.Error(w, "Missing tenant ID", http.StatusUnauthorized)
		return
	}

	dashboard, err := s.repo.GetAnalyticsDashboard(tenantID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch analytics dashboard")
		http.Error(w, "Failed to retrieve analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// GetDeliveryMetrics returns high-level delivery statistics
func (s *Server) GetDeliveryMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		http.Error(w, "Missing tenant ID", http.StatusUnauthorized)
		return
	}

	metrics, err := s.repo.GetDeliveryMetrics(tenantID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch delivery metrics")
		http.Error(w, "Failed to retrieve metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetDeliveryFunnel returns funnel analysis data
func (s *Server) GetDeliveryFunnel(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		http.Error(w, "Missing tenant ID", http.StatusUnauthorized)
		return
	}

	funnel, err := s.repo.GetDeliveryFunnel(tenantID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch delivery funnel")
		http.Error(w, "Failed to retrieve funnel data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(funnel)
}

// GetPlatformAnalytics returns per-platform breakdown
func (s *Server) GetPlatformAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		http.Error(w, "Missing tenant ID", http.StatusUnauthorized)
		return
	}

	analytics, err := s.repo.GetPlatformAnalytics(tenantID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch platform analytics")
		http.Error(w, "Failed to retrieve platform analytics", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"platforms": analytics,
		"total":     len(analytics),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRetryAnalytics returns retry engine performance metrics
func (s *Server) GetRetryAnalytics(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		http.Error(w, "Missing tenant ID", http.StatusUnauthorized)
		return
	}

	analytics, err := s.repo.GetRetryAnalytics(tenantID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch retry analytics")
		http.Error(w, "Failed to retrieve retry analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// GetTopNotifications returns the most impactful notifications
func (s *Server) GetTopNotifications(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		http.Error(w, "Missing tenant ID", http.StatusUnauthorized)
		return
	}

	limit := 10
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	notifications, err := s.repo.GetTopNotifications(tenantID, limit)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch top notifications")
		http.Error(w, "Failed to retrieve notifications", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"notifications": notifications,
		"total":         len(notifications),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetHourlyTrends returns notification volume over time
func (s *Server) GetHourlyTrends(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		http.Error(w, "Missing tenant ID", http.StatusUnauthorized)
		return
	}

	hours := 24
	if hoursParam := r.URL.Query().Get("hours"); hoursParam != "" {
		if parsed, err := strconv.Atoi(hoursParam); err == nil && parsed > 0 && parsed <= 720 {
			hours = parsed
		}
	}

	trends, err := s.repo.GetHourlyTrends(tenantID, hours)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch hourly trends")
		http.Error(w, "Failed to retrieve trends", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"trends": trends,
		"total":  len(trends),
		"hours":  hours,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthCheck returns API health status
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "healthy",
		"service": "pushiq-api",
		"version": "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

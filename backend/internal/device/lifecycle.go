package device

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pushiq/pushiq-backend/internal/model"
	"github.com/pushiq/pushiq-backend/internal/repository"
	"github.com/sirupsen/logrus"
)

// LifecycleService handles device token lifecycle management
type LifecycleService struct {
	repo   *repository.Repository
	logger *logrus.Logger
	ticker *time.Ticker
	stopCh chan struct{}
	doneCh chan struct{}
}

// NewLifecycleService creates a new device lifecycle service
func NewLifecycleService(repo *repository.Repository, logger *logrus.Logger) *LifecycleService {
	return &LifecycleService{
		repo:   repo,
		logger: logger,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

// Start begins the background lifecycle management worker
func (s *LifecycleService) Start() {
	s.ticker = time.NewTicker(5 * time.Minute)
	go s.worker()
	s.logger.Info("Device lifecycle service started")
}

// Stop gracefully stops the background worker
func (s *LifecycleService) Stop() {
	close(s.stopCh)
	<-s.doneCh
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.logger.Info("Device lifecycle service stopped")
}

// worker runs the background cleanup tasks
func (s *LifecycleService) worker() {
	defer close(s.doneCh)

	for {
		select {
		case <-s.ticker.C:
			s.runCleanupTasks()
		case <-s.stopCh:
			return
		}
	}
}

// runCleanupTasks executes all lifecycle maintenance tasks
func (s *LifecycleService) runCleanupTasks() {
	// Mark expired tokens as invalid
	if err := s.invalidateExpiredTokens(); err != nil {
		s.logger.WithError(err).Error("Failed to invalidate expired tokens")
	}

	// Cleanup stale presence data (no activity for 30 days)
	if err := s.cleanupStalePresence(); err != nil {
		s.logger.WithError(err).Error("Failed to cleanup stale presence")
	}

	// Mark devices as offline if no activity for 24 hours
	if err := s.markInactiveDevicesOffline(); err != nil {
		s.logger.WithError(err).Error("Failed to mark inactive devices offline")
	}
}

// invalidateExpiredTokens marks tokens past expiration as invalid
func (s *LifecycleService) invalidateExpiredTokens() error {
	count, err := s.repo.InvalidateExpiredTokens(0)
	if err != nil {
		return fmt.Errorf("invalidate expired tokens: %w", err)
	}
	if count > 0 {
		s.logger.WithField("count", count).Debug("Invalidated expired tokens")
	}
	return nil
}

// cleanupStalePresence removes presence records for devices with no activity for 30 days
func (s *LifecycleService) cleanupStalePresence() error {
	staleThreshold := time.Now().AddDate(0, 0, -30)
	count, err := s.repo.CleanupStalePresence(staleThreshold)
	if err != nil {
		return fmt.Errorf("cleanup stale presence: %w", err)
	}
	if count > 0 {
		s.logger.WithField("count", count).Debug("Cleaned up stale presence records")
	}
	return nil
}

// markInactiveDevicesOffline marks devices as offline if last_seen > 24 hours
func (s *LifecycleService) markInactiveDevicesOffline() error {
	inactiveThreshold := time.Now().Add(-24 * time.Hour)
	count, err := s.repo.MarkInactiveDevicesOffline(inactiveThreshold)
	if err != nil {
		return fmt.Errorf("mark inactive devices offline: %w", err)
	}
	if count > 0 {
		s.logger.WithField("count", count).Debug("Marked inactive devices offline")
	}
	return nil
}

// UpdateDevicePresence updates device presence data and logs activity
func (s *LifecycleService) UpdateDevicePresence(deviceID uuid.UUID, isOnline bool) error {
	now := time.Now()
	var lastOnlineAt *time.Time
	if isOnline {
		lastOnlineAt = &now
	}

	if err := s.repo.UpsertDevicePresence(&deviceID, isOnline, lastOnlineAt); err != nil {
		return fmt.Errorf("update device presence: %w", err)
	}

	// Log activity
	activity := model.DeviceActivity{
		ID:           uuid.New(),
		DeviceID:     deviceID,
		ActivityType: "presence_update",
		Details: map[string]interface{}{
			"is_online": isOnline,
		},
		CreatedAt: now,
	}

	if err := s.repo.LogDeviceActivity(activity); err != nil {
		s.logger.WithError(err).Warn("Failed to log device activity")
		// Don't fail the presence update if logging fails
	}

	return nil
}

// SetTokenExpiration sets the expiration time for a device token
func (s *LifecycleService) SetTokenExpiration(tokenID uuid.UUID, expiresAt time.Time) error {
	activity := model.DeviceActivity{
		ID:           uuid.New(),
		ActivityType: "token_expiration_set",
		Details: map[string]interface{}{
			"token_id":   tokenID.String(),
			"expires_at": expiresAt,
		},
		CreatedAt: time.Now(),
	}

	if err := s.repo.UpdateTokenExpiration(tokenID, expiresAt); err != nil {
		return fmt.Errorf("set token expiration: %w", err)
	}

	if err := s.repo.LogDeviceActivity(activity); err != nil {
		s.logger.WithError(err).Warn("Failed to log token expiration activity")
	}

	return nil
}

// GetDeviceHistory retrieves the activity history for a device
func (s *LifecycleService) GetDeviceHistory(deviceID uuid.UUID, limit int) ([]model.DeviceActivity, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	return s.repo.GetDeviceActivityHistory(deviceID, limit)
}

// ListDevices retrieves all devices with optional presence filter
func (s *LifecycleService) ListDevices(appID uuid.UUID, onlineOnly bool) ([]model.DeviceWithPresence, error) {
	return s.repo.ListDevicesWithPresence(appID, onlineOnly)
}

// GetDevicePresence retrieves the presence state of a device
func (s *LifecycleService) GetDevicePresence(deviceID uuid.UUID) (*model.DevicePresence, error) {
	return s.repo.GetDevicePresence(deviceID)
}

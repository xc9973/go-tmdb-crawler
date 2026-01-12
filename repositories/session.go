package repositories

import (
	"time"

	"github.com/xc9973/go-tmdb-crawler/models"
	"gorm.io/gorm"
)

// SessionRepository defines the interface for session data operations
type SessionRepository interface {
	Create(session *models.Session) error
	GetByID(id uint) (*models.Session, error)
	GetBySessionID(sessionID string) (*models.Session, error)
	GetByToken(token string) (*models.Session, error)
	Update(session *models.Session) error
	Delete(id uint) error
	DeleteBySessionID(sessionID string) error
	DeleteExpired() error
	Count() (int64, error)
	CountActive() (int64, error)
}

type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new session repository instance
func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

// Create creates a new session
func (r *sessionRepository) Create(session *models.Session) error {
	return r.db.Create(session).Error
}

// GetByID retrieves a session by ID
func (r *sessionRepository) GetByID(id uint) (*models.Session, error) {
	var session models.Session
	err := r.db.First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetBySessionID retrieves a session by session ID
func (r *sessionRepository) GetBySessionID(sessionID string) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetByToken retrieves a session by token
func (r *sessionRepository) GetByToken(token string) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Update updates a session
func (r *sessionRepository) Update(session *models.Session) error {
	return r.db.Save(session).Error
}

// Delete deletes a session by ID
func (r *sessionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Session{}, id).Error
}

// DeleteBySessionID deletes a session by session ID
func (r *sessionRepository) DeleteBySessionID(sessionID string) error {
	return r.db.Where("session_id = ?", sessionID).Delete(&models.Session{}).Error
}

// DeleteExpired deletes all expired sessions
func (r *sessionRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error
}

// Count returns the total number of sessions
func (r *sessionRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Session{}).Count(&count).Error
	return count, err
}

// CountActive returns the number of active (non-expired) sessions
func (r *sessionRepository) CountActive() (int64, error) {
	var count int64
	err := r.db.Model(&models.Session{}).
		Where("expires_at > ?", time.Now()).
		Count(&count).Error
	return count, err
}

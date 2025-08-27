package repositories

import (
	"time"

	"gorm.io/gorm"
	"online_medical_consultation_app/backend/internal/models"
)

type VideoSessionRepository interface {
	Create(session *models.VideoSession) error
	FindByID(id uint) (*models.VideoSession, error)
	FindByAppointmentID(appointmentID uint) ([]models.VideoSession, error)
	Update(session *models.VideoSession) error
	Delete(id uint) error
	LoadRelations(session *models.VideoSession) error
	FindActiveByAppointment(appointmentID uint) (*models.VideoSession, error)
	FindByRoomID(roomID string) (*models.VideoSession, error)
	UpdateStartedAt(sessionID uint, startedAt *time.Time) error
	UpdateEndedAt(sessionID uint, endedAt *time.Time) error
}

type videoSessionRepository struct {
	db *gorm.DB
}

func NewVideoSessionRepository(db *gorm.DB) VideoSessionRepository {
	return &videoSessionRepository{
		db: db,
	}
}

// Create ビデオセッションの作成
func (r *videoSessionRepository) Create(videoSession *models.VideoSession) error {
	return r.db.Create(videoSession).Error
}

// FindByID IDでビデオセッションを取得
func (r *videoSessionRepository) FindByID(id uint) (*models.VideoSession, error) {
	var videoSession models.VideoSession
	err := r.db.Where("id = ?", id).First(&videoSession).Error
	if err != nil {
		return nil, err
	}
	return &videoSession, nil
}

// FindByAppointmentID 予約IDでビデオセッション一覧を取得
func (r *videoSessionRepository) FindByAppointmentID(appointmentID uint) ([]models.VideoSession, error) {
	var videoSessions []models.VideoSession
	err := r.db.Where("appointment_id = ?", appointmentID).Order("created_at DESC").Find(&videoSessions).Error
	return videoSessions, err
}

// FindActiveByAppointment 予約IDでアクティブなビデオセッションを取得
func (r *videoSessionRepository) FindActiveByAppointment(appointmentID uint) (*models.VideoSession, error) {
	var videoSession models.VideoSession
	err := r.db.Where("appointment_id = ? AND started_at IS NOT NULL AND ended_at IS NULL", appointmentID).
		Order("created_at DESC").First(&videoSession).Error
	if err != nil {
		return nil, err
	}
	return &videoSession, nil
}

// FindByRoomID ルームIDでビデオセッションを取得
func (r *videoSessionRepository) FindByRoomID(roomID string) (*models.VideoSession, error) {
	var videoSession models.VideoSession
	err := r.db.Where("room_id = ?", roomID).First(&videoSession).Error
	if err != nil {
		return nil, err
	}
	return &videoSession, nil
}

// Update ビデオセッションの更新
func (r *videoSessionRepository) Update(videoSession *models.VideoSession) error {
	return r.db.Save(videoSession).Error
}

// UpdateStartedAt 開始時刻の更新
func (r *videoSessionRepository) UpdateStartedAt(sessionID uint, startedAt *time.Time) error {
	return r.db.Model(&models.VideoSession{}).Where("id = ?", sessionID).Update("started_at", startedAt).Error
}

// UpdateEndedAt 終了時刻の更新
func (r *videoSessionRepository) UpdateEndedAt(sessionID uint, endedAt *time.Time) error {
	return r.db.Model(&models.VideoSession{}).Where("id = ?", sessionID).Update("ended_at", endedAt).Error
}

// Delete ビデオセッションの削除
func (r *videoSessionRepository) Delete(id uint) error {
	return r.db.Delete(&models.VideoSession{}, id).Error
}

// LoadRelations 関連データの読み込み
func (r *videoSessionRepository) LoadRelations(videoSession *models.VideoSession) error {
	return r.db.Preload("Appointment").First(videoSession, videoSession.ID).Error
}

// FindRecentSessions 最近のビデオセッションを取得
func (r *videoSessionRepository) FindRecentSessions(limit int) ([]models.VideoSession, error) {
	var videoSessions []models.VideoSession
	err := r.db.Order("created_at DESC").Limit(limit).Find(&videoSessions).Error
	return videoSessions, err
}

// FindSessionsByDateRange 日付範囲でビデオセッションを取得
func (r *videoSessionRepository) FindSessionsByDateRange(startDate, endDate time.Time) ([]models.VideoSession, error) {
	var videoSessions []models.VideoSession
	err := r.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").Find(&videoSessions).Error
	return videoSessions, err
}

// GetSessionStats ビデオセッションの統計情報を取得
func (r *videoSessionRepository) GetSessionStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 総件数
	var totalCount int64
	if err := r.db.Model(&models.VideoSession{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount

	// アクティブセッション数
	var activeCount int64
	if err := r.db.Model(&models.VideoSession{}).
		Where("started_at IS NOT NULL AND ended_at IS NULL").
		Count(&activeCount).Error; err != nil {
		return nil, err
	}
	stats["active_count"] = activeCount

	// 完了セッション数
	var completedCount int64
	if err := r.db.Model(&models.VideoSession{}).
		Where("ended_at IS NOT NULL").
		Count(&completedCount).Error; err != nil {
		return nil, err
	}
	stats["completed_count"] = completedCount

	// 平均セッション時間（完了済みのみ）
	var avgDuration float64
	if completedCount > 0 {
		err := r.db.Raw(`
			SELECT AVG(EXTRACT(EPOCH FROM (ended_at - started_at))) 
			FROM video_sessions 
			WHERE started_at IS NOT NULL AND ended_at IS NOT NULL
		`).Scan(&avgDuration).Error
		if err != nil {
			return nil, err
		}
		stats["avg_duration_seconds"] = avgDuration
	}

	return stats, nil
}

// FindSessionsByUser ユーザーに関連するビデオセッションを取得
func (r *videoSessionRepository) FindSessionsByUser(userID uint, limit, offset int) ([]models.VideoSession, error) {
	var videoSessions []models.VideoSession
	err := r.db.Joins("JOIN appointments ON video_sessions.appointment_id = appointments.id").
		Where("appointments.patient_id = ? OR appointments.doctor_id = ?", userID, userID).
		Order("video_sessions.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&videoSessions).Error
	return videoSessions, err
}

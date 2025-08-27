package repositories

import (
	"time"

	"gorm.io/gorm"
	"online_medical_consultation_app/backend/internal/models"
)

type MessageRepository interface {
	Create(message *models.Message) error
	FindByID(id uint) (*models.Message, error)
	FindByAppointmentID(appointmentID uint, limit, offset int) ([]models.Message, error)
	Update(message *models.Message) error
	Delete(id uint) error
	LoadRelations(message *models.Message) error
	MarkAsRead(appointmentID, userID uint) error
	GetUnreadCount(appointmentID, userID uint) (int, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{
		db: db,
	}
}

// Create メッセージの作成
func (r *messageRepository) Create(message *models.Message) error {
	return r.db.Create(message).Error
}

// FindByID IDでメッセージを取得
func (r *messageRepository) FindByID(id uint) (*models.Message, error) {
	var message models.Message
	err := r.db.Where("id = ?", id).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// FindByAppointmentID 予約IDでメッセージ一覧を取得
func (r *messageRepository) FindByAppointmentID(appointmentID uint, limit, offset int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Where("appointment_id = ?", appointmentID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

// FindUnreadByAppointmentID 予約IDで未読メッセージ一覧を取得
func (r *messageRepository) FindUnreadByAppointmentID(appointmentID, userID uint) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Where("appointment_id = ? AND sender_user_id != ? AND read_at IS NULL", 
		appointmentID, userID).Order("created_at ASC").Find(&messages).Error
	return messages, err
}

// Update メッセージの更新
func (r *messageRepository) Update(message *models.Message) error {
	return r.db.Save(message).Error
}

// Delete メッセージの削除
func (r *messageRepository) Delete(id uint) error {
	return r.db.Delete(&models.Message{}, id).Error
}

// LoadRelations 関連データの読み込み
func (r *messageRepository) LoadRelations(message *models.Message) error {
	return r.db.Preload("Appointment").Preload("Sender").First(message, message.ID).Error
}

// MarkAsRead メッセージを既読にする
func (r *messageRepository) MarkAsRead(appointmentID, userID uint) error {
	now := time.Now()
	return r.db.Model(&models.Message{}).
		Where("appointment_id = ? AND sender_user_id != ? AND read_at IS NULL", 
			appointmentID, userID).
		Update("read_at", now).Error
}

// GetUnreadCount 未読メッセージ数を取得
func (r *messageRepository) GetUnreadCount(appointmentID, userID uint) (int, error) {
	var count int64
	err := r.db.Model(&models.Message{}).
		Where("appointment_id = ? AND sender_user_id != ? AND read_at IS NULL", 
			appointmentID, userID).
		Count(&count).Error
	return int(count), err
}

// FindRecentMessages 最近のメッセージを取得（通知用）
func (r *messageRepository) FindRecentMessages(userID uint, limit int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Joins("JOIN appointments ON messages.appointment_id = appointments.id").
		Where("(appointments.patient_id = ? OR appointments.doctor_id = ?) AND messages.sender_user_id != ?", 
			userID, userID, userID).
		Order("messages.created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

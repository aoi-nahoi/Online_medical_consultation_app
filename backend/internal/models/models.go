package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// User ユーザー基本情報
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	Role         string         `gorm:"not null;check:role IN ('patient','doctor')" json:"role"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// リレーション
	PatientProfile *PatientProfile `json:"patient_profile,omitempty"`
	DoctorProfile  *DoctorProfile  `json:"doctor_profile,omitempty"`
}

// PatientProfile 患者プロフィール
type PatientProfile struct {
	UserID    uint           `gorm:"primaryKey" json:"user_id"`
	Name      string         `gorm:"not null" json:"name"`
	Birthdate *time.Time     `json:"birthdate"`
	Phone     string         `json:"phone"`
	Address   string         `json:"address"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// リレーション
	User User `json:"user"`
}

// DoctorProfile 医師プロフィール
type DoctorProfile struct {
	UserID        uint           `gorm:"primaryKey" json:"user_id"`
	Name          string         `gorm:"not null" json:"name"`
	Specialty     string         `json:"specialty"`
	LicenseNumber string         `json:"license_number"`
	Bio           string         `json:"bio"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// リレーション
	User User `json:"user"`
}

// AvailabilitySlot 診療可能枠
type AvailabilitySlot struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	DoctorID    uint           `gorm:"not null" json:"doctor_id"`
	StartTime   time.Time      `gorm:"not null" json:"start_time"`
	EndTime     time.Time      `gorm:"not null" json:"end_time"`
	Status      string         `gorm:"not null;default:'open';check:status IN ('open','blocked')" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// リレーション
	Doctor      User           `json:"doctor"`
	Appointment *Appointment   `json:"appointment,omitempty"`
}

// MarshalJSON カスタムJSONマーシャリング
func (s AvailabilitySlot) MarshalJSON() ([]byte, error) {
	type Alias AvailabilitySlot
	return json.Marshal(&struct {
		*Alias
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		Alias:     (*Alias)(&s),
		StartTime: s.StartTime.Format(time.RFC3339),
		EndTime:   s.EndTime.Format(time.RFC3339),
		CreatedAt: s.CreatedAt.Format(time.RFC3339),
		UpdatedAt: s.UpdatedAt.Format(time.RFC3339),
	})
}

// Appointment 予約
type Appointment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PatientID uint           `gorm:"not null" json:"patient_id"`
	DoctorID  uint           `gorm:"not null" json:"doctor_id"`
	SlotID    *uint          `json:"slot_id"`
	Status    string         `gorm:"not null;default:'pending';check:status IN ('pending','confirmed','cancelled','completed')" json:"status"`
	Notes     string         `json:"notes"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// リレーション
	Patient      User           `json:"patient"`
	Doctor       User           `json:"doctor"`
	Slot         *AvailabilitySlot `json:"slot,omitempty"`
	Messages     []Message      `json:"messages,omitempty"`
	Prescriptions []Prescription `json:"prescriptions,omitempty"`
	VideoSessions []VideoSession `json:"video_sessions,omitempty"`
}

// Message チャットメッセージ
type Message struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	AppointmentID  uint           `gorm:"not null" json:"appointment_id"`
	SenderUserID   uint           `gorm:"not null" json:"sender_user_id"`
	Body           string         `json:"body"`
	AttachmentURL  *string        `json:"attachment_url"`
	ReadAt         *time.Time     `json:"read_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// リレーション
	Appointment Appointment `json:"appointment"`
	Sender      User        `json:"sender"`
}

// VideoSession ビデオセッション
type VideoSession struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	AppointmentID uint           `gorm:"not null" json:"appointment_id"`
	RoomID        string         `gorm:"not null" json:"room_id"`
	StartedAt     *time.Time     `json:"started_at"`
	EndedAt       *time.Time     `json:"ended_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// リレーション
	Appointment Appointment `json:"appointment"`
}

// Prescription 処方
type Prescription struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	AppointmentID      uint           `gorm:"not null" json:"appointment_id"`
	ItemsJSON          string         `gorm:"not null" json:"items_json"` // JSON文字列
	Notes              string         `json:"notes"`
	CreatedByDoctorID  uint           `gorm:"not null" json:"created_by_doctor_id"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	// リレーション
	Appointment     Appointment `json:"appointment"`
	CreatedByDoctor User        `json:"created_by_doctor"`
}

// AuditLog 監査ログ
type AuditLog struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    *uint          `json:"user_id"`
	Action    string         `gorm:"not null" json:"action"`
	Entity    string         `gorm:"not null" json:"entity"`
	EntityID  string         `gorm:"not null" json:"entity_id"`
	MetaJSON  string         `json:"meta_json"` // JSON文字列
	At        time.Time      `gorm:"not null;default:now()" json:"at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// リレーション
	User *User `json:"user,omitempty"`
}

// TableName テーブル名の指定
func (User) TableName() string {
	return "users"
}

func (PatientProfile) TableName() string {
	return "patient_profiles"
}

func (DoctorProfile) TableName() string {
	return "doctor_profiles"
}

func (AvailabilitySlot) TableName() string {
	return "availability_slots"
}

func (Appointment) TableName() string {
	return "appointments"
}

func (Message) TableName() string {
	return "messages"
}

func (VideoSession) TableName() string {
	return "video_sessions"
}

func (Prescription) TableName() string {
	return "prescriptions"
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

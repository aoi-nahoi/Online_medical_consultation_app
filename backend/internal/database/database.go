package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"online_medical_consultation_app/backend/internal/models"
)

var db *gorm.DB

// GetDB returns the global database instance
func GetDB() *gorm.DB {
	return db
}

// SetDB sets the global database instance
func SetDB(database *gorm.DB) {
	db = database
}

func Connect(databaseURL string) (*gorm.DB, error) {
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 接続テスト
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully")
	return db, nil
}

func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// テーブルの自動作成
	if err := db.AutoMigrate(
		&models.User{},
		&models.PatientProfile{},
		&models.DoctorProfile{},
		&models.AvailabilitySlot{},
		&models.Appointment{},
		&models.Message{},
		&models.VideoSession{},
		&models.Prescription{},
		&models.AuditLog{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// インデックスの作成
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// シードデータの作成
	if err := seedData(db); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func createIndexes(db *gorm.DB) error {
	// 予約の重複防止インデックス
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uniq_slot_confirmed 
		ON appointments(slot_id) 
		WHERE status IN ('pending','confirmed')
	`).Error; err != nil {
		return err
	}

	// その他のインデックス
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_appointments_patient_id ON appointments(patient_id);
		CREATE INDEX IF NOT EXISTS idx_appointments_doctor_id ON appointments(doctor_id);
		CREATE INDEX IF NOT EXISTS idx_messages_appointment_id ON messages(appointment_id);
		CREATE INDEX IF NOT EXISTS idx_slots_doctor_id ON availability_slots(doctor_id);
		CREATE INDEX IF NOT EXISTS idx_slots_start_time ON availability_slots(start_time_utc);
	`).Error; err != nil {
		return err
	}

	return nil
}

func seedData(db *gorm.DB) error {
	// 既存データがあるかチェック
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("Database already has data, skipping seed")
		return nil
	}

	log.Println("Creating seed data...")

	// 医師アカウントの作成
	doctor := &models.User{
		Email:        "doctor1@example.com",
		PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "pass"
		Role:         "doctor",
	}

	if err := db.Create(doctor).Error; err != nil {
		return err
	}

	doctorProfile := &models.DoctorProfile{
		UserID:        doctor.ID,
		Name:          "田中 医師",
		Specialty:     "内科",
		LicenseNumber: "123456",
		Bio:           "内科専門医として20年の経験があります。",
	}

	if err := db.Create(doctorProfile).Error; err != nil {
		return err
	}

	// 患者アカウントの作成
	patient := &models.User{
		Email:        "patient1@example.com",
		PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "pass"
		Role:         "patient",
	}

	if err := db.Create(patient).Error; err != nil {
		return err
	}

	// 日付文字列をtime.Timeに変換
	birthdate, _ := time.Parse("2006-01-02", "1985-03-15")
	
	patientProfile := &models.PatientProfile{
		UserID:    patient.ID,
		Name:      "佐藤 患者",
		Birthdate: &birthdate,
		Phone:     "090-1234-5678",
		Address:   "東京都渋谷区...",
	}

	if err := db.Create(patientProfile).Error; err != nil {
		return err
	}

	// 診療枠の作成（直近1週間）
	// ここで診療枠を作成するロジックを追加

	log.Println("Seed data created successfully")
	return nil
}

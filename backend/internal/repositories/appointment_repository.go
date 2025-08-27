package repositories

import (
	"time"

	"gorm.io/gorm"
	"online_medical_consultation_app/backend/internal/models"
)

type AppointmentRepository interface {
	Create(appointment *models.Appointment) error
	FindByID(id uint) (*models.Appointment, error)
	FindByPatientID(patientID uint) ([]models.Appointment, error)
	FindByDoctorID(doctorID uint) ([]models.Appointment, error)
	FindByDoctorAndTimeRange(doctorID uint, startTime, endTime time.Time) ([]models.Appointment, error)
	Update(appointment *models.Appointment) error
	Delete(id uint) error
	LoadRelations(appointment *models.Appointment) error
	FindPendingByDoctor(doctorID uint) ([]models.Appointment, error)
	FindConfirmedByDoctor(doctorID uint) ([]models.Appointment, error)
	FindUpcomingByPatient(patientID uint) ([]models.Appointment, error)
	FindCompletedByPatient(patientID uint) ([]models.Appointment, error)
}

type appointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepository{
		db: db,
	}
}

// Create 予約の作成
func (r *appointmentRepository) Create(appointment *models.Appointment) error {
	return r.db.Create(appointment).Error
}

// FindByID IDで予約を取得
func (r *appointmentRepository) FindByID(id uint) (*models.Appointment, error) {
	var appointment models.Appointment
	err := r.db.Where("id = ?", id).First(&appointment).Error
	if err != nil {
		return nil, err
	}
	return &appointment, nil
}

// FindByPatientID 患者IDで予約一覧を取得
func (r *appointmentRepository) FindByPatientID(patientID uint) ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Where("patient_id = ?", patientID).Order("created_at DESC").Find(&appointments).Error
	return appointments, err
}

// FindByDoctorID 医師IDで予約一覧を取得
func (r *appointmentRepository) FindByDoctorID(doctorID uint) ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Where("doctor_id = ?", doctorID).Order("created_at DESC").Find(&appointments).Error
	return appointments, err
}

// FindByDoctorAndTimeRange 医師IDと時間範囲で予約を取得
func (r *appointmentRepository) FindByDoctorAndTimeRange(doctorID uint, startTime, endTime time.Time) ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Where("doctor_id = ? AND ((start_time <= ? AND end_time >= ?) OR (start_time <= ? AND end_time >= ?) OR (start_time >= ? AND end_time <= ?))",
		doctorID, startTime, startTime, endTime, endTime, startTime, endTime).Find(&appointments).Error
	return appointments, err
}

// Update 予約の更新
func (r *appointmentRepository) Update(appointment *models.Appointment) error {
	return r.db.Save(appointment).Error
}

// Delete 予約の削除
func (r *appointmentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Appointment{}, id).Error
}

// LoadRelations 関連データの読み込み
func (r *appointmentRepository) LoadRelations(appointment *models.Appointment) error {
	return r.db.Preload("Patient").Preload("Doctor").Preload("Slot").Preload("Messages").Preload("Prescriptions").Preload("VideoSessions").First(appointment, appointment.ID).Error
}

// FindPendingByDoctor 医師の保留中予約を取得
func (r *appointmentRepository) FindPendingByDoctor(doctorID uint) ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Where("doctor_id = ? AND status = ?", doctorID, "pending").Order("created_at ASC").Find(&appointments).Error
	return appointments, err
}

// FindConfirmedByDoctor 医師の確定済み予約を取得
func (r *appointmentRepository) FindConfirmedByDoctor(doctorID uint) ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Where("doctor_id = ? AND status = ?", doctorID, "confirmed").Order("start_time ASC").Find(&appointments).Error
	return appointments, err
}

// FindUpcomingByPatient 患者の今後の予約を取得
func (r *appointmentRepository) FindUpcomingByPatient(patientID uint) ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Where("patient_id = ? AND status IN (?, ?) AND start_time > ?", 
		patientID, "pending", "confirmed", time.Now()).Order("start_time ASC").Find(&appointments).Error
	return appointments, err
}

// FindCompletedByPatient 患者の完了済み予約を取得
func (r *appointmentRepository) FindCompletedByPatient(patientID uint) ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Where("patient_id = ? AND status = ?", patientID, "completed").Order("start_time DESC").Find(&appointments).Error
	return appointments, err
}

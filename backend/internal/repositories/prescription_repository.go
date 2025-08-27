package repositories

import (
	"gorm.io/gorm"
	"online_medical_consultation_app/backend/internal/models"
)

type PrescriptionRepository interface {
	Create(prescription *models.Prescription) error
	FindByID(id uint) (*models.Prescription, error)
	FindByAppointmentID(appointmentID uint) ([]models.Prescription, error)
	Update(prescription *models.Prescription) error
	Delete(id uint) error
	LoadRelations(prescription *models.Prescription) error
}

type prescriptionRepository struct {
	db *gorm.DB
}

func NewPrescriptionRepository(db *gorm.DB) PrescriptionRepository {
	return &prescriptionRepository{
		db: db,
	}
}

// Create 処方の作成
func (r *prescriptionRepository) Create(prescription *models.Prescription) error {
	return r.db.Create(prescription).Error
}

// FindByID IDで処方を取得
func (r *prescriptionRepository) FindByID(id uint) (*models.Prescription, error) {
	var prescription models.Prescription
	err := r.db.Where("id = ?", id).First(&prescription).Error
	if err != nil {
		return nil, err
	}
	return &prescription, nil
}

// FindByAppointmentID 予約IDで処方一覧を取得
func (r *prescriptionRepository) FindByAppointmentID(appointmentID uint) ([]models.Prescription, error) {
	var prescriptions []models.Prescription
	err := r.db.Where("appointment_id = ?", appointmentID).Order("created_at DESC").Find(&prescriptions).Error
	return prescriptions, err
}

// FindByDoctorID 医師IDで処方一覧を取得
func (r *prescriptionRepository) FindByDoctorID(doctorID uint) ([]models.Prescription, error) {
	var prescriptions []models.Prescription
	err := r.db.Where("created_by_doctor_id = ?", doctorID).Order("created_at DESC").Find(&prescriptions).Error
	return prescriptions, err
}

// Update 処方の更新
func (r *prescriptionRepository) Update(prescription *models.Prescription) error {
	return r.db.Save(prescription).Error
}

// Delete 処方の削除
func (r *prescriptionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Prescription{}, id).Error
}

// LoadRelations 関連データの読み込み
func (r *prescriptionRepository) LoadRelations(prescription *models.Prescription) error {
	return r.db.Preload("Appointment").Preload("CreatedByDoctor").First(prescription, prescription.ID).Error
}

// FindRecentByPatient 患者の最近の処方を取得
func (r *prescriptionRepository) FindRecentByPatient(patientID uint, limit int) ([]models.Prescription, error) {
	var prescriptions []models.Prescription
	err := r.db.Joins("JOIN appointments ON prescriptions.appointment_id = appointments.id").
		Where("appointments.patient_id = ?", patientID).
		Order("prescriptions.created_at DESC").
		Limit(limit).
		Find(&prescriptions).Error
	return prescriptions, err
}

// FindByDateRange 日付範囲で処方を取得
func (r *prescriptionRepository) FindByDateRange(doctorID uint, startDate, endDate string) ([]models.Prescription, error) {
	var prescriptions []models.Prescription
	err := r.db.Where("created_by_doctor_id = ? AND DATE(created_at) BETWEEN ? AND ?", 
		doctorID, startDate, endDate).Order("created_at DESC").Find(&prescriptions).Error
	return prescriptions, err
}

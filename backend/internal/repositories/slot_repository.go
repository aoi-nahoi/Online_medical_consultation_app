package repositories

import (
	"time"
	"online_medical_consultation_app/backend/internal/models"
	"gorm.io/gorm"
)

type SlotRepository interface {
	Create(slot *models.AvailabilitySlot) error
	FindByID(id uint) (*models.AvailabilitySlot, error)
	FindByDoctorID(doctorID uint) ([]models.AvailabilitySlot, error)
	FindAvailableByDoctorIDAndDate(doctorID uint, startDate, endDate time.Time) ([]models.AvailabilitySlot, error)
	Update(slot *models.AvailabilitySlot) error
	Delete(id uint) error
}

type slotRepository struct {
	db *gorm.DB
}

func NewSlotRepository(db *gorm.DB) SlotRepository {
	return &slotRepository{
		db: db,
	}
}

func (r *slotRepository) Create(slot *models.AvailabilitySlot) error {
	return r.db.Create(slot).Error
}

func (r *slotRepository) FindByID(id uint) (*models.AvailabilitySlot, error) {
	var slot models.AvailabilitySlot
	if err := r.db.First(&slot, id).Error; err != nil {
		return nil, err
	}
	return &slot, nil
}

func (r *slotRepository) FindByDoctorID(doctorID uint) ([]models.AvailabilitySlot, error) {
	var slots []models.AvailabilitySlot
	if err := r.db.Where("doctor_id = ?", doctorID).Find(&slots).Error; err != nil {
		return nil, err
	}
	return slots, nil
}

func (r *slotRepository) FindAvailableByDoctorIDAndDate(doctorID uint, startDate, endDate time.Time) ([]models.AvailabilitySlot, error) {
	var slots []models.AvailabilitySlot
	if err := r.db.Where("doctor_id = ? AND start_time >= ? AND start_time <= ? AND status = ?", 
		doctorID, startDate, endDate, "open").Find(&slots).Error; err != nil {
		return nil, err
	}
	return slots, nil
}

func (r *slotRepository) Update(slot *models.AvailabilitySlot) error {
	return r.db.Save(slot).Error
}

func (r *slotRepository) Delete(id uint) error {
	return r.db.Delete(&models.AvailabilitySlot{}, id).Error
}

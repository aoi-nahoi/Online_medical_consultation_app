package repositories

import (
	"online_medical_consultation_app/backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindDoctors() ([]models.DoctorProfile, error)
	CreatePatientProfile(profile *models.PatientProfile) error
	CreateDoctorProfile(profile *models.DoctorProfile) error
	FindPatientProfileByUserID(userID uint) (*models.PatientProfile, error)
	FindDoctorProfileByUserID(userID uint) (*models.DoctorProfile, error)
	UpdatePatientProfile(profile *models.PatientProfile) error
	UpdateDoctorProfile(profile *models.DoctorProfile) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindDoctors() ([]models.DoctorProfile, error) {
	var doctors []models.DoctorProfile
	if err := r.db.Preload("User").Find(&doctors).Error; err != nil {
		return nil, err
	}
	return doctors, nil
}

func (r *userRepository) CreatePatientProfile(profile *models.PatientProfile) error {
	return r.db.Create(profile).Error
}

func (r *userRepository) CreateDoctorProfile(profile *models.DoctorProfile) error {
	return r.db.Create(profile).Error
}

func (r *userRepository) FindPatientProfileByUserID(userID uint) (*models.PatientProfile, error) {
	var profile models.PatientProfile
	if err := r.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *userRepository) FindDoctorProfileByUserID(userID uint) (*models.DoctorProfile, error) {
	var profile models.DoctorProfile
	if err := r.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *userRepository) UpdatePatientProfile(profile *models.PatientProfile) error {
	return r.db.Save(profile).Error
}

func (r *userRepository) UpdateDoctorProfile(profile *models.DoctorProfile) error {
	return r.db.Save(profile).Error
}

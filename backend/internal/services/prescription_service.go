package services

import (
	"encoding/json"
	"errors"

	"online_medical_consultation_app/backend/internal/models"
	"online_medical_consultation_app/backend/internal/repositories"
)

type PrescriptionService struct {
	prescriptionRepo repositories.PrescriptionRepository
	appointmentRepo  repositories.AppointmentRepository
	userRepo         repositories.UserRepository
}

type PrescriptionItem struct {
	MedicationName string `json:"medication_name" binding:"required"`
	Dosage         string `json:"dosage" binding:"required"`
	Frequency      string `json:"frequency" binding:"required"`
	Duration       string `json:"duration" binding:"required"`
	Instructions   string `json:"instructions"`
}

type CreatePrescriptionRequest struct {
	AppointmentID      uint               `json:"appointment_id"`
	Items             []PrescriptionItem `json:"items" binding:"required,min=1"`
	Notes             string             `json:"notes"`
	CreatedByDoctorID uint               `json:"created_by_doctor_id"`
}

type UpdatePrescriptionRequest struct {
	PrescriptionID uint               `json:"prescription_id"`
	DoctorID       uint               `json:"doctor_id"`
	Items          []PrescriptionItem `json:"items" binding:"required,min=1"`
	Notes          string             `json:"notes"`
}

func NewPrescriptionService(prescriptionRepo repositories.PrescriptionRepository, appointmentRepo repositories.AppointmentRepository, userRepo repositories.UserRepository) *PrescriptionService {
	return &PrescriptionService{
		prescriptionRepo: prescriptionRepo,
		appointmentRepo:  appointmentRepo,
		userRepo:         userRepo,
	}
}

// CreatePrescription 処方の作成
func (s *PrescriptionService) CreatePrescription(req CreatePrescriptionRequest) (*models.Prescription, error) {
	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(req.AppointmentID)
	if err != nil || appointment == nil {
		return nil, errors.New("appointment not found")
	}

	// 医師の権限確認
	if appointment.DoctorID != req.CreatedByDoctorID {
		return nil, errors.New("unauthorized to create prescription for this appointment")
	}

	// 医師の存在確認
	doctor, err := s.userRepo.FindByID(req.CreatedByDoctorID)
	if err != nil || doctor == nil || doctor.Role != "doctor" {
		return nil, errors.New("doctor not found")
	}

	// 処方項目のJSON変換
	itemsJSON, err := json.Marshal(req.Items)
	if err != nil {
		return nil, errors.New("invalid prescription items format")
	}

	// 処方の作成
	prescription := &models.Prescription{
		AppointmentID:     req.AppointmentID,
		ItemsJSON:         string(itemsJSON),
		Notes:             req.Notes,
		CreatedByDoctorID: req.CreatedByDoctorID,
	}

	if err := s.prescriptionRepo.Create(prescription); err != nil {
		return nil, err
	}

	// 関連データの読み込み
	if err := s.prescriptionRepo.LoadRelations(prescription); err != nil {
		return nil, err
	}

	return prescription, nil
}

// GetPrescriptions 処方一覧の取得
func (s *PrescriptionService) GetPrescriptions(appointmentID, userID uint) ([]models.Prescription, error) {
	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(appointmentID)
	if err != nil || appointment == nil {
		return nil, errors.New("appointment not found")
	}

	// 権限確認（患者または医師のみ）
	if appointment.PatientID != userID && appointment.DoctorID != userID {
		return nil, errors.New("unauthorized to view prescriptions for this appointment")
	}

	// 処方一覧の取得
	prescriptions, err := s.prescriptionRepo.FindByAppointmentID(appointmentID)
	if err != nil {
		return nil, err
	}

	// 関連データの読み込み
	for i := range prescriptions {
		if err := s.prescriptionRepo.LoadRelations(&prescriptions[i]); err != nil {
			return nil, err
		}
	}

	return prescriptions, nil
}

// GetPrescriptionDetails 処方詳細の取得
func (s *PrescriptionService) GetPrescriptionDetails(prescriptionID, userID uint) (*models.Prescription, error) {
	// 処方の存在確認
	prescription, err := s.prescriptionRepo.FindByID(prescriptionID)
	if err != nil || prescription == nil {
		return nil, errors.New("prescription not found")
	}

	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(prescription.AppointmentID)
	if err != nil || appointment == nil {
		return nil, errors.New("appointment not found")
	}

	// 権限確認（患者または医師のみ）
	if appointment.PatientID != userID && appointment.DoctorID != userID {
		return nil, errors.New("unauthorized to view this prescription")
	}

	// 関連データの読み込み
	if err := s.prescriptionRepo.LoadRelations(prescription); err != nil {
		return nil, err
	}

	return prescription, nil
}

// UpdatePrescription 処方の更新
func (s *PrescriptionService) UpdatePrescription(req UpdatePrescriptionRequest) (*models.Prescription, error) {
	// 処方の存在確認
	prescription, err := s.prescriptionRepo.FindByID(req.PrescriptionID)
	if err != nil || prescription == nil {
		return nil, errors.New("prescription not found")
	}

	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(prescription.AppointmentID)
	if err != nil || appointment == nil {
		return nil, errors.New("appointment not found")
	}

	// 医師の権限確認
	if appointment.DoctorID != req.DoctorID {
		return nil, errors.New("unauthorized to update this prescription")
	}

	// 処方項目のJSON変換
	itemsJSON, err := json.Marshal(req.Items)
	if err != nil {
		return nil, errors.New("invalid prescription items format")
	}

	// 処方の更新
	prescription.ItemsJSON = string(itemsJSON)
	prescription.Notes = req.Notes

	if err := s.prescriptionRepo.Update(prescription); err != nil {
		return nil, err
	}

	// 関連データの読み込み
	if err := s.prescriptionRepo.LoadRelations(prescription); err != nil {
		return nil, err
	}

	return prescription, nil
}

// DeletePrescription 処方の削除
func (s *PrescriptionService) DeletePrescription(prescriptionID, userID uint) error {
	// 処方の存在確認
	prescription, err := s.prescriptionRepo.FindByID(prescriptionID)
	if err != nil || prescription == nil {
		return errors.New("prescription not found")
	}

	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(prescription.AppointmentID)
	if err != nil || appointment == nil {
		return errors.New("appointment not found")
	}

	// 医師の権限確認
	if appointment.DoctorID != userID {
		return errors.New("unauthorized to delete this prescription")
	}

	return s.prescriptionRepo.Delete(prescriptionID)
}

// GetPrescriptionItems 処方項目の取得（JSONから構造体に変換）
func (s *PrescriptionService) GetPrescriptionItems(prescription *models.Prescription) ([]PrescriptionItem, error) {
	var items []PrescriptionItem
	if err := json.Unmarshal([]byte(prescription.ItemsJSON), &items); err != nil {
		return nil, err
	}
	return items, nil
}

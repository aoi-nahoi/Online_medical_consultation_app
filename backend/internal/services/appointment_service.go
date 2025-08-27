package services

import (
	"errors"
	"time"

	"online_medical_consultation_app/backend/internal/models"
	"online_medical_consultation_app/backend/internal/repositories"
)

type AppointmentService struct {
	appointmentRepo repositories.AppointmentRepository
	slotRepo       repositories.SlotRepository
	userRepo       repositories.UserRepository
}

type CreateAppointmentRequest struct {
	PatientID uint      `json:"patient_id"`
	DoctorID  uint      `json:"doctor_id" binding:"required"`
	SlotID    *uint     `json:"slot_id"`
	Notes     string    `json:"notes"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
}

type UpdateAppointmentStatusRequest struct {
	AppointmentID uint   `json:"appointment_id"`
	DoctorID      uint   `json:"doctor_id"`
	Status        string `json:"status" binding:"required,oneof=pending confirmed cancelled completed"`
	Notes         string `json:"notes"`
}

func NewAppointmentService(appointmentRepo repositories.AppointmentRepository, slotRepo repositories.SlotRepository, userRepo repositories.UserRepository) *AppointmentService {
	return &AppointmentService{
		appointmentRepo: appointmentRepo,
		slotRepo:       slotRepo,
		userRepo:       userRepo,
	}
}

// CreateAppointment 予約の作成
func (s *AppointmentService) CreateAppointment(req CreateAppointmentRequest) (*models.Appointment, error) {
	// 医師の存在確認
	doctor, err := s.userRepo.FindByID(req.DoctorID)
	if err != nil || doctor == nil || doctor.Role != "doctor" {
		return nil, errors.New("doctor not found")
	}

	// 患者の存在確認
	patient, err := s.userRepo.FindByID(req.PatientID)
	if err != nil || patient == nil || patient.Role != "patient" {
		return nil, errors.New("patient not found")
	}

	// 時間の妥当性チェック
	if req.StartTime.Before(time.Now()) {
		return nil, errors.New("start time cannot be in the past")
	}

	if req.EndTime.Before(req.StartTime) {
		return nil, errors.New("end time must be after start time")
	}

	// 既存の予約との重複チェック
	existingAppointments, err := s.appointmentRepo.FindByDoctorAndTimeRange(req.DoctorID, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}

	for _, existing := range existingAppointments {
		if existing.Status != "cancelled" {
			return nil, errors.New("time slot is already booked")
		}
	}

	// 予約の作成
	appointment := &models.Appointment{
		PatientID: req.PatientID,
		DoctorID:  req.DoctorID,
		SlotID:    req.SlotID,
		Status:    "pending",
		Notes:     req.Notes,
	}

	if err := s.appointmentRepo.Create(appointment); err != nil {
		return nil, err
	}

	// 関連データの読み込み
	if err := s.appointmentRepo.LoadRelations(appointment); err != nil {
		return nil, err
	}

	return appointment, nil
}

// GetPatientAppointments 患者の予約一覧取得
func (s *AppointmentService) GetPatientAppointments(patientID uint) ([]models.Appointment, error) {
	appointments, err := s.appointmentRepo.FindByPatientID(patientID)
	if err != nil {
		return nil, err
	}

	// 関連データの読み込み
	for i := range appointments {
		if err := s.appointmentRepo.LoadRelations(&appointments[i]); err != nil {
			return nil, err
		}
	}

	return appointments, nil
}

// GetDoctorAppointments 医師の予約一覧取得
func (s *AppointmentService) GetDoctorAppointments(doctorID uint) ([]models.Appointment, error) {
	appointments, err := s.appointmentRepo.FindByDoctorID(doctorID)
	if err != nil {
		return nil, err
	}

	// 関連データの読み込み
	for i := range appointments {
		if err := s.appointmentRepo.LoadRelations(&appointments[i]); err != nil {
			return nil, err
		}
	}

	return appointments, nil
}

// UpdateAppointmentStatus 予約ステータスの更新
func (s *AppointmentService) UpdateAppointmentStatus(req UpdateAppointmentStatusRequest) (*models.Appointment, error) {
	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(req.AppointmentID)
	if err != nil || appointment == nil {
		return nil, errors.New("appointment not found")
	}

	// 医師の権限確認
	if appointment.DoctorID != req.DoctorID {
		return nil, errors.New("unauthorized to update this appointment")
	}

	// ステータスの更新
	appointment.Status = req.Status
	if req.Notes != "" {
		appointment.Notes = req.Notes
	}

	if err := s.appointmentRepo.Update(appointment); err != nil {
		return nil, err
	}

	// 関連データの読み込み
	if err := s.appointmentRepo.LoadRelations(appointment); err != nil {
		return nil, err
	}

	return appointment, nil
}

// CancelAppointment 予約のキャンセル
func (s *AppointmentService) CancelAppointment(appointmentID, userID uint) error {
	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(appointmentID)
	if err != nil || appointment == nil {
		return errors.New("appointment not found")
	}

	// 権限確認（患者または医師のみ）
	if appointment.PatientID != userID && appointment.DoctorID != userID {
		return errors.New("unauthorized to cancel this appointment")
	}

	// キャンセル可能なステータスかチェック
	if appointment.Status == "completed" || appointment.Status == "cancelled" {
		return errors.New("appointment cannot be cancelled")
	}

	// ステータスの更新
	appointment.Status = "cancelled"
	return s.appointmentRepo.Update(appointment)
}

// GetAppointmentDetails 予約詳細の取得
func (s *AppointmentService) GetAppointmentDetails(appointmentID, userID uint) (*models.Appointment, error) {
	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(appointmentID)
	if err != nil || appointment == nil {
		return nil, errors.New("appointment not found")
	}

	// 権限確認（患者または医師のみ）
	if appointment.PatientID != userID && appointment.DoctorID != userID {
		return nil, errors.New("unauthorized to view this appointment")
	}

	// 関連データの読み込み
	if err := s.appointmentRepo.LoadRelations(appointment); err != nil {
		return nil, err
	}

	return appointment, nil
}

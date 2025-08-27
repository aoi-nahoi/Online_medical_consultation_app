package services

import (
	"errors"
	"log"
	"time"

	"online_medical_consultation_app/backend/internal/models"
	"online_medical_consultation_app/backend/internal/repositories"
)

type SlotService struct {
	slotRepo repositories.SlotRepository
}

type CreateSlotRequest struct {
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	Notes     string `json:"notes"`
}

type UpdateSlotRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

func NewSlotService(slotRepo repositories.SlotRepository) *SlotService {
	return &SlotService{
		slotRepo: slotRepo,
	}
}

// CreateSlot 診療枠の作成
func (s *SlotService) CreateSlot(doctorID uint, req CreateSlotRequest) (*models.AvailabilitySlot, error) {
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, errors.New("invalid start time format")
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, errors.New("invalid end time format")
	}

	if startTime.Before(time.Now()) {
		return nil, errors.New("start time cannot be in the past")
	}

	if startTime.After(endTime) || startTime.Equal(endTime) {
		return nil, errors.New("start time must be before end time")
	}

	slot := &models.AvailabilitySlot{
		DoctorID:  doctorID,
		StartTime: startTime,
		EndTime:   endTime,
		Status:    "open",
	}

	if err := s.slotRepo.Create(slot); err != nil {
		return nil, err
	}

	return slot, nil
}

// GetSlotsByDoctorID 医師の診療枠一覧取得
func (s *SlotService) GetSlotsByDoctorID(doctorID uint) ([]models.AvailabilitySlot, error) {
	return s.slotRepo.FindByDoctorID(doctorID)
}

// UpdateSlot 診療枠の更新
func (s *SlotService) UpdateSlot(slotID, doctorID uint, req UpdateSlotRequest) (*models.AvailabilitySlot, error) {
	slot, err := s.slotRepo.FindByID(slotID)
	if err != nil {
		return nil, err
	}

	if slot.DoctorID != doctorID {
		return nil, errors.New("unauthorized to update this slot")
	}

	if req.Status != "" {
		if req.Status != "open" && req.Status != "blocked" {
			return nil, errors.New("invalid status")
		}
		slot.Status = req.Status
	}

	if req.Notes != "" {
		// 備考フィールドがある場合は更新
		// 現在のモデルには備考フィールドがないため、必要に応じて追加
	}

	if err := s.slotRepo.Update(slot); err != nil {
		return nil, err
	}

	return slot, nil
}

// DeleteSlot 診療枠の削除
func (s *SlotService) DeleteSlot(slotID, doctorID uint) error {
	slot, err := s.slotRepo.FindByID(slotID)
	if err != nil {
		return err
	}

	if slot.DoctorID != doctorID {
		return errors.New("unauthorized to delete this slot")
	}

	// 予約が入っている診療枠は削除できない
	if slot.Appointment != nil {
		return errors.New("cannot delete slot with existing appointment")
	}

	return s.slotRepo.Delete(slotID)
}

// GetAvailableSlots 利用可能な診療枠の取得（患者用）
func (s *SlotService) GetAvailableSlots(doctorID uint, date string) ([]models.AvailabilitySlot, error) {
	// 日付文字列をパース
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, errors.New("invalid date format")
	}

	// 指定日の開始と終了
	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	slots, err := s.slotRepo.FindAvailableByDoctorIDAndDate(doctorID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	// 現在時刻より後の診療枠のみを返す
	var availableSlots []models.AvailabilitySlot
	now := time.Now()
	for _, slot := range slots {
		if slot.StartTime.After(now) && slot.Status == "open" {
			availableSlots = append(availableSlots, slot)
		}
	}

	// デバッグ用ログ
	for i, slot := range availableSlots {
		log.Printf("Slot %d: StartTime=%v, EndTime=%v", i, slot.StartTime, slot.EndTime)
	}

	return availableSlots, nil
}

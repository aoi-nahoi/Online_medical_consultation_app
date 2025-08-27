package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"online_medical_consultation_app/backend/internal/models"
	"online_medical_consultation_app/backend/internal/repositories"
)

type ChatService struct {
	messageRepo      repositories.MessageRepository
	appointmentRepo  repositories.AppointmentRepository
	userRepo         repositories.UserRepository
	uploadPath       string
}

type SendMessageRequest struct {
	AppointmentID  uint   `json:"appointment_id"`
	SenderUserID   uint   `json:"sender_user_id"`
	Body           string `json:"body" binding:"required"`
	AttachmentURL  *string `json:"attachment_url,omitempty"`
}

func NewChatService(messageRepo repositories.MessageRepository, appointmentRepo repositories.AppointmentRepository, userRepo repositories.UserRepository) *ChatService {
	uploadPath := os.Getenv("UPLOAD_PATH")
	if uploadPath == "" {
		uploadPath = "./uploads"
	}

	// アップロードディレクトリの作成
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		fmt.Printf("Warning: Failed to create upload directory: %v\n", err)
	}

	return &ChatService{
		messageRepo:     messageRepo,
		appointmentRepo: appointmentRepo,
		userRepo:        userRepo,
		uploadPath:      uploadPath,
	}
}

// SendMessage メッセージの送信
func (s *ChatService) SendMessage(req SendMessageRequest) (*models.Message, error) {
	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(req.AppointmentID)
	if err != nil || appointment == nil {
		return nil, errors.New("appointment not found")
	}

	// 送信者の権限確認（患者または医師のみ）
	if appointment.PatientID != req.SenderUserID && appointment.DoctorID != req.SenderUserID {
		return nil, errors.New("unauthorized to send message to this appointment")
	}

	// メッセージの作成
	message := &models.Message{
		AppointmentID: req.AppointmentID,
		SenderUserID:  req.SenderUserID,
		Body:          req.Body,
		AttachmentURL: req.AttachmentURL,
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	// 関連データの読み込み
	if err := s.messageRepo.LoadRelations(message); err != nil {
		return nil, err
	}

	return message, nil
}

// GetMessages メッセージ一覧の取得
func (s *ChatService) GetMessages(appointmentID, userID uint, limit, offset int) ([]models.Message, error) {
	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(appointmentID)
	if err != nil || appointment == nil {
		return nil, errors.New("appointment not found")
	}

	// 権限確認（患者または医師のみ）
	if appointment.PatientID != userID && appointment.DoctorID != userID {
		return nil, errors.New("unauthorized to view messages for this appointment")
	}

	// メッセージの取得
	messages, err := s.messageRepo.FindByAppointmentID(appointmentID, limit, offset)
	if err != nil {
		return nil, err
	}

	// 関連データの読み込み
	for i := range messages {
		if err := s.messageRepo.LoadRelations(&messages[i]); err != nil {
			return nil, err
		}
	}

	return messages, nil
}

// UploadAttachment 添付ファイルのアップロード
func (s *ChatService) UploadAttachment(file *multipart.FileHeader, appointmentID, userID uint) (string, error) {
	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(appointmentID)
	if err != nil || appointment == nil {
		return "", errors.New("appointment not found")
	}

	// 権限確認（患者または医師のみ）
	if appointment.PatientID != userID && appointment.DoctorID != userID {
		return "", errors.New("unauthorized to upload attachment for this appointment")
	}

	// ファイル名の生成（重複回避）
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%d_%s", appointmentID, timestamp, filepath.Base(file.Filename))
	filePath := filepath.Join(s.uploadPath, filename)

	// ディレクトリの作成
	uploadDir := filepath.Dir(filePath)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()

	// ファイルのコピー
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file: %v", err)
	}

	// ファイルURLの生成
	fileURL := fmt.Sprintf("/uploads/%s", filename)
	return fileURL, nil
}

// MarkMessagesAsRead メッセージを既読にする
func (s *ChatService) MarkMessagesAsRead(appointmentID, userID uint) error {
	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(appointmentID)
	if err != nil || appointment == nil {
		return errors.New("appointment not found")
	}

	// 権限確認（患者または医師のみ）
	if appointment.PatientID != userID && appointment.DoctorID != userID {
		return errors.New("unauthorized to mark messages as read for this appointment")
	}

	// 未読メッセージを既読にする
	return s.messageRepo.MarkAsRead(appointmentID, userID)
}

// GetUnreadCount 未読メッセージ数の取得
func (s *ChatService) GetUnreadCount(appointmentID, userID uint) (int, error) {
	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(appointmentID)
	if err != nil || appointment == nil {
		return 0, errors.New("appointment not found")
	}

	// 権限確認（患者または医師のみ）
	if appointment.PatientID != userID && appointment.DoctorID != userID {
		return 0, errors.New("unauthorized to get unread count for this appointment")
	}

	// 未読メッセージ数の取得
	return s.messageRepo.GetUnreadCount(appointmentID, userID)
}

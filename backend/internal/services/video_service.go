package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"online_medical_consultation_app/backend/internal/models"
	"online_medical_consultation_app/backend/internal/repositories"
)

type VideoService struct {
	videoSessionRepo repositories.VideoSessionRepository
	appointmentRepo  repositories.AppointmentRepository
	userRepo         repositories.UserRepository
}

type CreateVideoSessionRequest struct {
	AppointmentID     uint   `json:"appointment_id"`
	CreatedByUserID   uint   `json:"created_by_user_id"`
	RoomName          string `json:"room_name"`
	MaxParticipants   int    `json:"max_participants"`
	RecordingEnabled  bool   `json:"recording_enabled"`
}

type WebRTCAnswerRequest struct {
	Answer string `json:"answer" binding:"required"`
}

type SignalingInfo struct {
	RoomID      string   `json:"room_id"`
	ICEServers  []string `json:"ice_servers"`
	RoomToken   string   `json:"room_token"`
	ExpiresAt   string   `json:"expires_at"`
}

func NewVideoService(videoSessionRepo repositories.VideoSessionRepository, appointmentRepo repositories.AppointmentRepository, userRepo repositories.UserRepository) *VideoService {
	return &VideoService{
		videoSessionRepo: videoSessionRepo,
		appointmentRepo:  appointmentRepo,
		userRepo:         userRepo,
	}
}

// CreateVideoSession ビデオセッションの作成
func (s *VideoService) CreateVideoSession(req *CreateVideoSessionRequest, userID uint) (*models.VideoSession, error) {
	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(req.AppointmentID)
	if err != nil || appointment == nil {
		return nil, errors.New("appointment not found")
	}

	// 権限確認（予約に関連する患者または医師のみ）
	if appointment.PatientID != userID && appointment.DoctorID != userID {
		return nil, errors.New("unauthorized to create video session for this appointment")
	}

	// 既存のアクティブセッションのチェック
	existingSession, err := s.videoSessionRepo.FindActiveByAppointment(req.AppointmentID)
	if err == nil && existingSession != nil {
		if existingSession.StartedAt != nil && existingSession.EndedAt == nil {
			return nil, errors.New("active video session already exists for this appointment")
		}
	}

	// ルームIDの生成
	roomID, err := s.generateRoomID()
	if err != nil {
		return nil, err
	}

	// ビデオセッションの作成
	videoSession := &models.VideoSession{
		AppointmentID: req.AppointmentID,
		RoomID:        roomID,
	}

	if err := s.videoSessionRepo.Create(videoSession); err != nil {
		return nil, err
	}

	// 関連データの読み込み
	if err := s.videoSessionRepo.LoadRelations(videoSession); err != nil {
		return nil, err
	}

	return videoSession, nil
}

// GetVideoSession ビデオセッション情報の取得
func (s *VideoService) GetVideoSession(sessionID uint) (*models.VideoSession, error) {
	session, err := s.videoSessionRepo.FindByID(sessionID)
	if err != nil || session == nil {
		return nil, errors.New("video session not found")
	}

	// 関連データの読み込み
	if err := s.videoSessionRepo.LoadRelations(session); err != nil {
		return nil, err
	}

	return session, nil
}

// ValidateSessionAccess セッションアクセスの権限確認
func (s *VideoService) ValidateSessionAccess(sessionID, userID uint) error {
	session, err := s.videoSessionRepo.FindByID(sessionID)
	if err != nil || session == nil {
		return errors.New("video session not found")
	}

	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(session.AppointmentID)
	if err != nil || appointment == nil {
		return errors.New("appointment not found")
	}

	// 権限確認（予約に関連する患者または医師のみ）
	if appointment.PatientID != userID && appointment.DoctorID != userID {
		return errors.New("unauthorized to access this video session")
	}

	return nil
}

// StartVideoSession ビデオセッションの開始
func (s *VideoService) StartVideoSession(sessionID, userID uint) error {
	// 権限確認
	if err := s.ValidateSessionAccess(sessionID, userID); err != nil {
		return err
	}

	// セッションの開始
	now := time.Now()
	return s.videoSessionRepo.UpdateStartedAt(sessionID, &now)
}

// EndVideoSession ビデオセッションの終了
func (s *VideoService) EndVideoSession(sessionID, userID uint) error {
	// 権限確認
	if err := s.ValidateSessionAccess(sessionID, userID); err != nil {
		return err
	}

	// セッションの終了
	now := time.Now()
	return s.videoSessionRepo.UpdateEndedAt(sessionID, &now)
}

// GetVideoSessionsByAppointment 予約に関連するビデオセッション一覧の取得
func (s *VideoService) GetVideoSessionsByAppointment(appointmentID, userID uint) ([]models.VideoSession, error) {
	// 予約の存在確認
	appointment, err := s.appointmentRepo.FindByID(appointmentID)
	if err != nil || appointment == nil {
		return nil, errors.New("appointment not found")
	}

	// 権限確認（予約に関連する患者または医師のみ）
	if appointment.PatientID != userID && appointment.DoctorID != userID {
		return nil, errors.New("unauthorized to view video sessions for this appointment")
	}

	// ビデオセッション一覧の取得
	sessions, err := s.videoSessionRepo.FindByAppointmentID(appointmentID)
	if err != nil {
		return nil, err
	}

	// 関連データの読み込み
	for i := range sessions {
		if err := s.videoSessionRepo.LoadRelations(&sessions[i]); err != nil {
			return nil, err
		}
	}

	return sessions, nil
}

// GetSignalingInfo WebRTC用のシグナリング情報を取得
func (s *VideoService) GetSignalingInfo(sessionID, userID uint) (*SignalingInfo, error) {
	// 権限確認
	if err := s.ValidateSessionAccess(sessionID, userID); err != nil {
		return nil, err
	}

	session, err := s.videoSessionRepo.FindByID(sessionID)
	if err != nil || session == nil {
		return nil, errors.New("video session not found")
	}

	// ルームトークンの生成
	roomToken, err := s.generateRoomToken(session.RoomID, userID)
	if err != nil {
		return nil, err
	}

	// ICEサーバーの設定（STUN/TURNサーバー）
	iceServers := []string{
		"stun:stun.l.google.com:19302",
		"stun:stun1.l.google.com:19302",
	}

	// 有効期限の設定（1時間）
	expiresAt := time.Now().Add(1 * time.Hour).Format(time.RFC3339)

	return &SignalingInfo{
		RoomID:     session.RoomID,
		ICEServers: iceServers,
		RoomToken:  roomToken,
		ExpiresAt:  expiresAt,
	}, nil
}

// GetWebRTCOffer WebRTCオファーの取得
func (s *VideoService) GetWebRTCOffer(sessionID, userID uint) (string, error) {
	// 権限確認
	if err := s.ValidateSessionAccess(sessionID, userID); err != nil {
		return "", err
	}

	// 実際の実装では、WebRTCのオファー生成ロジックが必要
	// ここでは簡易的な実装
	return "webrtc_offer_data", nil
}

// SetWebRTCAnswer WebRTCアンサーの設定
func (s *VideoService) SetWebRTCAnswer(sessionID, userID uint, req WebRTCAnswerRequest) error {
	// 権限確認
	if err := s.ValidateSessionAccess(sessionID, userID); err != nil {
		return err
	}

	// 実際の実装では、WebRTCのアンサー処理ロジックが必要
	// ここでは簡易的な実装
	return nil
}

// generateRoomID ユニークなルームIDを生成
func (s *VideoService) generateRoomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateRoomToken ルームトークンを生成
func (s *VideoService) generateRoomToken(roomID string, userID uint) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	return fmt.Sprintf("%s_%d_%s", roomID, userID, token), nil
}

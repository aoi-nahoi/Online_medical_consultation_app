package services

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"online_medical_consultation_app/backend/internal/models"
	"online_medical_consultation_app/backend/internal/repositories"
)

type AuditService struct {
	auditRepo repositories.AuditRepository
	userRepo  repositories.UserRepository
}

type AuditLogFilter struct {
	Entity    string `json:"entity"`
	EntityID  string `json:"entity_id"`
	Action    string `json:"action"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

func NewAuditService(auditRepo repositories.AuditRepository, userRepo repositories.UserRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
		userRepo:  userRepo,
	}
}

// CreateAuditLog 監査ログの作成
func (s *AuditService) CreateAuditLog(userID *uint, action, entity, entityID string, meta interface{}) error {
	// メタデータのJSON変換
	var metaJSON string
	if meta != nil {
		metaBytes, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal meta data: %v", err)
		}
		metaJSON = string(metaBytes)
	}

	// 監査ログの作成
	auditLog := &models.AuditLog{
		UserID:   userID,
		Action:   action,
		Entity:   entity,
		EntityID: entityID,
		MetaJSON: metaJSON,
		At:       time.Now(),
	}

	return s.auditRepo.Create(auditLog)
}

// GetAuditLogs 監査ログ一覧の取得
func (s *AuditService) GetAuditLogs(filter AuditLogFilter, userID uint) ([]models.AuditLog, error) {
	// 管理者権限のチェック（簡易版）
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// 管理者のみアクセス可能（実際の実装ではより詳細な権限チェックが必要）
	if user.Role != "admin" {
		return nil, errors.New("insufficient permissions")
	}

	// フィルタの適用
	query := s.auditRepo.GetDB()
	
	if filter.Entity != "" {
		query = query.Where("entity = ?", filter.Entity)
	}
	if filter.EntityID != "" {
		query = query.Where("entity_id = ?", filter.EntityID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.StartDate != "" {
		query = query.Where("DATE(at) >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("DATE(at) <= ?", filter.EndDate)
	}

	// 監査ログの取得
	logs, err := s.auditRepo.FindWithFilter(query, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}

	// 関連データの読み込み
	for i := range logs {
		if err := s.auditRepo.LoadRelations(&logs[i]); err != nil {
			return nil, err
		}
	}

	return logs, nil
}

// GetUserAuditLogs 特定ユーザーの監査ログ取得
func (s *AuditService) GetUserAuditLogs(targetUserID uint, limit, offset int, userID uint) ([]models.AuditLog, error) {
	// 権限チェック
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// 自分自身のログまたは管理者のみアクセス可能
	if userID != targetUserID && user.Role != "admin" {
		return nil, errors.New("insufficient permissions")
	}

	// ユーザーの監査ログを取得
	logs, err := s.auditRepo.FindByUserID(targetUserID, limit, offset)
	if err != nil {
		return nil, err
	}

	// 関連データの読み込み
	for i := range logs {
		if err := s.auditRepo.LoadRelations(&logs[i]); err != nil {
			return nil, err
		}
	}

	return logs, nil
}

// GetEntityAuditLogs 特定エンティティの監査ログ取得
func (s *AuditService) GetEntityAuditLogs(entity, entityID string, limit, offset int, userID uint) ([]models.AuditLog, error) {
	// 権限チェック
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// 管理者のみアクセス可能
	if user.Role != "admin" {
		return nil, errors.New("insufficient permissions")
	}

	// エンティティの監査ログを取得
	logs, err := s.auditRepo.FindByEntity(entity, entityID, limit, offset)
	if err != nil {
		return nil, err
	}

	// 関連データの読み込み
	for i := range logs {
		if err := s.auditRepo.LoadRelations(&logs[i]); err != nil {
			return nil, err
		}
	}

	return logs, nil
}

// ExportAuditLogs 監査ログのエクスポート
func (s *AuditService) ExportAuditLogs(filter AuditLogFilter, format string, userID uint) ([]byte, string, error) {
	// 管理者権限のチェック
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil, "", errors.New("user not found")
	}

	if user.Role != "admin" {
		return nil, "", errors.New("insufficient permissions")
	}

	// 監査ログの取得
	logs, err := s.GetAuditLogs(filter, userID)
	if err != nil {
		return nil, "", err
	}

	// ファイル名の生成
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("audit_logs_%s.%s", timestamp, format)

	var data []byte
	if format == "csv" {
		data, err = s.exportToCSV(logs)
	} else if format == "json" {
		data, err = json.MarshalIndent(logs, "", "  ")
	} else {
		return nil, "", errors.New("unsupported export format")
	}

	if err != nil {
		return nil, "", err
	}

	return data, filename, nil
}

// exportToCSV CSV形式でのエクスポート
func (s *AuditService) exportToCSV(logs []models.AuditLog) ([]byte, error) {
	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	// ヘッダーの書き込み
	headers := []string{"ID", "User ID", "Action", "Entity", "Entity ID", "Meta Data", "Timestamp", "Created At"}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	// データの書き込み
	for _, log := range logs {
		userID := ""
		if log.UserID != nil {
			userID = fmt.Sprintf("%d", *log.UserID)
		}

		row := []string{
			fmt.Sprintf("%d", log.ID),
			userID,
			log.Action,
			log.Entity,
			log.EntityID,
			log.MetaJSON,
			log.At.Format(time.RFC3339),
			log.CreatedAt.Format(time.RFC3339),
		}

		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return []byte(buffer.String()), nil
}

// LogUserAction ユーザーアクションのログ記録（ヘルパー関数）
func (s *AuditService) LogUserAction(userID uint, action, entity, entityID string, meta interface{}) {
	// 非同期でログを記録（エラーは無視）
	go func() {
		if err := s.CreateAuditLog(&userID, action, entity, entityID, meta); err != nil {
			// ログ記録の失敗はシステムに影響しないよう無視
			fmt.Printf("Warning: Failed to create audit log: %v\n", err)
		}
	}()
}

// LogSystemAction システムアクションのログ記録（ヘルパー関数）
func (s *AuditService) LogSystemAction(action, entity, entityID string, meta interface{}) {
	// 非同期でログを記録（エラーは無視）
	go func() {
		if err := s.CreateAuditLog(nil, action, entity, entityID, meta); err != nil {
			// ログ記録の失敗はシステムに影響しないよう無視
			fmt.Printf("Warning: Failed to create audit log: %v\n", err)
		}
	}()
}

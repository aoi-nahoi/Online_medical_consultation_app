package repositories

import (
	"time"
	"online_medical_consultation_app/backend/internal/models"
	"gorm.io/gorm"
)

// AuditStatistics 監査ログの統計情報
type AuditStatistics struct {
	TotalLogs     int64
	UserActions   int64
	SystemActions int64
	TopActions    []string
	TopUsers      []uint
}

type AuditRepository interface {
	Create(log *models.AuditLog) error
	FindByID(id uint) (*models.AuditLog, error)
	FindByUserID(userID uint, limit, offset int) ([]models.AuditLog, error)
	FindByEntity(entityType, entityID string, limit, offset int) ([]models.AuditLog, error)
	FindByDateRange(startDate, endDate time.Time, limit, offset int) ([]models.AuditLog, error)
	FindByAction(action string, limit, offset int) ([]models.AuditLog, error)
	GetStatistics(startDate, endDate time.Time) (*AuditStatistics, error)
	FindWithFilter(query *gorm.DB, limit, offset int) ([]models.AuditLog, error)
	LoadRelations(log *models.AuditLog) error
	GetDB() *gorm.DB
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{
		db: db,
	}
}

// Create 監査ログの作成
func (r *auditRepository) Create(auditLog *models.AuditLog) error {
	return r.db.Create(auditLog).Error
}

// FindByID IDで監査ログを取得
func (r *auditRepository) FindByID(id uint) (*models.AuditLog, error) {
	var auditLog models.AuditLog
	err := r.db.Where("id = ?", id).First(&auditLog).Error
	if err != nil {
		return nil, err
	}
	return &auditLog, nil
}

// FindByUserID ユーザーIDで監査ログ一覧を取得
func (r *auditRepository) FindByUserID(userID uint, limit, offset int) ([]models.AuditLog, error) {
	var auditLogs []models.AuditLog
	err := r.db.Where("user_id = ?", userID).
		Order("at DESC").
		Limit(limit).
		Offset(offset).
		Find(&auditLogs).Error
	return auditLogs, err
}

// FindByEntity エンティティで監査ログ一覧を取得
func (r *auditRepository) FindByEntity(entity, entityID string, limit, offset int) ([]models.AuditLog, error) {
	var auditLogs []models.AuditLog
	err := r.db.Where("entity = ? AND entity_id = ?", entity, entityID).
		Order("at DESC").
		Limit(limit).
		Offset(offset).
		Find(&auditLogs).Error
	return auditLogs, err
}

// FindWithFilter フィルタ付きで監査ログ一覧を取得
func (r *auditRepository) FindWithFilter(query *gorm.DB, limit, offset int) ([]models.AuditLog, error) {
	var auditLogs []models.AuditLog
	err := query.Order("at DESC").
		Limit(limit).
		Offset(offset).
		Find(&auditLogs).Error
	return auditLogs, err
}

// Update 監査ログの更新
func (r *auditRepository) Update(auditLog *models.AuditLog) error {
	return r.db.Save(auditLog).Error
}

// Delete 監査ログの削除
func (r *auditRepository) Delete(id uint) error {
	return r.db.Delete(&models.AuditLog{}, id).Error
}

// LoadRelations 関連データの読み込み
func (r *auditRepository) LoadRelations(auditLog *models.AuditLog) error {
	return r.db.Preload("User").First(auditLog, auditLog.ID).Error
}

// GetDB データベースインスタンスを取得
func (r *auditRepository) GetDB() *gorm.DB {
	return r.db
}

// FindByAction アクションで監査ログ一覧を取得
func (r *auditRepository) FindByAction(action string, limit, offset int) ([]models.AuditLog, error) {
	var auditLogs []models.AuditLog
	err := r.db.Where("action = ?", action).
		Order("at DESC").
		Limit(limit).
		Offset(offset).
		Find(&auditLogs).Error
	return auditLogs, err
}

// FindByDateRange 日付範囲で監査ログ一覧を取得
func (r *auditRepository) FindByDateRange(startDate, endDate time.Time, limit, offset int) ([]models.AuditLog, error) {
	var auditLogs []models.AuditLog
	err := r.db.Where("DATE(at) BETWEEN ? AND ?", startDate, endDate).
		Order("at DESC").
		Limit(limit).
		Offset(offset).
		Find(&auditLogs).Error
	return auditLogs, err
}

// GetAuditLogStats 監査ログの統計情報を取得
func (r *auditRepository) GetAuditLogStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 総件数
	var totalCount int64
	if err := r.db.Model(&models.AuditLog{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount

	// アクション別件数
	var actionStats []struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	if err := r.db.Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Group("action").
		Find(&actionStats).Error; err != nil {
		return nil, err
	}
	stats["action_stats"] = actionStats

	// エンティティ別件数
	var entityStats []struct {
		Entity string `json:"entity"`
		Count  int64  `json:"count"`
	}
	if err := r.db.Model(&models.AuditLog{}).
		Select("entity, COUNT(*) as count").
		Group("entity").
		Find(&entityStats).Error; err != nil {
		return nil, err
	}
	stats["entity_stats"] = entityStats

	return stats, nil
}

// GetStatistics 監査ログの統計情報を取得
func (r *auditRepository) GetStatistics(startDate, endDate time.Time) (*AuditStatistics, error) {
	stats := &AuditStatistics{}

	// 日付範囲でのクエリ
	query := r.db.Model(&models.AuditLog{}).Where("at BETWEEN ? AND ?", startDate, endDate)

	// 総件数
	if err := query.Count(&stats.TotalLogs).Error; err != nil {
		return nil, err
	}

	// ユーザーアクション数
	if err := query.Where("user_id IS NOT NULL").Count(&stats.UserActions).Error; err != nil {
		return nil, err
	}

	// システムアクション数
	if err := query.Where("user_id IS NULL").Count(&stats.SystemActions).Error; err != nil {
		return nil, err
	}

	// トップアクション
	var topActions []struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	if err := query.Select("action, COUNT(*) as count").
		Group("action").
		Order("count DESC").
		Limit(5).
		Find(&topActions).Error; err != nil {
		return nil, err
	}

	for _, action := range topActions {
		stats.TopActions = append(stats.TopActions, action.Action)
	}

	// トップユーザー
	var topUsers []struct {
		UserID uint  `json:"user_id"`
		Count  int64 `json:"count"`
	}
	if err := query.Select("user_id, COUNT(*) as count").
		Where("user_id IS NOT NULL").
		Group("user_id").
		Order("count DESC").
		Limit(5).
		Find(&topUsers).Error; err != nil {
		return nil, err
	}

	for _, user := range topUsers {
		stats.TopUsers = append(stats.TopUsers, user.UserID)
	}

	return stats, nil
}

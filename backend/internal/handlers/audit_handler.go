package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"online_medical_consultation_app/backend/internal/services"
)

type AuditHandler struct {
	auditService *services.AuditService
}

func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetAuditLogs 監査ログ一覧の取得（管理者用）
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// クエリパラメータの取得
	entity := c.Query("entity")
	entityID := c.Query("entity_id")
	action := c.Query("action")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	
	limit := 100 // デフォルト値
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	logs, err := h.auditService.GetAuditLogs(services.AuditLogFilter{
		Entity:    entity,
		EntityID:  entityID,
		Action:    action,
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     limit,
		Offset:    offset,
	}, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"audit_logs": logs})
}

// GetUserAuditLogs 特定ユーザーの監査ログ取得
func (h *AuditHandler) GetUserAuditLogs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	targetUserID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// クエリパラメータの取得
	limit := 50 // デフォルト値
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	logs, err := h.auditService.GetUserAuditLogs(uint(targetUserID), limit, offset, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"audit_logs": logs})
}

// GetEntityAuditLogs 特定エンティティの監査ログ取得
func (h *AuditHandler) GetEntityAuditLogs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	entity := c.Param("entity")
	entityID := c.Param("entityId")

	// クエリパラメータの取得
	limit := 50 // デフォルト値
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	logs, err := h.auditService.GetEntityAuditLogs(entity, entityID, limit, offset, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"audit_logs": logs})
}

// ExportAuditLogs 監査ログのエクスポート（管理者用）
func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// クエリパラメータの取得
	entity := c.Query("entity")
	entityID := c.Query("entity_id")
	action := c.Query("action")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	format := c.Query("format")
	if format == "" {
		format = "csv"
	}

	if format != "csv" && format != "json" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported export format"})
		return
	}

	data, filename, err := h.auditService.ExportAuditLogs(services.AuditLogFilter{
		Entity:    entity,
		EntityID:  entityID,
		Action:    action,
		StartDate: startDate,
		EndDate:   endDate,
	}, format, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ファイルのダウンロード
	c.Header("Content-Disposition", "attachment; filename="+filename)
	if format == "csv" {
		c.Header("Content-Type", "text/csv")
	} else {
		c.Header("Content-Type", "application/json")
	}
	c.Data(http.StatusOK, c.GetHeader("Content-Type"), data)
}

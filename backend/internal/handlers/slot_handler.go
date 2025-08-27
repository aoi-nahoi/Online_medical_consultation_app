package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"online_medical_consultation_app/backend/internal/services"
)

type SlotHandler struct {
	slotService *services.SlotService
}

func NewSlotHandler(slotService *services.SlotService) *SlotHandler {
	return &SlotHandler{
		slotService: slotService,
	}
}

// CreateSlot 診療枠の作成
func (h *SlotHandler) CreateSlot(c *gin.Context) {
	var req services.CreateSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ユーザーIDを取得（JWTから）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	slot, err := h.slotService.CreateSlot(userID.(uint), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Slot created successfully",
		"slot":    slot,
	})
}

// GetSlots 医師の診療枠一覧取得
func (h *SlotHandler) GetSlots(c *gin.Context) {
	// ユーザーIDを取得（JWTから）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	slots, err := h.slotService.GetSlotsByDoctorID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"slots": slots,
	})
}

// UpdateSlot 診療枠の更新
func (h *SlotHandler) UpdateSlot(c *gin.Context) {
	slotID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot ID"})
		return
	}

	var req services.UpdateSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ユーザーIDを取得（JWTから）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	slot, err := h.slotService.UpdateSlot(uint(slotID), userID.(uint), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Slot updated successfully",
		"slot":    slot,
	})
}

// DeleteSlot 診療枠の削除
func (h *SlotHandler) DeleteSlot(c *gin.Context) {
	slotID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot ID"})
		return
	}

	// ユーザーIDを取得（JWTから）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.slotService.DeleteSlot(uint(slotID), userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Slot deleted successfully",
	})
}

// GetAvailableSlots 利用可能な診療枠の取得（患者用）
func (h *SlotHandler) GetAvailableSlots(c *gin.Context) {
	log.Printf("GetAvailableSlots called with params: %+v", c.Params)
	log.Printf("Doctor ID param: %s", c.Param("doctorId"))
	
	doctorID, err := strconv.ParseUint(c.Param("doctorId"), 10, 32)
	if err != nil {
		log.Printf("Error parsing doctor ID '%s': %v", c.Param("doctorId"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor ID"})
		return
	}

	log.Printf("Parsed doctor ID: %d", doctorID)

	date := c.Query("date")
	log.Printf("Date query: %s", date)
	
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date parameter is required"})
		return
	}

	slots, err := h.slotService.GetAvailableSlots(uint(doctorID), date)
	if err != nil {
		log.Printf("Error getting available slots: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Found %d available slots", len(slots))
	c.JSON(http.StatusOK, gin.H{
		"slots": slots,
	})
}

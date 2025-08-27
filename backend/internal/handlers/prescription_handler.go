package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"online_medical_consultation_app/backend/internal/services"
)

type PrescriptionHandler struct {
	prescriptionService *services.PrescriptionService
}

func NewPrescriptionHandler(prescriptionService *services.PrescriptionService) *PrescriptionHandler {
	return &PrescriptionHandler{
		prescriptionService: prescriptionService,
	}
}

// CreatePrescription 処方の作成（医師用）
func (h *PrescriptionHandler) CreatePrescription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	appointmentID, err := strconv.ParseUint(c.Param("appointmentId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}

	var req services.CreatePrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.AppointmentID = uint(appointmentID)
	req.CreatedByDoctorID = userID.(uint)

	prescription, err := h.prescriptionService.CreatePrescription(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Prescription created successfully",
		"prescription": prescription,
	})
}

// GetPrescriptions 処方一覧の取得
func (h *PrescriptionHandler) GetPrescriptions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	appointmentID, err := strconv.ParseUint(c.Param("appointmentId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}

	prescriptions, err := h.prescriptionService.GetPrescriptions(uint(appointmentID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"prescriptions": prescriptions})
}

// GetPrescriptionDetails 処方詳細の取得
func (h *PrescriptionHandler) GetPrescriptionDetails(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	prescriptionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prescription ID"})
		return
	}

	prescription, err := h.prescriptionService.GetPrescriptionDetails(uint(prescriptionID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"prescription": prescription})
}

// UpdatePrescription 処方の更新（医師用）
func (h *PrescriptionHandler) UpdatePrescription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	prescriptionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prescription ID"})
		return
	}

	var req services.UpdatePrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.PrescriptionID = uint(prescriptionID)
	req.DoctorID = userID.(uint)

	prescription, err := h.prescriptionService.UpdatePrescription(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Prescription updated successfully",
		"prescription": prescription,
	})
}

// DeletePrescription 処方の削除（医師用）
func (h *PrescriptionHandler) DeletePrescription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	prescriptionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prescription ID"})
		return
	}

	if err := h.prescriptionService.DeletePrescription(uint(prescriptionID), userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prescription deleted successfully"})
}

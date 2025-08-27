package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"online_medical_consultation_app/backend/internal/config"
	"online_medical_consultation_app/backend/internal/database"
	"online_medical_consultation_app/backend/internal/handlers"
	"online_medical_consultation_app/backend/internal/middleware"
	"online_medical_consultation_app/backend/internal/repositories"
	"online_medical_consultation_app/backend/internal/services"
)

func main() {
	// 環境変数の読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 設定の読み込み
	cfg := config.Load()

	// データベース接続の初期化
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "host=localhost user=postgres password=postgres dbname=medical_consultation port=5432 sslmode=disable"
	}

	db, err := database.Connect(databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// グローバルデータベースインスタンスを設定
	database.SetDB(db)

	// データベースマイグレーションの実行
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run database migrations:", err)
	}

	// リポジトリの初期化
	userRepo := repositories.NewUserRepository(db)
	slotRepo := repositories.NewSlotRepository(db)
	appointmentRepo := repositories.NewAppointmentRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	prescriptionRepo := repositories.NewPrescriptionRepository(db)
	auditRepo := repositories.NewAuditRepository(db)
	videoSessionRepo := repositories.NewVideoSessionRepository(db)

	// サービスの初期化
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	slotService := services.NewSlotService(slotRepo)
	appointmentService := services.NewAppointmentService(appointmentRepo, slotRepo, userRepo)
	chatService := services.NewChatService(messageRepo, appointmentRepo, userRepo)
	prescriptionService := services.NewPrescriptionService(prescriptionRepo, appointmentRepo, userRepo)
	auditService := services.NewAuditService(auditRepo, userRepo)
	videoService := services.NewVideoService(videoSessionRepo, appointmentRepo, userRepo)

	// ハンドラーの初期化
	authHandler := handlers.NewAuthHandler(authService)
	slotHandler := handlers.NewSlotHandler(slotService)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService)
	chatHandler := handlers.NewChatHandler(chatService)
	prescriptionHandler := handlers.NewPrescriptionHandler(prescriptionService)
	auditHandler := handlers.NewAuditHandler(auditService)
	videoHandler := handlers.NewVideoHandler(videoService)

	// Ginルーターの設定
	router := gin.Default()

	// ミドルウェアの設定
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// APIルートの設定
	api := router.Group("/api/v1")
	{
		// 認証
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// 認証が必要なルート
		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.JWTSecret))
		{
			// 医師関連（/meルートを最初に定義）
			doctors := protected.Group("/doctors")
			{
				// /meルートを最初に定義（パラメータ付きルートより優先）
				doctors.GET("/me/slots", slotHandler.GetSlots)
				doctors.POST("/me/slots", slotHandler.CreateSlot)
				doctors.PUT("/me/slots/:id", slotHandler.UpdateSlot)
				doctors.DELETE("/me/slots/:id", slotHandler.DeleteSlot)
				doctors.GET("/me/profile", func(c *gin.Context) {
					userID, _ := c.Get("user_id")
					profile, err := userRepo.FindDoctorProfileByUserID(userID.(uint))
					if err != nil {
						c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
						return
					}
					c.JSON(http.StatusOK, gin.H{"profile": profile})
				})
				doctors.PUT("/me/profile", func(c *gin.Context) {
					log.Printf("PUT /doctors/me/profile called")
					userID, _ := c.Get("user_id")
					log.Printf("User ID: %v", userID)
					
					var req map[string]interface{}
					if err := c.ShouldBindJSON(&req); err != nil {
						log.Printf("JSON binding error: %v", err)
						c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
						return
					}
					
					log.Printf("Request body: %+v", req)
					
					profile, err := userRepo.FindDoctorProfileByUserID(userID.(uint))
					if err != nil {
						log.Printf("Profile not found error: %v", err)
						c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
						return
					}
					
					log.Printf("Current profile: %+v", profile)
					
					// プロフィールの更新
					if name, ok := req["name"].(string); ok {
						profile.Name = name
					}
					if specialty, ok := req["specialty"].(string); ok {
						profile.Specialty = specialty
					}
					if licenseNumber, ok := req["licenseNumber"].(string); ok {
						profile.LicenseNumber = licenseNumber
					}
					if bio, ok := req["bio"].(string); ok {
						profile.Bio = bio
					}
					
					log.Printf("Updated profile: %+v", profile)
					
					if err := userRepo.UpdateDoctorProfile(profile); err != nil {
						log.Printf("Update error: %v", err)
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
						return
					}
					
					log.Printf("Profile updated successfully")
					c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully", "profile": profile})
				})
			}

			// 患者関連
			patients := protected.Group("/patients")
			{
				patients.GET("/appointments", appointmentHandler.GetPatientAppointments)
				patients.POST("/appointments", appointmentHandler.CreateAppointment)
				patients.GET("/appointments/:id", appointmentHandler.GetAppointmentDetails)
				patients.PUT("/appointments/:id/cancel", appointmentHandler.CancelAppointment)
			}

			// 医師の予約取得エンドポイント
			protected.GET("/doctors/me/appointments", appointmentHandler.GetDoctorAppointments)
			protected.PUT("/doctors/me/appointments/:id/status", appointmentHandler.UpdateAppointmentStatus)

			// 医師一覧（患者用）
			protected.GET("/doctors", func(c *gin.Context) {
				// 実際の医師データを取得
				doctors, err := userRepo.FindDoctors()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch doctors"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"doctors": doctors})
			})

					// 利用可能な診療枠（患者用）
		protected.GET("/doctors/:doctorId/slots", slotHandler.GetAvailableSlots)

		// チャット機能
		chat := protected.Group("/appointments/:appointmentId/chat")
		{
			chat.GET("/messages", chatHandler.GetMessages)
			chat.POST("/messages", chatHandler.SendMessage)
			chat.POST("/attachments", chatHandler.UploadAttachment)
			chat.PUT("/read", chatHandler.MarkAsRead)
			chat.GET("/unread-count", chatHandler.GetUnreadCount)
		}

		// 処方管理
		prescriptions := protected.Group("/appointments/:appointmentId/prescriptions")
		{
			prescriptions.GET("", prescriptionHandler.GetPrescriptions)
			prescriptions.POST("", prescriptionHandler.CreatePrescription)
			prescriptions.GET("/:id", prescriptionHandler.GetPrescriptionDetails)
			prescriptions.PUT("/:id", prescriptionHandler.UpdatePrescription)
			prescriptions.DELETE("/:id", prescriptionHandler.DeletePrescription)
		}

		// ビデオ通話
		video := protected.Group("/appointments/:appointmentId/video")
		{
			video.POST("/sessions", videoHandler.CreateVideoSession)
			video.GET("/sessions", videoHandler.GetVideoSessionsByAppointment)
			video.GET("/sessions/:sessionId", videoHandler.GetVideoSession)
			video.POST("/sessions/:sessionId/join", videoHandler.JoinVideoSession)
			video.PUT("/sessions/:sessionId/start", videoHandler.StartVideoSession)
			video.PUT("/sessions/:sessionId/end", videoHandler.EndVideoSession)
			video.GET("/sessions/:sessionId/offer", videoHandler.GetWebRTCOffer)
			video.POST("/sessions/:sessionId/answer", videoHandler.SetWebRTCAnswer)
		}

		// 監査ログ（管理者用）
		audit := protected.Group("/audit")
		{
			audit.GET("/logs", auditHandler.GetAuditLogs)
			audit.GET("/users/:userId/logs", auditHandler.GetUserAuditLogs)
			audit.GET("/entities/:entity/:entityId/logs", auditHandler.GetEntityAuditLogs)
			audit.GET("/export", auditHandler.ExportAuditLogs)
		}
		}
	}



	// サーバー起動
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

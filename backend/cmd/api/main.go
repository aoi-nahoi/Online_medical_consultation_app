package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"telemed/internal/config"
	"telemed/internal/database"
	"telemed/internal/handlers"
	"telemed/internal/middleware"
	"telemed/internal/repositories"
	"telemed/internal/services"
	"telemed/internal/websocket"
)

func main() {
	// 環境変数の読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 設定の読み込み
	cfg := config.Load()

	// データベース接続
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// マイグレーション実行
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// リポジトリの初期化
	userRepo := repositories.NewUserRepository(db)
	appointmentRepo := repositories.NewAppointmentRepository(db)
	slotRepo := repositories.NewSlotRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	prescriptionRepo := repositories.NewPrescriptionRepository(db)
	auditRepo := repositories.NewAuditRepository(db)

	// サービスの初期化
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	appointmentService := services.NewAppointmentService(appointmentRepo, slotRepo, auditRepo)
	chatService := services.NewChatService(messageRepo, auditRepo)
	videoService := services.NewVideoService(auditRepo)

	// WebSocketハンドラーの初期化
	wsHandler := websocket.NewHandler(chatService, videoService)

	// ハンドラーの初期化
	authHandler := handlers.NewAuthHandler(authService)
	patientHandler := handlers.NewPatientHandler(appointmentService, chatService)
	doctorHandler := handlers.NewDoctorHandler(appointmentService, slotRepo, prescriptionService)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService)
	chatHandler := handlers.NewChatHandler(chatService)

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
			// 患者関連
			patients := protected.Group("/patients")
			{
				patients.GET("/me", patientHandler.GetProfile)
				patients.PUT("/me", patientHandler.UpdateProfile)
				patients.GET("/appointments", patientHandler.GetAppointments)
				patients.POST("/appointments", patientHandler.CreateAppointment)
				patients.GET("/appointments/:id", patientHandler.GetAppointment)
				patients.POST("/appointments/:id/cancel", patientHandler.CancelAppointment)
				patients.GET("/prescriptions", patientHandler.GetPrescriptions)
			}

			// 医師関連
			doctors := protected.Group("/doctors")
			{
				doctors.GET("/me", doctorHandler.GetProfile)
				doctors.PUT("/me", doctorHandler.UpdateProfile)
				doctors.GET("/me/slots", doctorHandler.GetSlots)
				doctors.POST("/me/slots", doctorHandler.CreateSlot)
				doctors.PUT("/me/slots/:id", doctorHandler.UpdateSlot)
				doctors.DELETE("/me/slots/:id", doctorHandler.DeleteSlot)
				doctors.GET("/me/appointments", doctorHandler.GetAppointments)
				doctors.POST("/appointments/:id/confirm", doctorHandler.ConfirmAppointment)
				doctors.POST("/appointments/:id/reject", doctorHandler.RejectAppointment)
				doctors.POST("/prescriptions", doctorHandler.CreatePrescription)
			}

			// 予約関連
			appointments := protected.Group("/appointments")
			{
				appointments.GET("/:id", appointmentHandler.GetAppointment)
			}

			// チャット関連
			chat := protected.Group("/chat")
			{
				chat.GET("/appointments/:id/messages", chatHandler.GetMessages)
				chat.POST("/appointments/:id/messages", chatHandler.SendMessage)
			}

			// ビデオ関連
			video := protected.Group("/video")
			{
				video.POST("/sessions", videoHandler.CreateSession)
				video.POST("/sessions/:roomId/token", videoHandler.GetToken)
			}

			// ファイルアップロード
			uploads := protected.Group("/uploads")
			{
				uploads.POST("", uploadHandler.UploadFile)
			}
		}
	}

	// WebSocketルート
	router.GET("/ws/appointments/:id", wsHandler.HandleChat)
	router.GET("/ws/video/:roomId", wsHandler.HandleVideo)

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

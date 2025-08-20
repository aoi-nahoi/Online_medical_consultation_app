package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"telemed/internal/models"
	"telemed/internal/repositories"
)

type AuthService struct {
	userRepo  repositories.UserRepository
	jwtSecret string
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=patient doctor"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	User       models.User `json:"user"`
}

type ProfileRequest struct {
	Name      *string    `json:"name,omitempty"`
	Birthdate *time.Time `json:"birthdate,omitempty"`
	Phone     *string    `json:"phone,omitempty"`
	Address   *string    `json:"address,omitempty"`
	Specialty *string    `json:"specialty,omitempty"`
	Bio       *string    `json:"bio,omitempty"`
}

func NewAuthService(userRepo repositories.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Register ユーザー登録
func (s *AuthService) Register(req RegisterRequest) (*models.User, error) {
	// 既存ユーザーのチェック
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user already exists")
	}

	// パスワードのハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// ユーザーの作成
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// プロフィールの作成
	if req.Role == "patient" {
		profile := &models.PatientProfile{
			UserID: user.ID,
			Name:   req.Name,
		}
		if err := s.userRepo.CreatePatientProfile(profile); err != nil {
			return nil, err
		}
	} else if req.Role == "doctor" {
		profile := &models.DoctorProfile{
			UserID: user.ID,
			Name:   req.Name,
		}
		if err := s.userRepo.CreateDoctorProfile(profile); err != nil {
			return nil, err
		}
	}

	return user, nil
}

// Login ユーザーログイン
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	// ユーザーの検索
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// パスワードの検証
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// JWTトークンの生成
	token, err := s.generateJWT(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken: token,
		User:        *user,
	}, nil
}

// generateJWT JWTトークンを生成
func (s *AuthService) generateJWT(userID uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ValidateToken JWTトークンの検証
func (s *AuthService) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return &claims, nil
}

// GetUserByID ユーザーIDでユーザーを取得
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

// UpdateProfile プロフィール更新
func (s *AuthService) UpdateProfile(userID uint, req ProfileRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if user.Role == "patient" {
		profile, err := s.userRepo.FindPatientProfileByUserID(userID)
		if err != nil {
			return err
		}

		if req.Name != nil {
			profile.Name = *req.Name
		}
		if req.Birthdate != nil {
			profile.Birthdate = req.Birthdate
		}
		if req.Phone != nil {
			profile.Phone = *req.Phone
		}
		if req.Address != nil {
			profile.Address = *req.Address
		}

		return s.userRepo.UpdatePatientProfile(profile)
	} else if user.Role == "doctor" {
		profile, err := s.userRepo.FindDoctorProfileByUserID(userID)
		if err != nil {
			return err
		}

		if req.Name != nil {
			profile.Name = *req.Name
		}
		if req.Specialty != nil {
			profile.Specialty = *req.Specialty
		}
		if req.Bio != nil {
			profile.Bio = *req.Bio
		}

		return s.userRepo.UpdateDoctorProfile(profile)
	}

	return errors.New("invalid user role")
}

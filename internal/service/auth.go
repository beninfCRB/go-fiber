package service

import (
	"backend/internal/dto"
	"backend/internal/helper"
	"backend/internal/models"
	"backend/internal/repository"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("account is inactive")
	ErrUserNotVerified    = errors.New("akun belum diverifikasi, silakan cek email Anda")
	ErrInvalidToken       = errors.New("invalid or expired refresh token")
	ErrTokenExpired       = errors.New("token telah kadaluwarsa")
)

type AuthService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtPrivateKey    *rsa.PrivateKey
	jwtExpiry        time.Duration
	refreshExpiry    time.Duration
	mailer           *helper.Mailer
	appURL           string
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	privateKey *rsa.PrivateKey,
	jwtExpiry time.Duration,
	refreshExpiry time.Duration,
	mailer *helper.Mailer,
	appURL string,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtPrivateKey:    privateKey,
		jwtExpiry:        jwtExpiry,
		refreshExpiry:    refreshExpiry,
		mailer:           mailer,
		appURL:           appURL,
	}
}

func (s *AuthService) GetAppURL() string {
	return s.appURL
}

func (s *AuthService) Register(name, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	vToken, _ := helper.GenerateRandomToken()
	user := &models.User{
		Name:              name,
		Email:             email,
		Password:          string(hash),
		IsVerified:        false,
		VerificationToken: vToken,
	}
	if err := s.userRepo.Create(user); err != nil {
		return err
	}
	if err := s.userRepo.AssignRole(user.ID, models.RoleUser); err != nil {
		return err
	}

	// Kirim email verifikasi
	verifyUrl := fmt.Sprintf("%s/auth/verify-email?token=%s", s.appURL, vToken)
	emailBody := fmt.Sprintf("Halo %s,<br><br>Terima kasih telah mendaftar di platform kami. Silakan klik link berikut untuk memverifikasi email Anda:<br><br><a href='%s' style='background-color:#4CAF50;color:white;padding:10px 20px;text-decoration:none;border-radius:5px;'>Verifikasi Email Saya</a>", name, verifyUrl)
	_ = s.mailer.SendEmail(email, "Verifikasi Akun Pendaftaran Anda", emailBody)

	return nil
}

func (s *AuthService) Login(email, password string) (*dto.TokenResponse, uuid.UUID, error) {
	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, uuid.Nil, ErrInvalidCredentials
	}
	if !u.IsActive {
		return nil, uuid.Nil, ErrUserInactive
	}
	if !u.IsVerified {
		return nil, uuid.Nil, ErrUserNotVerified
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		return nil, uuid.Nil, ErrInvalidCredentials
	}
	tokens, err := s.generateTokenPair(u)
	if err != nil {
		return nil, uuid.Nil, err
	}
	return tokens, u.ID, nil
}

func (s *AuthService) Refresh(rawRefreshToken string) (*dto.TokenResponse, error) {
	tokenHash := helper.HashToken(rawRefreshToken)

	rt, err := s.refreshTokenRepo.FindByToken(tokenHash)
	if err != nil {
		return nil, ErrInvalidToken
	}
	// Token rotation — revoke old token immediately
	if err := s.refreshTokenRepo.RevokeByToken(tokenHash); err != nil {
		return nil, err
	}

	u, err := s.userRepo.FindByID(rt.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !u.IsActive {
		return nil, ErrUserInactive
	}
	return s.generateTokenPair(u)
}

func (s *AuthService) Logout(rawRefreshToken string) error {
	return s.refreshTokenRepo.RevokeByToken(helper.HashToken(rawRefreshToken))
}

func (s *AuthService) LogoutAll(userID uuid.UUID) error {
	return s.refreshTokenRepo.RevokeAllByUser(userID)
}

// generateTokenPair builds an access JWT + refresh token pair.
// Roles are read from u.Roles (preloaded from DB) — NOT from a static column.
func (s *AuthService) generateTokenPair(u *models.User) (*dto.TokenResponse, error) {
	// Collect all role names from the preloaded relation
	roleNames := make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		roleNames = append(roleNames, string(r.Name))
	}

	// Primary role = highest privilege in the list
	primaryRole := primaryRole(u.Roles)

	claims := jwt.MapClaims{
		"sub":   u.ID.String(),
		"name":  u.Name,
		"role":  primaryRole, // single role for RoleGuard simplicity
		"roles": roleNames,   // all roles for fine-grained checks
		"exp":   time.Now().Add(s.jwtExpiry).Unix(),
		"iat":   time.Now().Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(s.jwtPrivateKey)
	if err != nil {
		return nil, err
	}

	rawRefresh, err := helper.GenerateRandomToken()
	if err != nil {
		return nil, err
	}
	if err := s.refreshTokenRepo.Create(&models.RefreshToken{
		UserID:    u.ID,
		Token:     helper.HashToken(rawRefresh),
		ExpiresAt: time.Now().Add(s.refreshExpiry),
	}); err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtExpiry.Seconds()),
	}, nil
}

// primaryRole returns the highest-privilege role from a list.
// Priority: super_admin > admin > user
func primaryRole(roles []models.RoleModel) string {
	priority := map[models.Role]int{
		models.RoleSuperAdmin: 3,
		models.RoleAdmin:      2,
		models.RoleUser:       1,
	}
	best := string(models.RoleUser)
	bestScore := 0
	for _, r := range roles {
		if s := priority[r.Name]; s > bestScore {
			bestScore = s
			best = string(r.Name)
		}
	}
	return best
}

// VerifyEmail melakukan aktivasi akun jika token verifikasi yang dikirimkan valid.
func (s *AuthService) VerifyEmail(token string) error {
	u, err := s.userRepo.FindByVerificationToken(token)
	if err != nil {
		return errors.New("token verifikasi tidak valid")
	}

	updates := map[string]interface{}{
		"is_verified":        true,
		"verification_token": "",
	}
	return s.userRepo.Update(u.ID, updates)
}

// ForgotPassword membuat token reset baru dan mengirimkannya melalui email pengguna.
func (s *AuthService) ForgotPassword(email string) error {
	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		// Mencegah enumerasi akun dengan mengembalikan sukses semu
		return nil
	}

	resetToken, _ := helper.GenerateRandomToken()
	expiresAt := time.Now().Add(1 * time.Hour) // Token kadaluwarsa dalam 1 jam

	updates := map[string]interface{}{
		"reset_token":            resetToken,
		"reset_token_expires_at": expiresAt,
	}
	if err := s.userRepo.Update(u.ID, updates); err != nil {
		return err
	}

	resetUrl := fmt.Sprintf("%s/auth/reset-password?token=%s", s.appURL, resetToken)
	emailBody := fmt.Sprintf("Halo %s,<br><br>Kami menerima permintaan untuk mereset kata sandi Anda. Silakan klik tombol di bawah untuk menyusun kata sandi baru:<br><br><a href='%s' style='background-color:#008CBA;color:white;padding:10px 20px;text-decoration:none;border-radius:5px;'>Reset Kata Sandi Saya</a><br><br>Link ini hanya berlaku selama 1 jam.", u.Name, resetUrl)
	_ = s.mailer.SendEmail(u.Email, "Reset Kata Sandi Akun Anda", emailBody)

	return nil
}

// ResetPassword merubah password pengguna jika token reset valid dan belum kadaluwarsa.
func (s *AuthService) ResetPassword(token, newPassword string) error {
	u, err := s.userRepo.FindByResetToken(token)
	if err != nil {
		return errors.New("token reset password tidak valid")
	}

	if u.ResetTokenExpiresAt == nil || u.ResetTokenExpiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"password":               string(hash),
		"reset_token":            "",
		"reset_token_expires_at": nil,
	}
	return s.userRepo.Update(u.ID, updates)
}

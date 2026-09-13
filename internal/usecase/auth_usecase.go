package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"backend-api/internal/auth"
	"backend-api/internal/domain"
	"backend-api/internal/repository"
)

type RegisterRequest struct {
	Email    string
	Password string
}

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResponse struct {
	User         domain.User
	AccessToken  string
	RefreshToken string
}
type RefreshRequest struct {
	RefreshToken string
}
type AuthUsecase struct {
	employeeRepo    repository.EmployeeRepository
	userRepo        repository.UserRepository
	authSessionRepo repository.AuthSessionRepository
	jwtSecret       string
	accessTokenTTL  time.Duration
}

func NewAuthUsecase(
	employeeRepo repository.EmployeeRepository,
	authSessionRepo repository.AuthSessionRepository,
	userRepo repository.UserRepository,
	jwtSecret string,
	accessTokenTTL time.Duration,
) *AuthUsecase {
	return &AuthUsecase{
		employeeRepo:    employeeRepo,
		authSessionRepo: authSessionRepo,
		userRepo:        userRepo,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTokenTTL,
	}
}

func (u *AuthUsecase) Register(
	ctx context.Context,
	req RegisterRequest,
) (domain.User, error) {

	employee, err := u.employeeRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return domain.User{}, err
	}

	_, err = u.userRepo.GetByEmployeeID(ctx, employee.ID)
	if err == nil {
		return domain.User{}, domain.ErrUserAlreadyExists
	}

	if !errors.Is(err, domain.ErrUserNotFound) {
		return domain.User{}, err
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		EmployeeID:   employee.ID,
		PasswordHash: passwordHash,
	}

	return u.userRepo.Create(ctx, user)
}

func (u *AuthUsecase) Login(
	ctx context.Context,
	req LoginRequest,
) (LoginResponse, error) {

	employee, err := u.employeeRepo.GetByEmail(
		ctx,
		req.Email,
	)
	if err != nil {
		return LoginResponse{}, domain.ErrInvalidCredentials
	}

	user, err := u.userRepo.GetByEmployeeID(
		ctx,
		employee.ID,
	)
	if err != nil {
		return LoginResponse{}, domain.ErrInvalidCredentials
	}

	err = auth.CheckPassword(
		req.Password,
		user.PasswordHash,
	)
	if err != nil {
		return LoginResponse{}, domain.ErrInvalidCredentials
	}

	// 1. إنشاء Access Token
	accessToken, err := auth.GenerateAccessToken(
		user.ID,
		user.EmployeeID,
		u.jwtSecret,
		u.accessTokenTTL,
	)
	if err != nil {
		return LoginResponse{}, err
	}

	// 2. إنشاء Refresh Token
	refreshToken, refreshTokenHash, err := auth.GenerateRefreshToken()
	if err != nil {
		return LoginResponse{}, err
	}

	// 3. تخزين Session في قاعدة البيانات
	session := domain.AuthSession{
		ID:               uuid.NewString(),
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour),
	}

	err = u.authSessionRepo.Create(ctx, session)
	if err != nil {
		return LoginResponse{}, err
	}

	// 4. إرجاع Access + Refresh Token
	return LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
func (u *AuthUsecase) Refresh(
	ctx context.Context,
	req RefreshRequest,
) (LoginResponse, error) {

	tokenHash := auth.HashRefreshToken(req.RefreshToken)

	session, err := u.authSessionRepo.GetByRefreshTokenHash(
		ctx,
		tokenHash,
	)
	if err != nil {
		return LoginResponse{}, domain.ErrSessionNotFound
	}

	if session.RevokedAt != nil {
		return LoginResponse{}, domain.ErrSessionNotFound
	}

	if time.Now().After(session.ExpiresAt) {
		return LoginResponse{}, domain.ErrSessionNotFound
	}

	user, err := u.userRepo.GetByID(
		ctx,
		session.UserID,
	)
	if err != nil {
		return LoginResponse{}, domain.ErrSessionNotFound
	}

	accessToken, err := auth.GenerateAccessToken(
		user.ID,
		user.EmployeeID,
		u.jwtSecret,
		u.accessTokenTTL,
	)
	if err != nil {
		return LoginResponse{}, err
	}

	refreshToken, refreshTokenHash, err := auth.GenerateRefreshToken()
	if err != nil {
		return LoginResponse{}, err
	}

	newSession := domain.AuthSession{
		ID:               uuid.NewString(),
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour),
	}

	err = u.authSessionRepo.Rotate(
		ctx,
		session.ID,
		newSession,
	)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
func (u *AuthUsecase) Logout(
    ctx context.Context,
    refreshToken string,
) error {
    tokenHash := auth.HashRefreshToken(refreshToken)

    session, err := u.authSessionRepo.GetByRefreshTokenHash(
        ctx,
        tokenHash,
    )
    if err != nil {
        return domain.ErrSessionNotFound
    }

    return u.authSessionRepo.Revoke(ctx, session.ID)
}
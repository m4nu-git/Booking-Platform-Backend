package services

import (
	env "AuthServiceInGo/config/env"
	db "AuthServiceInGo/db/repositories"
	"AuthServiceInGo/dto"
	"AuthServiceInGo/models"
	"AuthServiceInGo/utils"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserService interface {
	GetAllUserService() ([]*models.User, error)
	GetUserById(id string) (*models.User, error)
	CreateUser(payload *dto.CreateUserRequestDTO) (*models.User, error)
	LoginUser(payload *dto.LoginUserRequestDTO) (*dto.LoginResponseDTO, error)
	GetUserRoles(userId int64) ([]*models.Role, error)
	RefreshAccessToken(refreshToken string) (string, error)
	Logout(refreshToken string) error
}

type UserServiceImpl struct {
	userRepository         db.UserRepository
	roleRepository         db.RoleRepository
	userRoleRepository     db.UserRoleRepository
	refreshTokenRepository db.RefreshTokenRepository
}

func NewUserService(
	_userRepository db.UserRepository,
	_roleRepository db.RoleRepository,
	_userRoleRepository db.UserRoleRepository,
	_refreshTokenRepository db.RefreshTokenRepository,
) UserService {
	return &UserServiceImpl{
		userRepository:         _userRepository,
		roleRepository:         _roleRepository,
		userRoleRepository:     _userRoleRepository,
		refreshTokenRepository: _refreshTokenRepository,
	}
}

func (u *UserServiceImpl) GetAllUserService() ([]*models.User, error) {
	fmt.Println("Fetching All User from Service Layer")
	users, err := u.userRepository.GetAll()
	if err != nil {
		fmt.Println("Got error while fetching user from service layer", err)
		return nil, err
	}
	return users, nil
}

func (u *UserServiceImpl) GetUserById(id string) (*models.User, error) {
	fmt.Println("Fetching user in UserService")
	user, err := u.userRepository.GetByID(id)
	if err != nil {
		fmt.Println("Error fetching user:", err)
		return nil, err
	}
	return user, nil
}

func (u *UserServiceImpl) CreateUser(payload *dto.CreateUserRequestDTO) (*models.User, error) {
	fmt.Println("Creating user in UserService")

	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return nil, err
	}

	user, err := u.userRepository.Create(payload.Username, payload.Email, hashedPassword)
	if err != nil {
		fmt.Println("Error creating user:", err)
		return nil, err
	}

	role, err := u.roleRepository.GetRoleByName("user")
	if err != nil {
		fmt.Println("Error fetching default role:", err)
		return nil, fmt.Errorf("failed to find default role 'user': %w", err)
	}

	if err := u.userRoleRepository.AssignRoleToUser(user.Id, role.Id); err != nil {
		fmt.Println("Error assigning role to user:", err)
		return nil, fmt.Errorf("failed to assign role to user: %w", err)
	}

	fmt.Printf("User created successfully with default role '%s': %+v\n", role.Name, user)

	return user, nil
}

func (u *UserServiceImpl) GetUserRoles(userId int64) ([]*models.Role, error) {
	return u.userRoleRepository.GetUserRoles(userId)
}

func (u *UserServiceImpl) LoginUser(payload *dto.LoginUserRequestDTO) (*dto.LoginResponseDTO, error) {
	user, err := u.userRepository.GetByEmail(payload.Email)
	if err != nil {
		fmt.Println("Error fetching user by email:", err)
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("no user found with email: %s", payload.Email)
	}

	if !utils.CheckPasswordHash(payload.Password, user.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate access token (15 min expiry)
	jwtPayload := jwt.MapClaims{
		"email": user.Email,
		"id":    user.Id,
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtPayload)
	accessToken, err := token.SignedString([]byte(env.GetString("JWT_SECRET", "TOKEN")))
	if err != nil {
		fmt.Println("Error signing access token:", err)
		return nil, err
	}

	// Generate refresh token (random 32-byte hex, 7 days expiry)
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := u.refreshTokenRepository.Create(user.Id, refreshToken, expiresAt); err != nil {
		fmt.Println("Error storing refresh token:", err)
		return nil, err
	}

	return &dto.LoginResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *UserServiceImpl) RefreshAccessToken(refreshToken string) (string, error) {
	rt, err := u.refreshTokenRepository.GetByToken(refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("invalid refresh token")
		}
		return "", err
	}

	// Check expiry
	expiresAt, err := time.Parse("2006-01-02 15:04:05", rt.ExpiresAt)
	if err != nil {
		return "", fmt.Errorf("error parsing token expiry")
	}

	if time.Now().After(expiresAt) {
		u.refreshTokenRepository.DeleteByToken(refreshToken)
		return "", fmt.Errorf("refresh token expired, please login again")
	}

	// Issue new access token
	jwtPayload := jwt.MapClaims{
		"id":  rt.UserId,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtPayload)
	accessToken, err := token.SignedString([]byte(env.GetString("JWT_SECRET", "TOKEN")))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (u *UserServiceImpl) Logout(refreshToken string) error {
	return u.refreshTokenRepository.DeleteByToken(refreshToken)
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

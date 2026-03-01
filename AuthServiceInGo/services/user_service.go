package services

import (
	env "AuthServiceInGo/config/env"
	db "AuthServiceInGo/db/repositories"
	"AuthServiceInGo/dto"
	"AuthServiceInGo/models"
	"AuthServiceInGo/utils"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type UserService interface {
	GetAllUserService() ([]*models.User, error)
	GetUserById(id string) (*models.User, error)
	CreateUser(payload *dto.CreateUserRequestDTO) (*models.User, error)
	LoginUser(payload *dto.LoginUserRequestDTO) (string, error)
}

type UserServiceImpl struct {
	userRepository     db.UserRepository
	roleRepository     db.RoleRepository
	userRoleRepository db.UserRoleRepository
}

func NewUserService(_userRepository db.UserRepository, _roleRepository db.RoleRepository, _userRoleRepository db.UserRoleRepository) UserService {
	return &UserServiceImpl{
		userRepository:     _userRepository,
		roleRepository:     _roleRepository,
		userRoleRepository: _userRoleRepository,
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

	// Step 1. Hash the password using utils.HashPassword
	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return nil, err
	}

	// Step 2. Call the repository to create the user (pure insert)
	user, err := u.userRepository.Create(payload.Username, payload.Email, hashedPassword)
	if err != nil {
		fmt.Println("Error creating user:", err)
		return nil, err
	}

	// Step 3. Assign the default "user" role via RoleRepository + UserRoleRepository
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

	// Step 4. Return the created user
	return user, nil
}

func (u *UserServiceImpl) LoginUser(payload *dto.LoginUserRequestDTO) (string, error) {
	// Pre-requisite: This function will be given email and password as parameter, which we can hardcode for now
	email := payload.Email
	password := payload.Password

	// Step 1. Make a repository call to get the user by email
	user, err := u.userRepository.GetByEmail(email)

	if err != nil {
		fmt.Println("Error fetching user by email:", err)
		return "", err
	}

	// Step 2. If user exists, or not. If not exists, return error
	if user == nil {
		fmt.Println("No user found with the given email")
		return "", fmt.Errorf("no user found with email: %s", email)
	}

	// Step 3. If user exists, check the password using utils.CheckPasswordHash
	isPasswordValid := utils.CheckPasswordHash(password, user.Password)

	if !isPasswordValid {
		fmt.Println("Password does not match")
		return "", nil
	}

	// Step 4. If password matches, print a JWT token, else return error saying password does not match
	jwtPayload := jwt.MapClaims{
		"email": user.Email,
		"id":    user.Id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtPayload)

	tokenString, err := token.SignedString([]byte(env.GetString("JWT_SECRET", "TOKEN")))

	if err != nil {
		fmt.Println("Error signing token:", err)
		return "", err
	}

	fmt.Println("JWT Token:", tokenString)

	return tokenString, nil
}

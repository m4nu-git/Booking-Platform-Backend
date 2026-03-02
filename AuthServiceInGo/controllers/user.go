package controllers

import (
	"AuthServiceInGo/dto"
	"AuthServiceInGo/services"
	"AuthServiceInGo/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserController struct {
	UserService services.UserService
}

func NewUserController(_userService services.UserService) *UserController {
	return &UserController{
		UserService: _userService,
	}
}

func (uc *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	payload := r.Context().Value("payload").(dto.CreateUserRequestDTO)

	fmt.Println("Payload received:", payload)

	user, err := uc.UserService.CreateUser(&payload)

	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusCreated, "User created successfully", user)
	fmt.Println("User created successfully:", user)
}

func (uc *UserController) LoginUser(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Logging in user in UserController")

	payload := r.Context().Value("payload").(dto.LoginUserRequestDTO)

	fmt.Println("Payload received:", payload)

	jwtToken, err := uc.UserService.LoginUser(&payload)

	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to login user", err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "User logged in successfully", jwtToken)

}

func (uc *UserController) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")
	if userId == "" {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "User ID is required", fmt.Errorf("missing user ID"))
		return
	}

	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	roles, err := uc.UserService.GetUserRoles(id)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to fetch user roles", err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "User roles fetched successfully", roles)
}

func (uc *UserController) GetUserById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Fetching user by ID in UserController")
	userId := r.URL.Query().Get("id")
	if userId == "" {
		userId = r.Context().Value("userID").(string)
	}

	fmt.Println("User ID from context or query:", userId)

	if userId == "" {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "User ID is required", fmt.Errorf("missing user ID"))
		return
	}

	user, err := uc.UserService.GetUserById(userId)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to fetch user", err)
		return
	}
	if user == nil {
		utils.WriteJsonErrorResponse(w, http.StatusNotFound, "User not found", fmt.Errorf("user with ID %s not found", userId))
		return
	}

	userIdInt, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	roles, err := uc.UserService.GetUserRoles(userIdInt)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to fetch user roles", err)
		return
	}

	roleDTOs := make([]dto.RoleDTO, 0, len(roles))
	for _, r := range roles {
		roleDTOs = append(roleDTOs, dto.RoleDTO{Id: r.Id, Name: r.Name})
	}

	profile := dto.ProfileResponseDTO{
		Id:        user.Id,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Roles:     roleDTOs,
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "User fetched successfully", profile)
	fmt.Println("User fetched successfully:", user)
}
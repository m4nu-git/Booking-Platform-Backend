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

type PermissionController struct {
	PermissionService services.PermissionService
}

func NewPermissionController(permissionService services.PermissionService) *PermissionController {
	return &PermissionController{
		PermissionService: permissionService,
	}
}

func (pc *PermissionController) GetPermissionById(w http.ResponseWriter, r *http.Request) {
	permissionId := chi.URLParam(r, "id")
	if permissionId == "" {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "Permission ID is required", fmt.Errorf("missing permission ID"))
		return
	}

	id, err := strconv.ParseInt(permissionId, 10, 64)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid permission ID", err)
		return
	}

	permission, err := pc.PermissionService.GetPermissionById(id)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to fetch permission", err)
		return
	}

	if permission == nil {
		utils.WriteJsonErrorResponse(w, http.StatusNotFound, "Permission not found", fmt.Errorf("permission with ID %s not found", permissionId))
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Permission fetched successfully", permission)
}

func (pc *PermissionController) GetAllPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := pc.PermissionService.GetAllPermissions()
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to fetch permissions", err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Permissions fetched successfully", permissions)
}

func (pc *PermissionController) CreatePermission(w http.ResponseWriter, r *http.Request) {
	payload := r.Context().Value("payload").(dto.CreatePermissionRequestDTO)

	permission, err := pc.PermissionService.CreatePermission(payload.PermissionName, payload.Description, payload.Resource, payload.Action)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to create permission", err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusCreated, "Permission created successfully", permission)
}

func (pc *PermissionController) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	permissionId := chi.URLParam(r, "id")
	if permissionId == "" {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "Permission ID is required", fmt.Errorf("missing permission ID"))
		return
	}

	id, err := strconv.ParseInt(permissionId, 10, 64)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid permission ID", err)
		return
	}

	payload := r.Context().Value("payload").(dto.UpdatePermissionRequestDTO)

	permission, err := pc.PermissionService.UpdatePermission(id, payload.PermissionName, payload.Description, payload.Resource, payload.Action)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to update permission", err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Permission updated successfully", permission)
}

func (pc *PermissionController) DeletePermission(w http.ResponseWriter, r *http.Request) {
	permissionId := chi.URLParam(r, "id")
	if permissionId == "" {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "Permission ID is required", fmt.Errorf("missing permission ID"))
		return
	}

	id, err := strconv.ParseInt(permissionId, 10, 64)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid permission ID", err)
		return
	}

	err = pc.PermissionService.DeletePermissionById(id)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to delete permission", err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Permission deleted successfully", nil)
}

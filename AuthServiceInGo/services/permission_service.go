package services

import (
	repositories "AuthServiceInGo/db/repositories"
	"AuthServiceInGo/models"
)

type PermissionService interface {
	GetPermissionById(id int64) (*models.Permission, error)
	GetPermissionByName(name string) (*models.Permission, error)
	GetAllPermissions() ([]*models.Permission, error)
	CreatePermission(name string, description string, resource string, action string) (*models.Permission, error)
	DeletePermissionById(id int64) error
	UpdatePermission(id int64, name string, description string, resource string, action string) (*models.Permission, error)
}

type PermissionServiceImpl struct {
	permissionRepository repositories.PermissionRepository
}

func NewPermissionService(permissionRepo repositories.PermissionRepository) PermissionService {
	return &PermissionServiceImpl{
		permissionRepository: permissionRepo,
	}
}

func (s *PermissionServiceImpl) GetPermissionById(id int64) (*models.Permission, error) {
	return s.permissionRepository.GetPermissionById(id)
}

func (s *PermissionServiceImpl) GetPermissionByName(name string) (*models.Permission, error) {
	return s.permissionRepository.GetPermissionByName(name)
}

func (s *PermissionServiceImpl) GetAllPermissions() ([]*models.Permission, error) {
	return s.permissionRepository.GetAllPermissions()
}

func (s *PermissionServiceImpl) CreatePermission(name string, description string, resource string, action string) (*models.Permission, error) {
	return s.permissionRepository.CreatePermission(name, description, resource, action)
}

func (s *PermissionServiceImpl) DeletePermissionById(id int64) error {
	return s.permissionRepository.DeletePermissionById(id)
}

func (s *PermissionServiceImpl) UpdatePermission(id int64, name string, description string, resource string, action string) (*models.Permission, error) {
	return s.permissionRepository.UpdatePermission(id, name, description, resource, action)
}

package router

import (
	"AuthServiceInGo/controllers"
	"AuthServiceInGo/middlewares"

	"github.com/go-chi/chi/v5"
)

type PermissionRouter struct {
	permissionController *controllers.PermissionController
}

func NewPermissionRouter(_permissionController *controllers.PermissionController) Router {
	return &PermissionRouter{
		permissionController: _permissionController,
	}
}

func (pr *PermissionRouter) Register(r chi.Router) {
	r.Get("/permissions/{id}", pr.permissionController.GetPermissionById)
	r.Get("/permissions", pr.permissionController.GetAllPermissions)
	r.With(middlewares.CreatePermissionRequestValidator).Post("/permissions", pr.permissionController.CreatePermission)
	r.With(middlewares.UpdatePermissionRequestValidator).Put("/permissions/{id}", pr.permissionController.UpdatePermission)
	r.Delete("/permissions/{id}", pr.permissionController.DeletePermission)
}

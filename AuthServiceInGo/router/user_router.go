package router

import (
	"AuthServiceInGo/controllers"
	"AuthServiceInGo/middlewares"

	"github.com/go-chi/chi/v5"
)

type UserRouter struct {
	userController *controllers.UserController
}

func NewUserRouter(_userController *controllers.UserController) Router {
	return &UserRouter{
		userController: _userController,
	}
}

func (ur *UserRouter) Register(r chi.Router) {
	r.With(middlewares.JWTAuthMiddleware, middlewares.RequireAnyRole("user", "admin")).Get("/profile", ur.userController.GetUserById)
	r.With(middlewares.JWTAuthMiddleware, middlewares.RequireAnyRole("user", "admin"), middlewares.UpdateProfileRequestValidator).Patch("/profile", ur.userController.UpdateProfile)
	r.With(middlewares.JWTAuthMiddleware, middlewares.RequireAnyRole("user", "admin"), middlewares.ChangePasswordRequestValidator).Post("/change-password", ur.userController.ChangePassword)
	r.With(middlewares.UserCreateRequestValidator).Post("/signup", ur.userController.CreateUser)
	r.With(middlewares.UserLoginRequestValidator).Post("/login", ur.userController.LoginUser)
	r.Get("/users/{id}/roles", ur.userController.GetUserRoles)
	r.With(middlewares.ForgotPasswordRequestValidator).Post("/forgot-password", ur.userController.ForgotPassword)
	r.Get("/reset-password", ur.userController.ServeResetPasswordPage)
	r.With(middlewares.ResetPasswordRequestValidator).Post("/reset-password", ur.userController.ResetPassword)
	r.With(middlewares.RefreshTokenRequestValidator).Post("/refresh", ur.userController.RefreshToken)
	r.With(middlewares.RefreshTokenRequestValidator).Post("/logout", ur.userController.Logout)
}

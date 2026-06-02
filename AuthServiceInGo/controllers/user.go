package controllers

import (
	"AuthServiceInGo/dto"
	"AuthServiceInGo/services"
	"AuthServiceInGo/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

var resetPasswordTmpl = template.Must(template.New("reset").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
  <title>Reset Password</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #f5f5f5; display: flex; align-items: center; justify-content: center; min-height: 100vh; }
    .card { background: #fff; padding: 2.5rem 2rem; border-radius: 12px; box-shadow: 0 4px 24px rgba(0,0,0,.08); width: 100%; max-width: 400px; }
    h1 { font-size: 1.5rem; color: #1a1a1a; margin-bottom: .5rem; }
    p.sub { color: #666; font-size: .9rem; margin-bottom: 1.75rem; }
    label { display: block; font-size: .85rem; font-weight: 600; color: #333; margin-bottom: .4rem; }
    input[type=password] { width: 100%; padding: .65rem .9rem; border: 1.5px solid #ddd; border-radius: 8px; font-size: 1rem; outline: none; transition: border .2s; }
    input[type=password]:focus { border-color: #FF5A5F; }
    .field { margin-bottom: 1.1rem; }
    button { width: 100%; padding: .75rem; background: #FF5A5F; color: #fff; border: none; border-radius: 8px; font-size: 1rem; font-weight: 600; cursor: pointer; margin-top: .5rem; transition: background .2s; }
    button:hover { background: #e0484c; }
    button:disabled { background: #ccc; cursor: not-allowed; }
    .msg { margin-top: 1rem; padding: .75rem 1rem; border-radius: 8px; font-size: .9rem; }
    .msg.success { background: #e6f9ef; color: #1a7a3f; }
    .msg.error   { background: #fdecea; color: #c62828; }
  </style>
</head>
<body>
<div class="card">
  <h1>Reset your password</h1>
  {{if .Token}}
  <p class="sub">Enter a new password for your account.</p>
  <form id="form" data-token="{{.Token}}">
    <div class="field">
      <label for="pw">New password</label>
      <input id="pw" type="password" placeholder="At least 8 characters" minlength="8" required/>
    </div>
    <div class="field">
      <label for="pw2">Confirm password</label>
      <input id="pw2" type="password" placeholder="Repeat your password" minlength="8" required/>
    </div>
    <button type="submit" id="btn">Reset password</button>
  </form>
  <div class="msg" id="msg"></div>
  <script>
    document.getElementById("form").addEventListener("submit", async function(e) {
      e.preventDefault();
      const token = document.getElementById("form").dataset.token;
      const pw  = document.getElementById("pw").value;
      const pw2 = document.getElementById("pw2").value;
      const msg = document.getElementById("msg");
      const btn = document.getElementById("btn");
      msg.className = "msg"; msg.textContent = "";
      if (pw !== pw2) { msg.className="msg error"; msg.textContent="Passwords do not match."; return; }
      btn.disabled = true; btn.textContent = "Resetting…";
      try {
        const res = await fetch("/reset-password", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ token: token, new_password: pw })
        });
        const data = await res.json();
        if (res.ok) {
          msg.className="msg success";
          msg.textContent = data.message || "Password reset! You can now log in.";
          document.getElementById("form").style.display="none";
        } else {
          msg.className="msg error";
          msg.textContent = data.message || "Something went wrong. The link may have expired.";
          btn.disabled=false; btn.textContent="Reset password";
        }
      } catch {
        msg.className="msg error"; msg.textContent="Network error. Please try again.";
        btn.disabled=false; btn.textContent="Reset password";
      }
    });
  </script>
  {{else}}
  <div class="msg error">This reset link is invalid or missing. Please request a new one from the forgot password page.</div>
  {{end}}
</div>
</body>
</html>
`))

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

	loginResponse, err := uc.UserService.LoginUser(&payload)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusUnauthorized, "Failed to login user", err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "User logged in successfully", loginResponse)
}

func (uc *UserController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	payload := r.Context().Value("payload").(dto.RefreshTokenRequestDTO)

	accessToken, err := uc.UserService.RefreshAccessToken(payload.RefreshToken)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Token refreshed successfully", map[string]string{
		"access_token": accessToken,
	})
}

func (uc *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	payload := r.Context().Value("payload").(dto.RefreshTokenRequestDTO)

	if err := uc.UserService.Logout(payload.RefreshToken); err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to logout", err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Logged out successfully", nil)
}

func (uc *UserController) ServeResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	resetPasswordTmpl.Execute(w, map[string]string{"Token": token})
}

func (uc *UserController) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	payload := r.Context().Value("payload").(dto.ForgotPasswordRequestDTO)
	if err := uc.UserService.ForgotPassword(&payload); err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to process request", err)
		return
	}
	// Generic message — never reveal whether the email exists
	utils.WriteJsonSuccessResponse(w, http.StatusOK, "If that email is registered you will receive a reset link shortly", nil)
}

func (uc *UserController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	payload := r.Context().Value("payload").(dto.ResetPasswordRequestDTO)
	if err := uc.UserService.ResetPassword(&payload); err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Password reset successfully. Please log in with your new password.", nil)
}

func (uc *UserController) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	payload := r.Context().Value("payload").(dto.UpdateProfileRequestDTO)
	userIDStr := r.Context().Value("userID").(string)
	userId, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid user ID in token", err)
		return
	}
	user, err := uc.UserService.UpdateUserProfile(userId, &payload)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to update profile", err)
		return
	}
	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Profile updated successfully", user)
}

func (uc *UserController) ChangePassword(w http.ResponseWriter, r *http.Request) {
	payload := r.Context().Value("payload").(dto.ChangePasswordRequestDTO)
	userIDStr := r.Context().Value("userID").(string)
	userId, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "Invalid user ID in token", err)
		return
	}
	if err := uc.UserService.ChangePassword(userId, &payload); err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Password changed successfully", nil)
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
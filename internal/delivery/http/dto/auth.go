package dto

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

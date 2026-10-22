package authhandler

import "github.com/victorinolavida/gateway/internal/application/dtos"

// AuthUserResponse is the public HTTP contract of an authenticated user.
// It is kept in the http adapter so changes in the application layer do not
// silently change the API.
type AuthUserResponse struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Token     string `json:"token"`
}

func newAuthUserResponse(result dtos.AuthUserResponse) AuthUserResponse {
	return AuthUserResponse{
		UserID:    result.UserID,
		FirstName: result.FirstName,
		LastName:  result.LastName,
		Token:     result.Token,
	}
}

type WhoamiUserResponse struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

func newWhoamiResponse(result dtos.WhoamiUserResponse) WhoamiUserResponse {
	return WhoamiUserResponse{
		UserID:    result.UserID,
		FirstName: result.FirstName,
		LastName:  result.LastName,
		Email:     result.Email,
	}
}

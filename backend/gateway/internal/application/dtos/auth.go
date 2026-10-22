package dtos

// LoginRequest is the application input of the login use case.
type LoginRequest struct {
	Email    string
	Password string
}

// RegisterRequest is the application input of the register use case.
type RegisterRequest struct {
	Email     string
	FirstName string
	LastName  string
	Password  string
}

// AuthUserResponse is the application output of the auth use cases.
type AuthUserResponse struct {
	UserID    string
	FirstName string
	LastName  string
	Token     string
}

type WhoamiUserRequest struct {
	Token string
}

type WhoamiUserResponse struct {
	UserID    string
	FirstName string
	LastName  string
	Token     string
	Email     string
}

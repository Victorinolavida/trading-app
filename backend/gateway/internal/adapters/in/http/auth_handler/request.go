package authhandler

// LoginRequest is the HTTP request body of POST /auth/login.
//
// The json and validate tags belong to the http adapter: the application
// layer receives the same data without any transport concern.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest is the HTTP request body of POST /auth/register.
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Password  string `json:"password" validate:"required"`
}

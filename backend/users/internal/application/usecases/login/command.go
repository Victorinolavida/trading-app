package login

type LoginCommand struct {
	Email    string
	Password string
}

type LoginResult struct {
	User struct {
		ID        string
		Email     string
		FirstName string
		LastName  string
	}
	Token string
}

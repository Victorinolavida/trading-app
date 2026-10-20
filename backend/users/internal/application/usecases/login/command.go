package login

type Command struct {
	Email    string
	Password string
}

type Result struct {
	User struct {
		ID        string
		Email     string
		FirstName string
		LastName  string
	}
	Token string
}


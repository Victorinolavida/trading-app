package register

type RegisterCommand struct {
	Email     string
	FirstName string
	LastName  string
	Password  string
}

type RegisterResult struct {
	User struct {
		ID        string
		Email     string
		FirstName string
		LastName  string
	}
	Token string
}

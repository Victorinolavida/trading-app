package whoami

type Command struct {
	Token string
}

type Result struct {
	UserID    string
	FirstName string
	LastName  string
	Email     string
}

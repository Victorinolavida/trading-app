package whoami

type Command struct {
	UserID string
}

type Result struct {
	UserID    string
	FirstName string
	LastName  string
	Email     string
}

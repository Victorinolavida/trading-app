package getsecuritybyparams

import (
	"time"
)

type Command struct {
	ID             string
	ISIN           string
	Serie          string
	SecurityType   string
	IssueBefore    time.Time
	IssueAfter     time.Time
	MaturityBefore time.Time
	MaturityAfter  time.Time
}

type Security struct {
	ID           string
	Isin         string
	Name         string
	Type         string
	Serie        string
	IssueDate    string
	MaturityDate string
	NominalValue string
	Currency     string
	Fx           string
	Enable       bool
}

type Result struct {
	Security Security
}

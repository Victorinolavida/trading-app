package getsecurities

type (
	Security struct {
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
	Result struct {
		Results []Security
	}
)

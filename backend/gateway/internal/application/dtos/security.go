package dtos

type ListSecuritiesRequest struct {
	Id             string
	Ids            []string
	Isin           string
	Isins          []string
	Currency       string
	Serie          string
	Series         []string
	SecurityType   string
	SecurityTypes  []string
	IssueBefore    string
	IssueAfter     string
	MaturityBefore string
	MaturityAfter  string
	Name           string
	Query          string
	Q              string
}

type Security struct {
	ID           string  `json:"id"`
	Isin         string  `json:"isin"`
	Name         string  `json:"name"`
	SecurityType string  `json:"security_type"`
	Serie        string  `json:"serie"`
	IssueDate    string  `json:"issue_date"`
	MaturityDate string  `json:"maturity_date"`
	NominalValue float64 `json:"nominal_value"`
	Currency     string  `json:"currency"`
	Fx           float64 `json:"fx"`
	Enable       bool    `json:"enable"`
}

type ListSecuritiesResponse struct {
	Securities []Security `json:"securities"`
}

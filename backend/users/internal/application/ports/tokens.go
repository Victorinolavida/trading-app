package ports

import "context"

type Claims struct {
	UserID    string
	FirstName string
	LastName  string
}

type TokenVerifier interface {
	Verify(ctx context.Context, raw string) (Claims, error)
}

type TokenIssuer interface {
	Issue(ctx context.Context, p Claims) (string, error)
}

package users

import (
	"context"

	"github.com/google/uuid"
	"github.com/victorinolavida/orders/internal/application/ports"
	authpb "github.com/victorinolavida/proto/auth"
	"google.golang.org/grpc"
)

type Service struct {
	client authpb.AuthServiceClient
}

func New(conn *grpc.ClientConn) *Service {
	return &Service{client: authpb.NewAuthServiceClient(conn)}
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*ports.UserInfo, error) {
	res, err := s.client.Whoami(ctx, &authpb.WhoamiRequest{UserId: id.String()})
	if err != nil {
		return nil, err
	}
	userID, _ := uuid.Parse(res.Id)
	return &ports.UserInfo{
		ID:        userID,
		FirstName: res.FirstName,
		LastName:  res.LastName,
		Email:     res.Email,
	}, nil
}

func (s *Service) Whoami(ctx context.Context, userID uuid.UUID) (*ports.UserInfo, error) {
	return s.GetByID(ctx, userID)
}

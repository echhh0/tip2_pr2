package grpcapi

import (
	"context"
	"errors"

	"tip2_pr2/proto"
	"tip2_pr2/services/auth/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	proto.UnimplementedAuthServiceServer
	authService *service.AuthService
}

func New(authService *service.AuthService) *Server {
	return &Server{authService: authService}
}

func (s *Server) Verify(ctx context.Context, req *proto.VerifyRequest) (*proto.VerifyResponse, error) {
	subject, err := s.authService.Verify(ctx, req.Token)
	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.VerifyResponse{
		Valid:   true,
		Subject: subject,
	}, nil
}

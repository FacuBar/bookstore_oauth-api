package oauth_grpc

import (
	"context"
	"net"
	"net/http"

	"github.com/FacuBar/bookstore_oauth-api/pkg/core/ports"
	"github.com/FacuBar/bookstore_oauth-api/pkg/infraestructure/http/grpc/oauth/oauthpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	as ports.AcessTokenService
}

func (s *server) ValidateToken(
	ctx context.Context,
	req *oauthpb.ValidateTokenRequest,
) (*oauthpb.ValidateTokenResponse, error) {
	at := req.GetAccessToken()

	// if ctx.Err() == context.Canceled {
	// 	return nil, status.Error(codes.Canceled, "request is canceled")
	// }

	// if ctx.Err() == context.DeadlineExceeded {
	// 	return nil, status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	// }

	accessToken, err := s.as.GetById(at)
	if err != nil {
		switch err.Status() {
		case http.StatusNotFound:
			return nil, status.Error(codes.NotFound, err.Message())

		case http.StatusBadRequest:
			return nil, status.Error(codes.InvalidArgument, err.Message())

		case http.StatusUnauthorized:
			return nil, status.Error(codes.Unauthenticated, err.Message())

		default:
			return nil, status.Error(codes.Internal, err.Message())
		}

	}

	res := &oauthpb.ValidateTokenResponse{
		UserPayload: &oauthpb.ValidateTokenResponse_UserPayload{
			UserId: accessToken.UserId,
			Role:   getRole(accessToken.UserRole),
		},
	}

	return res, nil
}

func NewGRPCServer(address string, as ports.AcessTokenService) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer()

	oauthpb.RegisterOauthServiceServer(s, &server{as: as})

	go s.Serve(lis)

	return s, nil
}

func getRole(role string) oauthpb.ValidateTokenResponse_UserPayload_Role {
	switch role {
	case "user":
		return oauthpb.ValidateTokenResponse_UserPayload_USER
	case "admin":
		return oauthpb.ValidateTokenResponse_UserPayload_ADMIN
	default:
		return oauthpb.ValidateTokenResponse_UserPayload_UNKNOWN
	}
}

package oauth_grpc

import (
	"context"
	"testing"

	"github.com/FacuBar/bookstore_oauth-api/pkg/core/domain"
	"github.com/FacuBar/bookstore_oauth-api/pkg/core/ports"
	"github.com/FacuBar/bookstore_oauth-api/pkg/infraestructure/http/grpc/oauth/oauthpb"
	"github.com/FacuBar/bookstore_utils-go/rest_errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type atServiceMock struct{}

var (
	funcGetById func(string) (*domain.AccessToken, rest_errors.RestErr)
)

var serviceMock ports.AcessTokenService = &atServiceMock{}

func (*atServiceMock) Create(email string, password string) (*domain.AccessToken, rest_errors.RestErr) {
	return nil, nil
}
func (*atServiceMock) GetById(id string) (*domain.AccessToken, rest_errors.RestErr) {
	return funcGetById(id)
}

func TestValidateToken(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		funcGetById = func(s string) (*domain.AccessToken, rest_errors.RestErr) {
			return &domain.AccessToken{
				UserId:   1,
				UserRole: "admin",
			}, nil
		}

		s := server{as: serviceMock}

		req := &oauthpb.ValidateTokenRequest{
			AccessToken: "b255ce76-4a87-4293-ae19-08768c96ea05",
		}

		res, err := s.ValidateToken(context.Background(), req)

		assert.Nil(t, err)

		assert.EqualValues(t, 1, res.GetUserPayload().GetUserId())
		assert.EqualValues(t, oauthpb.ValidateTokenResponse_UserPayload_ADMIN, res.GetUserPayload().GetRole())
	})

	t.Run("BadRequest", func(t *testing.T) {
		funcGetById = func(s string) (*domain.AccessToken, rest_errors.RestErr) {
			return nil, rest_errors.NewBadRequestError("invalid access token id")
		}

		s := server{as: serviceMock}

		req := &oauthpb.ValidateTokenRequest{
			AccessToken: "",
		}

		res, err := s.ValidateToken(context.Background(), req)

		assert.Nil(t, res)
		assert.NotNil(t, err)
		assert.EqualValues(t, "rpc error: code = InvalidArgument desc = invalid access token id", err.Error())
	})

	t.Run("Unauthorized", func(t *testing.T) {
		funcGetById = func(s string) (*domain.AccessToken, rest_errors.RestErr) {
			return nil, rest_errors.NewUnauthorizedError("Token expired")
		}

		s := server{as: serviceMock}

		req := &oauthpb.ValidateTokenRequest{
			AccessToken: "b255ce76-4a87-4293-ae19-08768c96ea05",
		}

		res, err := s.ValidateToken(context.Background(), req)

		assert.Nil(t, res)
		assert.NotNil(t, err)
		assert.EqualValues(t, "rpc error: code = Unauthenticated desc = Token expired", err.Error())
	})

	t.Run("InternalServerError", func(t *testing.T) {
		funcGetById = func(s string) (*domain.AccessToken, rest_errors.RestErr) {
			return nil, rest_errors.NewInternalServerError("db error")
		}

		s := server{as: serviceMock}

		req := &oauthpb.ValidateTokenRequest{
			AccessToken: "b255ce76-4a87-4293-ae19-08768c96ea05",
		}

		res, err := s.ValidateToken(context.Background(), req)

		assert.Nil(t, res)
		assert.NotNil(t, err)
		assert.EqualValues(t, "rpc error: code = Internal desc = db error", err.Error())
	})

	t.Run("StatusNotFound", func(t *testing.T) {
		funcGetById = func(s string) (*domain.AccessToken, rest_errors.RestErr) {
			return nil, rest_errors.NewNotFoundError("access_token not found")
		}

		s := server{as: serviceMock}

		req := &oauthpb.ValidateTokenRequest{
			AccessToken: "b255ce76-4a87-4293-ae19-08768c96ea05",
		}

		res, err := s.ValidateToken(context.Background(), req)

		assert.Nil(t, res)
		assert.NotNil(t, err)
		assert.EqualValues(t, "rpc error: code = NotFound desc = access_token not found", err.Error())
	})
}

func TestServer(t *testing.T) {
	funcGetById = func(s string) (*domain.AccessToken, rest_errors.RestErr) {
		return &domain.AccessToken{
			UserId:   2,
			UserRole: "user",
		}, nil
	}

	s, err := NewGRPCServer("0.0.0.0:50051", &atServiceMock{})
	defer s.Stop()

	cc, _ := grpc.Dial("0.0.0.0:50051", grpc.WithInsecure())
	defer cc.Close()

	c := oauthpb.NewOauthServiceClient(cc)

	req := &oauthpb.ValidateTokenRequest{
		AccessToken: "b255ce76-4a87-4293-ae19-08768c96ea05",
	}
	res, errCall := c.ValidateToken(context.Background(), req)

	uId := res.GetUserPayload().GetUserId()
	uRole := res.GetUserPayload().GetRole()

	assert.EqualValues(t, 2, uId)
	assert.EqualValues(t, oauthpb.ValidateTokenResponse_UserPayload_Role_value["USER"], uRole)
	assert.Nil(t, err)
	assert.Nil(t, errCall)
	assert.NotNil(t, s)
}

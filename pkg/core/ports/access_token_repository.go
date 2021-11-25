package ports

import (
	"github.com/FacuBar/bookstore_oauth-api/pkg/core/domain"
	"github.com/FacuBar/bookstore_utils-go/rest_errors"
)

type AccessTokenRepository interface {
	// Db where access_tokens will be stored
	Create(domain.AccessToken) rest_errors.RestErr
	GetById(string) (*domain.AccessToken, rest_errors.RestErr)

	// Rest client that will interact with the users_api service
	LoginUser(string, string) (*domain.User, rest_errors.RestErr)
}

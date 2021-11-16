package ports

import (
	"github.com/FacuBar/bookstore_oauth-api/pkg/core/domain"
	"github.com/FacuBar/bookstore_utils-go/rest_errors"
)

type AcessTokenService interface {
	Create(domain.User) (*domain.AccessToken, rest_errors.RestErr)
	GetById(string) (*domain.AccessToken, rest_errors.RestErr)
	// UpdateExpirationTime ()
}

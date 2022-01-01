package ports

import (
	"github.com/FacuBar/bookstore_oauth-api/pkg/core/domain"
	"github.com/FacuBar/bookstore_utils-go/rest_errors"
)

type UsersService interface {
	Login(string, string) (*domain.User, rest_errors.RestErr)
}

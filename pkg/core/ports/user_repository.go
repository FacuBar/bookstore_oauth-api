package ports

import (
	"github.com/FacuBar/bookstore_oauth-api/pkg/core/domain"
	"github.com/FacuBar/bookstore_utils-go/rest_errors"
)

type UsersRepository interface {
	GetByEmail(string) (*domain.User, rest_errors.RestErr)
	Save(*domain.User) rest_errors.RestErr
}

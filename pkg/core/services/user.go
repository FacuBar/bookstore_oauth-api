package services

import (
	"net/http"
	"strings"
	"sync"

	"github.com/FacuBar/bookstore_oauth-api/pkg/core/domain"
	"github.com/FacuBar/bookstore_oauth-api/pkg/core/ports"
	"github.com/FacuBar/bookstore_utils-go/rest_errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	onceUsersService     sync.Once
	instanceUsersService *usersService
)

type usersService struct {
	repo ports.UsersRepository
}

func NewUsersService(repo ports.UsersRepository) ports.UsersService {
	onceUsersService.Do(func() {
		instanceUsersService = &usersService{
			repo: repo,
		}
	})
	return instanceUsersService
}

func (s *usersService) Login(email, password string) (*domain.User, rest_errors.RestErr) {
	user, err := s.repo.GetByEmail(strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		switch err.Status() {
		case http.StatusInternalServerError:
			return nil, rest_errors.NewInternalServerError("error while trying to login, try again later")
		case http.StatusNotFound:
			return nil, rest_errors.NewBadRequestError("user not registered")
		default:
			return nil, err
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, rest_errors.NewBadRequestError("invalid credentials")
	}
	return user, nil
}

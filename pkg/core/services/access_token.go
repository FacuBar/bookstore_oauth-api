package services

import (
	"strings"
	"sync"
	"time"

	"github.com/FacuBar/bookstore_oauth-api/pkg/core/domain"
	"github.com/FacuBar/bookstore_oauth-api/pkg/core/ports"
	"github.com/FacuBar/bookstore_utils-go/rest_errors"
	uuid "github.com/satori/go.uuid"
)

var (
	onceTokenService     sync.Once
	instanceTokenService *accessTokenService
)

type accessTokenService struct {
	repo     ports.AccessTokenRepository
	uservice ports.UsersService
}

func NewAccessTokenService(repo ports.AccessTokenRepository, servc ports.UsersService) ports.AcessTokenService {
	onceTokenService.Do(func() {
		instanceTokenService = &accessTokenService{
			repo:     repo,
			uservice: servc,
		}
	})

	return instanceTokenService
}

const (
	expirationTime = 48
)

func (s *accessTokenService) Create(email string, password string) (*domain.AccessToken, rest_errors.RestErr) {
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return nil, rest_errors.NewBadRequestError("not valid credentials")
	}

	var user *domain.User
	var err rest_errors.RestErr

	user, err = s.uservice.Login(email, password)
	if err != nil {
		if err.Message() != "user not registered" {
			return nil, err
		}
		// in case replication of user is delayed, a call
		// to the users microservice is done
		user, err = s.repo.LoginUser(email, password)
		if err != nil {
			return nil, err
		}
	}

	accestToken := domain.AccessToken{
		UserId:      user.Id,
		UserRole:    user.Role,
		Expires:     time.Now().UTC().Add(expirationTime * time.Hour).Unix(),
		AccessToken: uuid.NewV4().String(),
		TokenType:   "Bearer",
	}

	if err := s.repo.Create(accestToken); err != nil {
		return nil, err
	}

	return &accestToken, nil
}

func (s *accessTokenService) GetById(id string) (*domain.AccessToken, rest_errors.RestErr) {
	id = strings.TrimSpace(id)
	if len(id) == 0 {
		return nil, rest_errors.NewBadRequestError("invalid access token id")
	}

	accessToken, err := s.repo.GetById(id)
	if err != nil {
		return nil, err
	}

	if time.Now().After(time.Unix(accessToken.Expires, 0)) {
		return nil, rest_errors.NewUnauthorizedError("Token expired")
	}

	return accessToken, nil
}

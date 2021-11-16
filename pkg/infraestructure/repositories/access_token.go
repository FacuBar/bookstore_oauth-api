package repositories

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/FacuBar/bookstore_oauth-api/pkg/core/domain"
	"github.com/FacuBar/bookstore_oauth-api/pkg/core/ports"
	"github.com/FacuBar/bookstore_utils-go/rest_errors"
	"github.com/gocql/gocql"
)

type accessTokenRepository struct {
	db   *gocql.Session
	rest *http.Client
}

var (
	onceTokenRepo     sync.Once
	instanceTokenRepo *accessTokenRepository
)

func NewAccessTokenRepository(db *gocql.Session, rest *http.Client) ports.AccessTokenRepository {
	onceTokenRepo.Do(func() {
		instanceTokenRepo = &accessTokenRepository{
			db:   db,
			rest: rest,
		}
	})

	return instanceTokenRepo
}

const (
	queryGetAccessToken    = "SELECT access_token, user_id, expires FROM access_tokens WHERE access_token=?;"
	queryCreateAccessToken = "INSERT INTO access_tokens(access_token, user_id,  expires) VALUES (?, ?, ?)"
	queryUpdateExpires     = "UPDATE access_tokens SET expires=? WHERE access_token=?;"
)

func (r *accessTokenRepository) Create(at domain.AccessToken) rest_errors.RestErr {
	if err := r.db.Query(queryCreateAccessToken,
		at.AccessToken,
		at.UserId,
		at.Expires,
	).Exec(); err != nil {
		return rest_errors.NewInternalServerError(err.Error())
	}
	return nil
}

func (r *accessTokenRepository) GetById(Id string) (*domain.AccessToken, rest_errors.RestErr) {
	var result domain.AccessToken
	if err := r.db.Query(queryGetAccessToken, Id).Scan(
		&result.AccessToken,
		&result.UserId,
		&result.Expires,
	); err != nil {
		if err == gocql.ErrNotFound {
			return nil, rest_errors.NewNotFoundError("no access token")
		}
		return nil, rest_errors.NewInternalServerError(err.Error())
	}

	return &result, nil
}

func (r *accessTokenRepository) LoginUser(email string, password string) (*domain.User, rest_errors.RestErr) {
	request := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    email,
		Password: password,
	}

	jsonReq, _ := json.Marshal(request)
	b := bytes.NewBuffer(jsonReq)

	response, err := r.rest.Post("http://localhost:8080/users/login", "application/json", b)
	if err != nil {
		return nil, rest_errors.NewInternalServerError(err.Error())
	}
	defer response.Body.Close()

	bodyBytes, _ := io.ReadAll(response.Body)

	if response.StatusCode > 299 {
		restErr, err := rest_errors.NewRestErrorFromBytes(bodyBytes)
		if err != nil {
			return nil, rest_errors.NewInternalServerError("invalid restclient response")
		}
		return nil, restErr
	}

	var user domain.User
	if err := json.Unmarshal(bodyBytes, &user); err != nil {
		return nil, rest_errors.NewInternalServerError("error when trying to unmarshal users response")
	}
	return &user, nil
}

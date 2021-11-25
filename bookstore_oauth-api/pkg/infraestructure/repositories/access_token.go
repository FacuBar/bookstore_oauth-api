package repositories

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/FacuBar/bookstore_oauth-api/pkg/core/domain"
	"github.com/FacuBar/bookstore_oauth-api/pkg/core/ports"
	"github.com/FacuBar/bookstore_utils-go/rest_errors"
)

type accessTokenRepository struct {
	db   *sql.DB
	rest *http.Client
}

var (
	onceTokenRepo     sync.Once
	instanceTokenRepo *accessTokenRepository
)

func NewAccessTokenRepository(db *sql.DB, rest *http.Client) ports.AccessTokenRepository {
	onceTokenRepo.Do(func() {
		instanceTokenRepo = &accessTokenRepository{
			db:   db,
			rest: rest,
		}
	})

	return instanceTokenRepo
}

const (
	queryGetAccessToken    = "SELECT access_token, user_id, user_role, expires FROM access_tokens WHERE access_token=?;"
	queryCreateAccessToken = "INSERT INTO access_tokens(access_token, user_id, user_role, expires) VALUES (?, ?, ?, ?)"
	queryDeleteAccessToken = "DELETE FROM access_tokens WHERE user_id=?"
	// queryUpdateExpires     = "UPDATE access_tokens SET expires=? WHERE access_token=?;"
)

func (r *accessTokenRepository) Create(at domain.AccessToken) rest_errors.RestErr {
	if _, err := r.db.Exec(queryDeleteAccessToken, at.UserId); err != nil {
		return rest_errors.NewInternalServerError(err.Error())
	}

	stmt, err := r.db.Prepare(queryCreateAccessToken)
	if err != nil {
		return rest_errors.NewInternalServerError(err.Error())
	}
	defer stmt.Close()

	if _, err := stmt.Exec(at.AccessToken, at.UserId, at.UserRole, at.Expires); err != nil {
		return rest_errors.NewInternalServerError(err.Error())
	}

	return nil
}

func (r *accessTokenRepository) GetById(Id string) (*domain.AccessToken, rest_errors.RestErr) {
	var at domain.AccessToken
	stmt, err := r.db.Prepare(queryGetAccessToken)
	if err != nil {
		return nil, rest_errors.NewInternalServerError(err.Error())
	}

	result := stmt.QueryRow(Id)
	if err := result.Scan(&at.AccessToken, &at.UserId, &at.UserRole, &at.Expires); err != nil {
		return nil, rest_errors.NewInternalServerError(err.Error())
	}

	return &at, nil
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

	if response.StatusCode > 299 {
		bodyBytes, _ := io.ReadAll(response.Body)
		restErr, err := rest_errors.NewRestErrorFromBytes(bodyBytes)
		if err != nil {
			return nil, rest_errors.NewInternalServerError("invalid restclient response")
		}
		return nil, restErr
	}

	var user domain.User
	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		return nil, rest_errors.NewInternalServerError("error when trying to unmarshal users response")
	}
	return &user, nil
}

package repositories

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/FacuBar/bookstore_oauth-api/pkg/core/domain"
	"github.com/stretchr/testify/assert"
)

func NewMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

var (
	atTest = domain.AccessToken{
		AccessToken: "084a4a0f-92cc-46e6-9b57-1d2aed3c389e",
		UserId:      1,
		UserRole:    "user",
	}
)

func TestCreate(t *testing.T) {
	queryCreate := "INSERT INTO access_tokens\\(access_token, user_id, user_role, expires\\) VALUES \\(\\?, \\?, \\?, \\?\\)"
	queryDelete := "DELETE FROM access_tokens WHERE user_id\\=?"

	t.Run("NoError", func(t *testing.T) {
		db, mock := NewMock()
		mock.ExpectExec(queryDelete).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectPrepare(queryCreate).ExpectExec().WithArgs(atTest.AccessToken, atTest.UserId, atTest.UserRole, atTest.Expires).WillReturnResult(sqlmock.NewResult(1, 1))

		atRepo := accessTokenRepository{db: db, rest: nil}
		err := atRepo.Create(atTest)

		assert.Nil(t, err)
	})

	t.Run("ErrorDeletingTokens", func(t *testing.T) {
		db, mock := NewMock()
		mock.ExpectExec(queryDelete).WithArgs(1).WillReturnError(sql.ErrConnDone)

		atRepo := accessTokenRepository{db: db, rest: nil}
		err := atRepo.Create(atTest)

		assert.NotNil(t, err)
	})

	t.Run("ErrorPreparingInsert", func(t *testing.T) {
		db, mock := NewMock()
		mock.ExpectExec(queryDelete).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectPrepare(queryCreate).WillReturnError(sql.ErrConnDone)

		atRepo := accessTokenRepository{db: db, rest: nil}
		err := atRepo.Create(atTest)

		assert.NotNil(t, err)
	})

	t.Run("ErrorInserting", func(t *testing.T) {
		db, mock := NewMock()
		mock.ExpectExec(queryDelete).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectPrepare(queryCreate).ExpectExec().WillReturnError(sql.ErrConnDone)

		atRepo := accessTokenRepository{db: db, rest: nil}
		err := atRepo.Create(atTest)

		assert.NotNil(t, err)
	})
}

func TestGetById(t *testing.T) {
	query := "SELECT access_token, user_id, user_role, expires FROM access_tokens WHERE access_token=\\?;"

	t.Run("NoError", func(t *testing.T) {
		db, mock := NewMock()
		row := mock.NewRows([]string{"access_token", "user_id", "user_role", "expires"}).
			AddRow("084a4a0f-92cc-46e6-9b57-1d2aed3c389e", 1, "user", 1637510344)
		mock.ExpectPrepare(query).ExpectQuery().WillReturnRows(row)

		atRepo := accessTokenRepository{db: db, rest: nil}
		at, err := atRepo.GetById("084a4a0f-92cc-46e6-9b57-1d2aed3c389e")

		assert.Nil(t, err)
		assert.NotNil(t, at)
		assert.EqualValues(t, "084a4a0f-92cc-46e6-9b57-1d2aed3c389e", at.AccessToken)
		assert.EqualValues(t, "user", at.UserRole)
	})

	t.Run("ErrorPreparingStatement", func(t *testing.T) {
		db, mock := NewMock()
		mock.ExpectPrepare(query).WillReturnError(errors.New(""))

		atRepo := accessTokenRepository{db: db, rest: nil}
		at, err := atRepo.GetById("084a4a0f-92cc-46e6-9b57-1d2aed3c389e")

		assert.Nil(t, at)
		assert.NotNil(t, err)
	})

	t.Run("ErrorScaningResult", func(t *testing.T) {
		db, mock := NewMock()
		mock.ExpectPrepare(query).ExpectQuery().WillReturnError(errors.New(""))

		atRepo := accessTokenRepository{db: db, rest: nil}
		at, err := atRepo.GetById("084a4a0f-92cc-46e6-9b57-1d2aed3c389e")

		assert.Nil(t, at)
		assert.NotNil(t, err)
	})

	t.Run("ErrorNoRows", func(t *testing.T) {
		db, mock := NewMock()
		mock.ExpectPrepare(query).ExpectQuery().WillReturnError(sql.ErrNoRows)

		atRepo := accessTokenRepository{db: db, rest: nil}
		at, err := atRepo.GetById("084a4a0f-92cc-46e6-9b57-1d2aed3c389e")

		assert.Nil(t, at)
		assert.NotNil(t, err)
		assert.EqualValues(t, "access_token not found", err.Message())
	})
}

func TestLoginUser(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(`{"id":1,"first_name":"Oscar","last_name":"Isaac","email": "oscaac@gmail.com","date_created": "2021-11-19 03:07:42","role":"user"}`))
		}))
		testServer.Listener.Close()
		l, _ := net.Listen("tcp", "127.0.0.1:8080")
		testServer.Listener = l
		testServer.Start()
		defer testServer.Close()

		httpClient := testServer.Client()
		atRepo := accessTokenRepository{db: nil, rest: httpClient}

		user, err := atRepo.LoginUser("oscaac@gmail.com", "password")

		assert.Nil(t, err)
		assert.NotNil(t, user)
		assert.EqualValues(t, 1, user.Id)
		assert.EqualValues(t, "oscaac@gmail.com", user.Email)
	})

	t.Run("ErrorSendingRequest", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))
		defer testServer.Close()

		httpClient := testServer.Client()
		atRepo := accessTokenRepository{db: nil, rest: httpClient}

		user, err := atRepo.LoginUser("oscaac@gmail.com", "password")

		assert.Nil(t, user)
		assert.NotNil(t, err)
		assert.EqualValues(t, `Post "http://localhost:8080/users/login": dial tcp [::1]:8080: connect: connection refused`, err.Message())
	})

	t.Run("ErrorReceivedFromRequest", func(t *testing.T) {
		testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{"message": "Invalid credentials","status": 400,"error": "bad_request"}`))
		}))
		testServer.Listener.Close()
		l, _ := net.Listen("tcp", "127.0.0.1:8080")
		testServer.Listener = l
		testServer.Start()
		defer testServer.Close()

		httpClient := testServer.Client()
		atRepo := accessTokenRepository{db: nil, rest: httpClient}

		user, err := atRepo.LoginUser("oscaac@gmail.com", "password")

		assert.Nil(t, user)
		assert.NotNil(t, err)
		assert.EqualValues(t, http.StatusBadRequest, err.Status())
		assert.EqualValues(t, "Invalid credentials", err.Message())
	})

	t.Run("ErrorUnmarshalingErrorResponse", func(t *testing.T) {
		testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{"message": "Invalid credentials","status": "400","error": "bad_request"}`))
		}))
		testServer.Listener.Close()
		l, _ := net.Listen("tcp", "127.0.0.1:8080")
		testServer.Listener = l
		testServer.Start()
		defer testServer.Close()

		httpClient := testServer.Client()
		atRepo := accessTokenRepository{db: nil, rest: httpClient}

		user, err := atRepo.LoginUser("oscaac@gmail.com", "password")

		assert.Nil(t, user)
		assert.NotNil(t, err)
		assert.EqualValues(t, "invalid restclient response", err.Message())
	})

	t.Run("ErrorUnmarshalingUserResponse", func(t *testing.T) {
		testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(`{"id":"1","first_name":"Oscar","last_name":"Isaac"}`))
		}))
		testServer.Listener.Close()
		l, _ := net.Listen("tcp", "127.0.0.1:8080")
		testServer.Listener = l
		testServer.Start()
		defer testServer.Close()

		httpClient := testServer.Client()
		atRepo := accessTokenRepository{db: nil, rest: httpClient}

		user, err := atRepo.LoginUser("oscaac@gmail.com", "password")

		assert.Nil(t, user)
		assert.NotNil(t, err)
		assert.EqualValues(t, "error when trying to unmarshal users response", err.Message())
	})
}

// func TestDeleteById(t testing.T) {
// 	query := regexp.QuoteMeta(queryDeleteAccessToken)
// 	db, mock := NewMock()
// 		mock.ExpectPrepare(query).WillReturnError(errors.New(""))

// 		atRepo := accessTokenRepository{db: db, rest: nil}
// 		at, err := atRepo.GetById("084a4a0f-92cc-46e6-9b57-1d2aed3c389e")

// 		assert.Nil(t, at)
// 		assert.NotNil(t, err)
// }

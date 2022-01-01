package repositories

// TODO: pulish implementation, add proper error handling and finish testing

import (
	"database/sql"
	"net/http"
	"strings"
	"sync"

	"github.com/FacuBar/bookstore_oauth-api/pkg/core/domain"
	"github.com/FacuBar/bookstore_oauth-api/pkg/core/ports"
	"github.com/FacuBar/bookstore_utils-go/rest_errors"
)

var (
	onceUsersRepo     sync.Once
	instanceUsersRepo *usersRepository
)

type usersRepository struct {
	db *sql.DB
}

func NewUsersRepository(db *sql.DB) ports.UsersRepository {
	onceUsersRepo.Do(func() {
		instanceUsersRepo = &usersRepository{
			db: db,
		}
	})
	return instanceUsersRepo
}

const (
	queryGetUserByEmail = "SELECT id, email, role, password FROM users WHERE email=?;"
	queryInsertUser     = "INSERT INTO users(id, email, password, role) VALUES(?, ?, ?, ?);"
)

const (
	errNoRow = "no rows in result"
)

func (r *usersRepository) GetByEmail(email string) (*domain.User, rest_errors.RestErr) {
	stmt, err := r.db.Prepare(queryGetUserByEmail)
	if err != nil {
		return nil, rest_errors.NewInternalServerError("db error")
	}
	defer stmt.Close()

	var user domain.User
	result := stmt.QueryRow(email)
	if err := result.Scan(&user.Id, &user.Email, &user.Role, &user.Password); err != nil {
		if strings.Contains(err.Error(), errNoRow) {
			return nil, rest_errors.NewNotFoundError("user not found")
		}
		return nil, rest_errors.NewInternalServerError("db error")
	}
	return &user, nil
}

func (r *usersRepository) Save(user *domain.User) rest_errors.RestErr {
	stmt, err := r.db.Prepare(queryInsertUser)
	if err != nil {
		return rest_errors.NewInternalServerError("db error")
	}
	defer stmt.Close()

	insertResult, err := stmt.Exec(user.Id, user.Email, user.Password, user.Role)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return rest_errors.NewRestError("email already in user", http.StatusConflict, "conflict")
		}
		return rest_errors.NewInternalServerError("db error")
	}

	userId, err := insertResult.LastInsertId()
	if err != nil {
		return rest_errors.NewInternalServerError("db error")
	}
	user.Id = userId
	return nil
}

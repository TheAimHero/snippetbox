package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	Get(id int) (*User, error)
	PasswordUpdate(id int, currentPassword, newPassword string) error
}

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (u *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (name, email, hashed_password, created) VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`
	_, err = u.DB.Exec(stmt, name, email, hashedPassword)
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (u *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte
	stmt := `SELECT id, hashed_password FROM users WHERE email = $1`
	err := u.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		return 0, ErrInvalidCredentials
	}
	return id, nil
}

func (u *UserModel) Exists(id int) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT true FROM users WHERE id = $1)`
	err := u.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}

func (u *UserModel) Get(id int) (*User, error) {
	var user User
	stmt := `SELECT id, name, email, created FROM users WHERE id = $1`
	err := u.DB.QueryRow(stmt, id).Scan(&user.ID, &user.Name, &user.Email, &user.Created)
	if err != nil {
		return nil, ErrNoRecord
	}
	return &user, nil
}

func (u *UserModel) PasswordUpdate(id int, currentPassword, newPassword string) error {
	stmt := `SELECT hashed_password FROM users WHERE id = $1`
	var hashedPassword []byte
	err := u.DB.QueryRow(stmt, id).Scan(&hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return err
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(currentPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}
		return err
	}
	hashedPassword, err = bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	stmt = `UPDATE users SET hashed_password = $1 WHERE id = $2`
	_, err = u.DB.Exec(stmt, hashedPassword, id)
	return err
}

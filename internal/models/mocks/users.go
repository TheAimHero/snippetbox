package mocks

import (
	"time"

	"github.com/TheAimHero/sb/internal/models"
)

type UserModel struct{}

func (m *UserModel) Insert(name, email, password string) error {
	switch email {
	case "dupe@example.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	if email == "alice@example.com" && password == "pa$$word" {
		return 1, nil
	}
	return 0, models.ErrInvalidCredentials
}

func (m *UserModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (m *UserModel) Get(id int) (*models.User, error) {
	if id == 1 {
		return &models.User{
			ID:      1,
			Name:    "Alice",
			Email:   "alice@example.com",
			Created: time.Now(),
		}, nil
	} else {
		return nil, models.ErrNoRecord
	}
}

func (m *UserModel) PasswordUpdate(id int, currentPassword, newPassword string) error {
	return nil
}

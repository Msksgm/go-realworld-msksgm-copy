package conduit

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uint      `json:"-"`
	Email        string    `json:"email,omitempty"`
	Username     string    `json:"username,omitempty"`
	Bio          string    `json:"bio,omitempty"`
	Image        string    `json:"image,omitempty"`
	Token        string    `json:"token,omitempty"`
	Following    []*User   `json:"-"`
	Followers    []*User   `json:"-"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"-" db:"created_at"`
	UpdatedAt    time.Time `json:"-" db:"updated_at"`
}

func (u *User) SetPassword(password string) error {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// return better error message
		return err
	}

	u.PasswordHash = string(hashBytes)

	return nil
}

type UserService interface {
	CreateUser(context.Context, *User) error
}

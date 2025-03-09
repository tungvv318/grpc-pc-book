package service

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username      string
	HasedPassword string
	Role          string
}

func NewUser(username string, password string, role string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}
	return &User{
		Username:      username,
		HasedPassword: string(hashedPassword),
		Role:          role,
	}, nil
}

func (user *User) IsCorrectPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HasedPassword), []byte(password))
	return err == nil
}

func (user *User) Clone() *User {
	return &User{
		Username:      user.Username,
		HasedPassword: user.HasedPassword,
		Role:          user.Role,
	}
}

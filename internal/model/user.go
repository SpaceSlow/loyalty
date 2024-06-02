package model

import (
	"crypto/sha512"
)

type User struct {
	Username          string `json:"login"`
	Password          string `json:"password"`
	passwordHash      string
	passwordSalt      string
	passwordIteration int
}

func (u *User) GetPasswordHash() string {
	h := sha512.New()
	h.Write([]byte(u.Password))
	resultHash := make([]byte, 512)

	for range u.passwordIteration {
		resultHash = h.Sum([]byte(u.passwordSalt))
	}

	return string(resultHash)
}

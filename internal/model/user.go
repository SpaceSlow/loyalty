package model

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	PBKDF2SHA512Alg     = "pbkdf2-sha512"
	PBKDF2KeyIterations = 500000
)

type User struct {
	Username     string `json:"login"`
	Password     string `json:"password"`
	PasswordHash string
}

func (u *User) GenerateHash() error {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return err
	}
	u.PasswordHash = fmt.Sprintf(
		"%s$%s$%v$%s",
		PBKDF2SHA512Alg,
		base64.StdEncoding.EncodeToString(salt),
		PBKDF2KeyIterations,
		base64.StdEncoding.EncodeToString(pbkdf2.Key([]byte(u.Password), salt, PBKDF2KeyIterations, 32, sha512.New)),
	)

	return nil
}

func (u *User) CheckPassword(passwordHash string) (bool, error) {
	fields := strings.Split(passwordHash, "$")

	if len(fields) != 4 {
		return false, errors.New("invalid password hash layout")
	}
	alg := fields[0]
	switch alg {
	case PBKDF2SHA512Alg:
	default:
		return false, errors.New("unknown password hash algorithm")
	}
	salt, err := base64.StdEncoding.DecodeString(fields[1])
	if err != nil {
		return false, err
	}
	iterationNumber, err := strconv.Atoi(fields[2])
	if err != nil {
		return false, err
	}
	storedHash := fields[3]
	calculateHash := base64.StdEncoding.EncodeToString(pbkdf2.Key([]byte(u.Password), salt, iterationNumber, 32, sha512.New))

	return calculateHash == storedHash, nil
}

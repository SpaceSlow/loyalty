package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func FuzzUser_GenerateHash(f *testing.F) {
	f.Fuzz(func(t *testing.T, username, password string) {
		u := &User{
			Username:     username,
			Password:     password,
			PasswordHash: "",
		}
		assert.NoError(t, u.GenerateHash(), "")
		assert.NotEmpty(t, u.PasswordHash, "GenerateHash() doesn't set User.PasswordHash field")
		assert.Len(t, u.PasswordHash, 110, "")
	})
}

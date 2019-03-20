package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_hash(t *testing.T) {
	password := "a"
	sb64, dkb64 := SaltAndHashPassword("a")
	result := CheckPasswordAgainstSaltAndHash(password, sb64, dkb64)
	assert.True(t, result)
}

func Test_isEmailInValidFormat(t *testing.T) {
	assert.True(t, isEmailInValidFormat("a@b.com"))
	assert.False(t, isEmailInValidFormat("a@bcom"))
	assert.False(t, isEmailInValidFormat("a@b@c"))
	assert.False(t, isEmailInValidFormat("a.b.com"))
	assert.False(t, isEmailInValidFormat("@.com"))
}

func Test_validatePassword(t *testing.T) {
	assert.Nil(t, validatePassword("v&._9Vpd"))
	assert.NotNil(t, validatePassword("aaaaaaaabbbbbbbb"))
	assert.NotNil(t, validatePassword("#9fA"))
}

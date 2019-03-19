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

package app

import (
	"sort"
	"testing"

	"github.com/budden/semdict/pkg/shared"
	"github.com/budden/semdict/pkg/user"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateSecretConfigDataExample(t *testing.T) {
	// thiss not merely a test, but also an utility: it creates
	// secret-data.config.json.example
	// FIXME use make or at least bash script to generate an example2
	fn := "../../" + ConfigFileName + ".example"
	scd, err := SaveSecretConfigDataExample(fn)
	assert.Nilf(t, err, "Error %#v in SaveSecretConfigDataExample", err)
	err2 := LoadSecretConfigData(fn)
	assert.Nilf(t, err2, "Error %#v in LoadSecretConfigData", err)
	assert.Equal(t, *scd, shared.SecretConfigData)
}

func Test_Nonce(t *testing.T) {
	const countOfNonces = 500
	const allowedNumberOfNonUniqueNonces = 3
	const lengthForNonceTest = 16
	var unsorted [countOfNonces]string

	// generate them
	for i := 0; i < countOfNonces; i++ {
		unsorted[i] = user.GenNonce(lengthForNonceTest)
	}

	sorted := unsorted[:]
	// sort them (and destroy unsorted)
	sort.Slice(sorted, func(n1, n2 int) bool {
		return sorted[n1] < sorted[n2]
	})

	countOfNonUniqueNonces := 0
	for i := 0; i < countOfNonces-1; i++ {
		if sorted[i] == sorted[i+1] {
			countOfNonUniqueNonces++
		}
	}

	assert.True(t, countOfNonUniqueNonces <= allowedNumberOfNonUniqueNonces)
}

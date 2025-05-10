package tests

import (
	"auth-service/jwtutils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreation(t *testing.T) {
	assert := assert.New(t)

	tokenString, err := jwtutils.CreateJWT()
	assert.Nil(err)
	assert.NotEqual(len(tokenString), 0)
}

func TestVerify(t *testing.T) {
	assert := assert.New(t)

	tokenString, err := jwtutils.CreateJWT()
	assert.Nil(err)
	assert.NotEqual(len(tokenString), 0)

	claims, err := jwtutils.VerifyJWT(tokenString)
	assert.Nil(err)
	assert.Equal(claims["username"], "guest")
}

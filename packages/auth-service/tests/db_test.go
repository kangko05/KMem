package tests

import (
	"auth-service/dbutils"
	"fmt"
	"testing"

	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestInsertUser(t *testing.T) {
	guest := "guest"

	assert := assert.New(t)

	userDB, err := dbutils.Connect()
	assert.Nil(err)
	defer userDB.Close()

	assert.Nil(userDB.Ping())
	err = userDB.InsertUser(guest, "guest_password")
	if err == sqlite3.ErrConstraintUnique {
		fmt.Println("its ok")
	}

	pass, exists := userDB.FindUser(guest)
	assert.True(exists)

	fmt.Println(pass)
}

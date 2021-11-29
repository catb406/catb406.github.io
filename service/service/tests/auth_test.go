package tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestLogin(t *testing.T) {
	t.Run("Test /auth/login", func(t *testing.T) {
		api, teardown := configureEnvironment(t)
		defer teardown()

		authParams := JSON{
			"username": "test",
			"password": "123",
		}
		response := post(api, "/auth/login", authParams)
		fmt.Println(response.Body.String())
		assert.Equal(t, http.StatusOK, response.Code)
	})
}

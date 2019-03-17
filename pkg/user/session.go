package user

import (
	"fmt"

	"github.com/budden/semdict/pkg/apperror"

	"github.com/gin-gonic/gin"
)

// EnsureLoggedIn causes a request to be aborted with an error
// if the user is not logged in
func EnsureLoggedIn(c *gin.Context) {
	loggedInInterface, _ := c.Get("is_logged_in")
	loggedIn := loggedInInterface.(bool)
	if !loggedIn {
		apperror.Panic500If(apperror.ErrDummy, "Please log in to view this page")
	}
}

// EnsureNotLoggedIn ensures that a request will be aborted with an error
// if the user is already logged in
func EnsureNotLoggedIn(c *gin.Context) {
	loggedInInterface, _ := c.Get("is_logged_in")
	loggedIn := loggedInInterface.(bool)
	if loggedIn {
		// FIXME - invent "Panic401" or do whatever else meaningful here,
		// but don't return 500.
		apperror.Panic500If(apperror.ErrDummy, "Please log in to view this page")
	}
}

// SetUserStatus sets a flag indicating whether the request was from an authenticated user or not
func SetUserStatus() gin.HandlerFunc {
	return setUserStatusFn
}

func setUserStatusFn(c *gin.Context) {
	fmt.Println("setUserStatusFn!")
	if token, err := c.Cookie("token"); err == nil || token != "" {
		c.Set("is_logged_in", true)
	} else {
		c.Set("is_logged_in", false)
	}
}

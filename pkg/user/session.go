package user

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/shared"

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
	if token, err := c.Cookie("token"); err == nil || token != "" {
		c.Set("is_logged_in", true)
	} else {
		c.Set("is_logged_in", false)
	}
}

// PerformLogin handles login route
func PerformLogin(c *gin.Context) {
	// We could check that user is not yet logged in, but we won't do
	// Obtain the POSTed username and password values
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Check if the username/password combination is valid
	if isUserValid(username, password) {
		// If the username/password is valid set the token in a cookie
		token := generateSessionToken()
		c.SetCookie("token", token, 3600, "", "", false, true)

		c.HTML(http.StatusOK, "general.html",
			shared.GeneralTemplateParams{Message: "Welcome, a citizen!"})

	} else {
		c.HTML(http.StatusBadRequest, "general.html",
			shared.GeneralTemplateParams{Message: "Go away, stranger!"})
	}
}

func isUserValid(username, password string) bool {
	// TODO do actual things
	return true
}

func generateSessionToken() string {
	// TODO do actual check
	return "blablabla"
}

// Logout performs a logout
func Logout(c *gin.Context) {
	// Clear the cookie
	c.SetCookie("token", "", -1, "", "", false, true)

	// Redirect to the home page
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// LoginFormPageHandler renders a /loginform page
func LoginFormPageHandler(c *gin.Context) {
	EnsureNotLoggedIn(c)
	c.HTML(http.StatusOK,
		"loginform.html",
		shared.GeneralTemplateParams{Message: "Login"})
}

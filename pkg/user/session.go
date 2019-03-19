package user

import (
	"fmt"
	"net/http"

	"github.com/budden/semdict/pkg/database"

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

// LoginFormSubmitPostHandler handles login route
func LoginFormSubmitPostHandler(c *gin.Context) {
	// We could check that user is not yet logged in, but we won't do
	// Obtain the POSTed username and password values
	nickname := c.PostForm("nickname")
	password := c.PostForm("password")

	// Check if the username/password combination is valid
	if isUserValid(nickname, password) {
		// If the username/password is valid set the token in a cookie
		token := generateSessionToken()
		c.SetCookie("token", token, 3600, "", "", false, true)

		c.HTML(http.StatusOK, "general.html",
			shared.GeneralTemplateParams{Message: fmt.Sprintf("Welcome, %s!", nickname)})

	} else {
		c.HTML(http.StatusBadRequest, "general.html",
			shared.GeneralTemplateParams{Message: "Go away, stranger!"})
	}
}

func isUserValid(nickname, password string) bool {
	// TODO do actual things
	if !isNicknameInValidFormat(nickname) {
		apperror.Panic500If(apperror.ErrDummy, "Nickname has an illegal format (e.g. invalid characters)")
	}

	if !isPasswordInValidFormat(nickname) {
		apperror.Panic500If(apperror.ErrDummy, "Password has an illegal format (e.g. invalid characters)")
	}

	var sud SDUserData
	getSDUserDataFromDb(nickname, &sud)

	return CheckPasswordAgainstSaltAndHash(password, sud.Salt, sud.Hash)
}

// function could be general, but it's error messages are login process specific. FIXME
func getSDUserDataFromDb(nickname string, sud *SDUserData) {
	db := database.SDUsersDb
	// have <= 1 record only due to unique index
	res, err := db.Db.Queryx("select * from sduser where nickname = $1 limit 1", nickname)
	apperror.Panic500If(err, "Unable to login, sorry")
	dataFound := false
	for res.Next() {
		err1 := res.StructScan(sud)
		apperror.GracefullyExitAppIf(err1, "Failed to read sduser's record: «%s»", err1)
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500IfLogError(err, "Attempt to log on as a non-existing user «%s»", nickname)
	}
	return
}

func generateSessionToken() string {
	return GenNonce(32)
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

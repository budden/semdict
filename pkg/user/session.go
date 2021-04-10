package user

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/budden/semdict/pkg/privilegecode"
	"github.com/budden/semdict/pkg/sddb"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/shared"

	"github.com/gin-gonic/gin"
)

// EnsureLoggedIn causes a request to be aborted with an error
// if the user is not logged in. Can only be used downstream from SetUserStatusMiddleware
// middleware
func EnsureLoggedIn(c *gin.Context) {
	if !IsLoggedIn(c) {
		apperror.Panic500If(apperror.ErrDummy, "Please log in to view this page")
	}
}

// EnsureNotLoggedIn ensures that a request will be aborted with an error
// if the user is already logged in.  Can only be used downstream from SetUserStatusMiddleware
// middleware
func EnsureNotLoggedIn(c *gin.Context) {
	if IsLoggedIn(c) {
		// FIXME - invent "Panic401" or do whatever else meaningful here,
		// but don't return 500.
		apperror.Panic500If(apperror.ErrDummy, "Please log OUT to view this page")
	}
}

// IsLoggedIn is true if the user is logged in with valid credentials.
// Can only be used downstream from SetUserStatusMiddleware middleware
func IsLoggedIn(c *gin.Context) bool {
	loggedInInterface, exists := c.Get("is_logged_in")
	loggedIn := loggedInInterface.(bool)
	if !exists {
		apperror.GracefullyExitAppIf(apperror.ErrDummy, "Only call this one after SetUserStatusMiddleware")
	}
	return loggedIn
}

// GetSDUserIdOrZero returns sduserid for a logged in user, or 0 for a not logged
// Use downstream from SetUserStatusMiddleware middleware
func GetSDUserIdOrZero(c *gin.Context) int32 {
	if !IsLoggedIn(c) {
		return 0
	}
	sduserid, mustExist := c.Get("sduserid")
	if !mustExist {
		apperror.GracefullyExitAppIf(apperror.ErrDummy, "sduser must be set if a user is logged in")
	}
	if sduserIdint32, ok := sduserid.(int32); ok {
		return sduserIdint32
	}
	apperror.GracefullyExitAppIf(apperror.ErrDummy, "sduser must be int32 in a gin context")
	return -1 // this must never happen
}

// SetUserStatusMiddleware sets a flag indicating whether the request was from an authenticated user or not
func SetUserStatusMiddleware() gin.HandlerFunc {
	return setUserStatusFn
}

func setUserStatusFn(c *gin.Context) {
	tokenPresent, tokenValid, sduserid := getAndValidateToken(c)
	isLoggedIn := false
	if !tokenPresent {
		// ok, it will be false
	} else if !tokenValid {
		// session expired, or, worse, it is an attack
		apperror.LogAttack(c, errors.New("setUserStatusFn: invalid token"))
		endSessionIfThereIsOne(c)
	} else {
		// hence token is present and valid
		isLoggedIn = true
	}
	c.Set("is_logged_in", isLoggedIn)
	if isLoggedIn {
		c.Set("sduserid", int32(sduserid))
	} else {
		c.Set("sduserid", nil)
	}
}

func getAndValidateToken(c *gin.Context) (tokenPresent, tokenValid bool, sduserid int) {
	if shared.SecretConfigData.UserAlwaysLoggedIn != 0 {
		// special case for debugging
		tokenPresent = true
		tokenValid = true
		sduserid = shared.SecretConfigData.UserAlwaysLoggedIn
	} else {
		var token string
		token, tokenPresent = getSessionToken(c)
		if !tokenPresent {
			return
		}
		params := map[string]interface{}{"token": token}
		res, err := sddb.NamedReadQuery(`select sduserid from session
				where eid=:token and expireat > current_timestamp limit 1`,
			params)
		apperror.Panic500AndErrorIf(err, "Failed to check validity of your session, sorry. Please logout and retry")
		defer sddb.CloseRows(res)()
		for res.Next() {
			err1 := res.Scan(&sduserid)
			apperror.GracefullyExitAppIf(err1, "Failed to check if session is present, error is «%#v»", err1)
		}
		tokenValid = (sduserid != 0)
	}
	return
}

// LoginSubmitPostHandler handles login route
func LoginSubmitPostHandler(c *gin.Context) {
	// We could check that user is not yet logged in, but we won't do
	// Obtain the POSTed username and password values
	nickname := c.PostForm("nickname")
	password := c.PostForm("password")

	// just in case, if there is an old session, we close it.
	endSessionIfThereIsOne(c)

	// Check if the username/password combination is valid
	sduserid := isUserValid(nickname, password)
	if sduserid != 0 {
		// check if the use have the login privilege
		if !isUserHavePrivilege(sduserid, privilegecode.Login) {
			c.HTML(http.StatusBadRequest, "general.t.html",
				shared.GeneralTemplateParams{Message: fmt.Sprintf("Sorry, %s, but you have no login privilege", nickname)})
			return
		}

		// If the username/password is valid set the token in a cookie
		token := generateSessionToken()

		recordSessionTokenIntoDb(nickname, token)

		c.SetCookie("token", token, 3600, "", "", false, true)

		c.HTML(http.StatusOK, "general.t.html",
			shared.GeneralTemplateParams{Message: fmt.Sprintf("Welcome, %s!", nickname)})

	} else {
		c.HTML(http.StatusBadRequest, "general.t.html",
			shared.GeneralTemplateParams{Message: "Go away, stranger!"})
	}
}

// isUserValid checks that the user is valid, in this case non-zero sduserid is returned
func isUserValid(nickname, password string) (sduserid int32) {
	if !isNicknameInValidFormat(nickname) {
		apperror.Panic500If(apperror.ErrDummy, "Nickname has an illegal format (e.g. invalid characters)")
	}

	if validatePassword(password) != nil {
		apperror.Panic500If(apperror.ErrDummy, "Password has an illegal format (e.g. invalid characters)")
	}

	var sud SDUserData
	getSDUserDataFromDb(nickname, &sud)

	if CheckPasswordAgainstSaltAndHash(password, sud.Salt, sud.Hash) {
		sduserid = sud.ID
	} else {
		sduserid = 0
	}
	return
}

// function could be general, but it's error messages are login process specific. FIXME
func getSDUserDataFromDb(nickname string, sud *SDUserData) {
	// have <= 1 record only due to unique index
	params := map[string]interface{}{"nickname": nickname}
	res, err := sddb.NamedReadQuery("select * from sduser where nickname = :nickname limit 1", params)
	apperror.Panic500If(err, "Unable to login, sorry")
	defer sddb.CloseRows(res)()
	dataFound := false
	for res.Next() {
		err1 := res.StructScan(sud)
		apperror.GracefullyExitAppIf(err1, "Failed to read sduser's record: «%s»", err1)
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(err, "Attempt to log on as a non-existing user «%s»", nickname)
	}
	return
}

func isUserHavePrivilege(sduserid int32, privilegekind privilegecode.Enum) (granted bool) {
	params := map[string]interface{}{"sduserid": sduserid, "privilegekind": privilegekind}
	res, err := sddb.NamedReadQuery("select isuserhaveprivilege(:sduserid, :privilegekind)", params)
	apperror.GracefullyExitAppIf(err, "Error obtaining user privilege: «%s»", err)
	defer sddb.CloseRows(res)()
	dataFound := false
	for res.Next() {
		err1 := res.Scan(&granted)
		apperror.GracefullyExitAppIf(err1, "Failed to read sduser's record: «%s»", err1)
		dataFound = true
	}
	if !dataFound {
		apperror.GracefullyExitAppIf(apperror.ErrDummy, "got no data from isuserhaveprivilege")
	}
	return granted
}

func generateSessionToken() string {
	return GenNonce(32)
}

func recordSessionTokenIntoDb(nickname, token string) {
	err := sddb.WithTransaction(func(trans *sddb.TransactionType) (err1 error) {
		res, err1 := trans.Tx.Queryx("select begin_session($1,$2)", nickname, token)
		// FIXME process exception with too_many_sessions mentioned
		apperror.GracefullyExitAppIf(err1, "Failed to begin session, error is «%#v»", err1)
		for res.Next() {
			// it returns an id, but we don't need it
		}
		return
	})
	apperror.GracefullyExitAppIf(err, "Failed to begin session 2, error is «%#v»", err)
}

func getSessionToken(c *gin.Context) (token string, found bool) {
	token, err := c.Cookie("token")
	if err == http.ErrNoCookie {
		return
	}
	apperror.GracefullyExitAppIf(err, "Unknown error getting session token: «%s»", err)
	found = true
	return
}

// Logout performs a logout
func Logout(c *gin.Context) {
	endSessionIfThereIsOne(c)
	// Redirect to the home page
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// clear the cookie and delete the session from db
func endSessionIfThereIsOne(c *gin.Context) {
	token, tokenFound := getSessionToken(c)
	if !tokenFound {
		return
	}

	c.SetCookie("token", "", -1, "", "", false, true)

	err := sddb.WithTransaction(func(trans *sddb.TransactionType) (err1 error) {
		res, err1 := trans.Tx.Queryx("select end_session($1)", token)
		apperror.GracefullyExitAppIf(err1, "Failed to end session, error is «%#v»", err1)
		for res.Next() {
			// don't need the result
		}
		return
	})
	apperror.GracefullyExitAppIf(err, "Failed to end session 2, error is «%#v»", err)
}

// LoginFormPageHandler renders a /loginform page
func LoginFormPageHandler(c *gin.Context) {
	EnsureNotLoggedIn(c)
	c.HTML(http.StatusOK,
		"loginform.t.html",
		shared.LoginFormParams{ /*CaptchaID: "100500"*/ })
}

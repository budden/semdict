package user

import (
	"fmt"
	"net/http"

	"github.com/budden/a/pkg/shared"
	"github.com/gin-gonic/gin"
)

// PlayWithEmail sends an email

// RegistrationFormPageHandler renders a /registrationform page
func RegistrationFormPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK,
		"registrationform.html",
		shared.GeneralTemplateParams{Message: "Search Form"})
}

// RegistrationFormSubmitPostHandler processes a registrationformsubmit form post request
func RegistrationFormSubmitPostHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	registrationemail := c.PostForm("registrationemail")
	fmt.Printf("RegistrationFormSubmit: %s, %s, %s\n", username, password, registrationemail)

	c.HTML(http.StatusOK,
		"general.html",
		shared.GeneralTemplateParams{Message: "Check your E-Mail for a confirmation code"})
}
